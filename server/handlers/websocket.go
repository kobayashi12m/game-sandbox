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
			// Allow all origins (for development)
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// WebSocketHandler handles WebSocket connections
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

				gameInstance.Mu.Lock()
				gameInstance.AddPlayer(playerID, playerName, conn)
				player = gameInstance.Players[playerID]
				
				if len(gameInstance.Players) == 1 && !gameInstance.Running {
					gameInstance.Start()
				}
				gameInstance.Mu.Unlock()

				// Send join confirmation
				response := map[string]interface{}{
					"type":     "gameJoined",
					"playerId": playerID,
				}
				conn.WriteJSON(response)

				// Send current game state
				gameInstance.Mu.RLock()
				state := gameInstance.GetState()
				gameInstance.Mu.RUnlock()
				
				stateMsg := map[string]interface{}{
					"type":  "gameState",
					"state": state,
				}
				conn.WriteJSON(stateMsg)

			case "changeDirection":
				if player == nil || gameInstance == nil {
					continue
				}
				
				direction, _ := msg["direction"].(string)
				if newDir, ok := utils.DIRECTIONS[direction]; ok {
					gameInstance.Mu.Lock()
					if player.Snake.Alive {
						player.Snake.ChangeDirection(newDir)
					}
					gameInstance.Mu.Unlock()
				}
			}
		}

		// Clean up on disconnect
		if gameInstance != nil && playerID != "" {
			gameInstance.Mu.Lock()
			gameInstance.RemovePlayer(playerID)
			if len(gameInstance.Players) == 0 {
				gameInstance.Running = false
				hub.RemoveGame(gameInstance.ID)
			}
			gameInstance.Mu.Unlock()
		}
	}
}