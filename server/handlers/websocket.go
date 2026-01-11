package handlers

import (
	"encoding/json"
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

		// WebSocket接続確立ログ
		utils.Info("WebSocket established", map[string]interface{}{
			"event":       "ws_established",
			"remote_addr": r.RemoteAddr,
		})

		var player *models.Player
		var gameInstance *game.Game
		var client *game.Client
		var playerID string

		// メッセージ読み取りループ
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

				// プレイヤーを追加（内部でClientも作成される）
				gameInstance.AddPlayer(playerID, playerName, conn)
				player, _ = gameInstance.GetPlayer(playerID)
				client, _ = gameInstance.GetClient(playerID)

				utils.LogConnectionEvent("connect", playerID, playerName, false)

				gameInstance.ShouldStart()

				// 初期データを送信
				sendInitialData(client, player, gameInstance)

			case "setAcceleration":
				if player == nil || gameInstance == nil {
					continue
				}

				x, xOk := msg["x"].(float64)
				y, yOk := msg["y"].(float64)
				if xOk && yOk && player.Celestial != nil && player.Celestial.Alive {
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
					gameInstance.SendCommand(game.ShootCommand{
						Player:  player,
						TargetX: targetX,
						TargetY: targetY,
					})
				}
			}
		}

		// 切断時のクリーンアップ
		cleanup(gameInstance, hub, player, playerID)
	}
}

// sendInitialData は接続時の初期データを送信する
func sendInitialData(client *game.Client, player *models.Player, gameInstance *game.Game) {
	if client == nil || player == nil {
		return
	}

	// 参加確認を送信
	joinResponse := map[string]interface{}{
		"type":     "gameJoined",
		"playerId": player.ID,
	}
	if data, err := json.Marshal(joinResponse); err == nil {
		client.Send(data)
	}

	// ゲーム設定を送信
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
	if data, err := json.Marshal(configMsg); err == nil {
		client.Send(data)
	}

	// 現在のゲーム状態を送信
	state := gameInstance.GetState()
	stateMsg := map[string]interface{}{
		"type":  "gameState",
		"state": state,
	}
	if data, err := json.Marshal(stateMsg); err == nil {
		client.Send(data)
	}

	// スコアボード情報を送信
	scoreboard := gameInstance.GetScoreboard()
	myScore := models.ScoreInfo{
		ID:    player.ID,
		Name:  player.Name,
		Score: player.Score,
		Alive: player.Celestial.Alive,
		Color: player.Celestial.Color,
	}
	scoreMsg := map[string]interface{}{
		"type":       "scoreboard",
		"scoreboard": scoreboard,
		"myScore":    myScore,
	}
	if data, err := json.Marshal(scoreMsg); err == nil {
		client.Send(data)
	}
}

// cleanup は切断時のクリーンアップ処理を行う
func cleanup(gameInstance *game.Game, hub *game.Hub, player *models.Player, playerID string) {
	if gameInstance == nil || playerID == "" {
		return
	}

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
