package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 開発中は全てのオリジンを許可
	},
}

// プレイヤー構造体
type Player struct {
	ID        string  `json:"id"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	CoreSize  float64 `json:"coreSize"`
	GuardSize float64 `json:"guardSize"`
}

// メッセージタイプ
type GameMessage struct {
	Type string      `json:"type"` // "chat", "gameInit", "gameState", "move"
	Data interface{} `json:"data"`
}

// 移動データ
type MoveData struct {
	ID string  `json:"id"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
}

// クライアント管理
type Client struct {
	conn   *websocket.Conn
	player *Player
}

// クライアント管理
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

var hub = &Hub{
	clients:    make(map[*Client]bool),
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
}

// Hub実行
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			fmt.Printf("新しいクライアントが接続しました。現在の接続数: %d\n", len(h.clients))

			// ゲーム初期化メッセージを送信
			initMsg := GameMessage{
				Type: "gameInit",
				Data: client.player,
			}
			msgBytes, _ := json.Marshal(initMsg)
			client.conn.WriteMessage(websocket.TextMessage, msgBytes)

			// 既存のプレイヤー情報を新規クライアントに送信
			h.mu.RLock()
			for existingClient := range h.clients {
				if existingClient != client && existingClient.player != nil {
					stateMsg := GameMessage{
						Type: "gameState",
						Data: existingClient.player,
					}
					stateMsgBytes, _ := json.Marshal(stateMsg)
					client.conn.WriteMessage(websocket.TextMessage, stateMsgBytes)
				}
			}
			h.mu.RUnlock()

			// 新規プレイヤーを既存のクライアントに通知
			newPlayerMsg := GameMessage{
				Type: "gameState",
				Data: client.player,
			}
			newPlayerMsgBytes, _ := json.Marshal(newPlayerMsg)
			h.mu.RLock()
			for existingClient := range h.clients {
				if existingClient != client {
					existingClient.conn.WriteMessage(websocket.TextMessage, newPlayerMsgBytes)
				}
			}
			h.mu.RUnlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.conn.Close()
				
				// 切断通知を他のクライアントに送信
				disconnectMsg := GameMessage{
					Type: "playerDisconnect",
					Data: client.player.ID,
				}
				disconnectMsgBytes, _ := json.Marshal(disconnectMsg)
				
				for remainingClient := range h.clients {
					remainingClient.conn.WriteMessage(websocket.TextMessage, disconnectMsgBytes)
				}
			}
			h.mu.Unlock()
			fmt.Printf("クライアントが切断しました。現在の接続数: %d\n", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
					client.conn.Close()
				}
			}
			h.mu.RUnlock()
		}
	}
}

var playerIDCounter = 0
var playerIDMutex sync.Mutex

func generatePlayerID() string {
	playerIDMutex.Lock()
	defer playerIDMutex.Unlock()
	playerIDCounter++
	return fmt.Sprintf("Player%d", playerIDCounter)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	// 新しいプレイヤーを作成
	player := &Player{
		ID:        generatePlayerID(),
		X:         400,
		Y:         300,
		CoreSize:  10,
		GuardSize: 30,
	}

	client := &Client{
		conn:   conn,
		player: player,
	}

	// クライアントを登録
	hub.register <- client

	// 切断時の処理
	defer func() {
		hub.unregister <- client
	}()

	// メッセージの読み取りループ
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("読み取りエラー:", err)
			break
		}

		if messageType == websocket.TextMessage {
			fmt.Printf("受信: %s\n", message)
			
			// メッセージをパースしてタイプを確認
			var gameMsg GameMessage
			if err := json.Unmarshal(message, &gameMsg); err != nil {
				// 古い形式のメッセージとして処理（後方互換性）
				hub.broadcast <- message
			} else {
				// 新しい形式のメッセージとして処理
				switch gameMsg.Type {
				case "chat":
					// チャットメッセージは全クライアントにブロードキャスト
					hub.broadcast <- message
				case "move":
					// 移動メッセージの処理
					moveDataJSON, _ := json.Marshal(gameMsg.Data)
					var moveData MoveData
					if err := json.Unmarshal(moveDataJSON, &moveData); err == nil {
						// プレイヤーの位置を更新
						client.player.X = moveData.X
						client.player.Y = moveData.Y
						
						// 全クライアントに位置更新を送信
						stateMsg := GameMessage{
							Type: "gameState",
							Data: client.player,
						}
						stateMsgBytes, _ := json.Marshal(stateMsg)
						hub.broadcast <- stateMsgBytes
					}
				}
			}
		}
	}
}

func main() {
	// Hubを起動
	go hub.run()

	http.HandleFunc("/ws", handleWebSocket)

	fmt.Println("サーバーを起動します: http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("サーバー起動エラー:", err)
	}
}