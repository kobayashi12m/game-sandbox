package game

import (
	"sync"
	"time"

	"game-sandbox/server/models"
	"game-sandbox/server/utils"

	"github.com/gorilla/websocket"
)

const (
	// 送信キューのサイズ（これを超えると古いメッセージをドロップ）
	sendQueueSize = 10

	// 書き込みタイムアウト
	writeTimeout = 100 * time.Millisecond
)

// Client はWebSocket接続をラップし、非同期書き込みを提供する
type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	Player *models.Player // プレイヤーへの参照
	closed bool
	mu     sync.RWMutex
}

// NewClient は新しいClientを作成し、書き込みgoroutineを開始する
func NewClient(conn *websocket.Conn, player *models.Player) *Client {
	c := &Client{
		conn:   conn,
		send:   make(chan []byte, sendQueueSize),
		Player: player,
	}
	go c.writePump()
	return c
}

// Send はメッセージを非同期で送信する（ブロックしない）
// キューが満杯の場合は古いメッセージをドロップして新しいメッセージを追加
func (c *Client) Send(data []byte) bool {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return false
	}
	c.mu.RUnlock()

	select {
	case c.send <- data:
		return true
	default:
		// キューが満杯：古いメッセージを1つ捨てて新しいのを追加
		select {
		case <-c.send:
			// 古いメッセージを捨てた
		default:
		}
		select {
		case c.send <- data:
			return true
		default:
			return false
		}
	}
}

// Close はClientを閉じる
func (c *Client) Close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	close(c.send)
	c.mu.Unlock()

	c.conn.Close()
}

// IsClosed はClientが閉じているかを返す
func (c *Client) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// GetConn は生のWebSocket接続を返す（読み取り用）
func (c *Client) GetConn() *websocket.Conn {
	return c.conn
}

// writePump は送信キューからメッセージを読み取り、WebSocketに書き込む
func (c *Client) writePump() {
	playerID := ""
	if c.Player != nil {
		playerID = c.Player.ID
	}

	defer func() {
		if r := recover(); r != nil {
			utils.LogPanicRecovery("client_writePump", playerID, r)
		}
		c.conn.Close()
	}()

	for data := range c.send {
		c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))

		start := time.Now()
		if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			utils.LogWebSocketError(playerID, "client_write", err)
			c.Close()
			return
		}

		duration := time.Since(start)
		utils.LogPerformanceWarning("websocket_write", duration, 10*time.Millisecond)
	}
}
