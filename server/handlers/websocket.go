package handlers

import (
	"net/http"
	"time"

	"game-sandbox/server/game"
	"game-sandbox/server/models"
	"game-sandbox/server/utils"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// 全てのオリジンを許可（開発用）
			return true
		},
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true, // gzip圧縮を有効化
	}
)

// WebSocketHandler はWebSocket接続を処理する
func WebSocketHandler(hub *game.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			utils.Error("Failed to upgrade WebSocket connection", map[string]interface{}{
				"error":       err.Error(),
				"remote_addr": r.RemoteAddr,
			})
			return
		}
		defer conn.Close()

		var player *models.Player
		var gameInstance *game.Game
		var playerID string

		for {
			var msg map[string]interface{}
			err := conn.ReadJSON(&msg)
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					// Normal disconnect
					break
				}
				utils.LogWebSocketError(playerID, "read_message", err)
				break
			}

			msgType, ok := msg["type"].(string)
			if !ok {
				continue
			}

			switch msgType {
			case "join":
				roomID, _ := msg["roomId"].(string)
				playerName, _ := msg["playerName"].(string)
				if playerName == "" {
					playerName = utils.GenerateRandomNickname()
				}

				playerID = utils.GenerateID()
				gameInstance = hub.GetOrCreateGame(roomID)

				gameInstance.AddPlayer(playerID, playerName, conn)
				player, _ = gameInstance.GetPlayer(playerID)

				utils.LogConnectionEvent("connect", playerID, playerName, false)

				gameInstance.ShouldStart()

				// WebSocket書き込みを同期化
				func() {
					player.ConnMu.Lock()
					defer player.ConnMu.Unlock()

					// 参加確認を送信
					response := map[string]interface{}{
						"type":     "gameJoined",
						"playerId": playerID,
					}
					conn.WriteJSON(response)

					// ゲーム設定を送信（グリッド線含む）
					gridLines := gameInstance.GetSpatialGridLines()

					config := models.GameConfig{
						FieldWidth:      utils.FIELD_WIDTH,
						FieldHeight:     utils.FIELD_HEIGHT,
						SphereRadius:    utils.SPHERE_RADIUS,
						CullingWidth:    utils.CULLING_WIDTH,
						CullingHeight:   utils.CULLING_HEIGHT,
						CameraZoomScale: utils.CAMERA_ZOOM_SCALE,
						GridLines:       gridLines,
					}
					configMsg := map[string]interface{}{
						"type":   "gameConfig",
						"config": config,
					}
					conn.WriteJSON(configMsg)

					// 現在のゲーム状態を送信
					state := gameInstance.GetState()

					stateMsg := map[string]interface{}{
						"type":  "gameState",
						"state": state,
					}
					conn.WriteJSON(stateMsg)
				}()

			case "setAcceleration":
				if player == nil || gameInstance == nil {
					continue
				}

				x, xOk := msg["x"].(float64)
				y, yOk := msg["y"].(float64)
				if xOk && yOk && player.Celestial != nil && player.Celestial.Alive {
					// 加速度コマンドを送信
					gameInstance.SendCommand(game.AccelerationCommand{
						Player: player,
						X:      x,
						Y:      y,
					})
				}

			case "ejectSatellite":
				if player == nil || gameInstance == nil {
					continue
				}

				targetX, xOk := msg["targetX"].(float64)
				targetY, yOk := msg["targetY"].(float64)
				if xOk && yOk {
					// 射撃コマンドを送信
					gameInstance.SendCommand(game.ShootCommand{
						Player:  player,
						TargetX: targetX,
						TargetY: targetY,
					})
				}
			}
		}

		// 切断時のクリーンアップ
		if gameInstance != nil && playerID != "" {
			if player != nil {
				utils.LogConnectionEvent("disconnect", playerID, player.Name, player.IsNPC)
			}
			gameInstance.RemovePlayer(playerID)

			// 人間プレイヤーがいなくなったらゲームを停止
			if gameInstance.GetHumanPlayerCount() == 0 {
				endTime := time.Now()
				duration := endTime.Sub(gameInstance.GetStartTime())
				utils.LogGameSessionEvent("game_end", gameInstance.ID, 0, len(gameInstance.GetPlayers()), duration)
				gameInstance.Stop()
				hub.RemoveGame(gameInstance.ID)
			}
		}
	}
}
