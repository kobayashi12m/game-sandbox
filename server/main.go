package main

import (
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

// クライアント管理
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.RWMutex
}

var hub = &Hub{
	clients:    make(map[*websocket.Conn]bool),
	broadcast:  make(chan []byte),
	register:   make(chan *websocket.Conn),
	unregister: make(chan *websocket.Conn),
}

// Hub実行
func (h *Hub) run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.clients[conn] = true
			h.mu.Unlock()
			fmt.Printf("新しいクライアントが接続しました。現在の接続数: %d\n", len(h.clients))

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
			h.mu.Unlock()
			fmt.Printf("クライアントが切断しました。現在の接続数: %d\n", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for conn := range h.clients {
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					conn.Close()
				}
			}
			h.mu.RUnlock()
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	// クライアントを登録
	hub.register <- conn

	// 切断時の処理
	defer func() {
		hub.unregister <- conn
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
			// 全クライアントにブロードキャスト
			hub.broadcast <- message
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
