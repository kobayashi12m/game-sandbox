package handlers

import (
	"log"
	"net/http"

	"chess-mmo/server/game"
	"chess-mmo/server/models"
	"chess-mmo/server/utils"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// 全てのオリジンを許可（開発用）
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// WebSocketHandler はWebSocket接続を処理する
func WebSocketHandler(hub *game.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
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
				log.Printf("Error reading message: %v", err)
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
					playerName = "Player"
				}

				playerID = utils.GenerateID()
				gameInstance = hub.GetOrCreateGame(roomID)

				gameInstance.AddPlayer(playerID, playerName, conn)
				player, _ = gameInstance.GetPlayer(playerID)

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
						FieldWidth:    utils.FIELD_WIDTH,
						FieldHeight:   utils.FIELD_HEIGHT,
						SphereRadius:  utils.SPHERE_RADIUS,
						FoodRadius:    utils.FOOD_RADIUS,
						CullingWidth:  utils.CULLING_WIDTH,
						CullingHeight: utils.CULLING_HEIGHT,
						GridLines:     gridLines,
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
				if xOk && yOk {
					gameInstance.SetPlayerAcceleration(playerID, x, y)
				}
			}
		}

		// 切断時のクリーンアップ
		if gameInstance != nil && playerID != "" {
			gameInstance.RemovePlayer(playerID)

			// 人間プレイヤーがいなくなったらゲームを停止
			if gameInstance.GetHumanPlayerCount() == 0 {
				gameInstance.Stop()
				hub.RemoveGame(gameInstance.ID)
				log.Printf("Game stopped - no human players remaining")
			}
		}
	}
}
