package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 開発中は全てのオリジンを許可
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	fmt.Println("新しいクライアントが接続しました")

	// メッセージの読み取りループ
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("読み取りエラー:", err)
			break
		}

		fmt.Printf("受信: %s\n", message)

		// エコーバック
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Println("書き込みエラー:", err)
			break
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)

	fmt.Println("サーバーを起動します: http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("サーバー起動エラー:", err)
	}
}
