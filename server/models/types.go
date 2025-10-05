package models

import (
	"chess-mmo/server/utils"
	"time"

	"github.com/gorilla/websocket"
)

// Position はゲームフィールド上の座標を表す（浮動小数点）
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Snake はゲーム内の蛇を表す
type Snake struct {
	ID         string          `json:"id"`
	Body       []Position      `json:"body"`
	Direction  utils.Direction `json:"direction"`
	Color      string          `json:"color"`
	Alive      bool            `json:"alive"`
	Growing    int             `json:"-"`
	Respawning bool            `json:"-"`
	DeathTime  time.Time       `json:"-"`
	Speed      float64         `json:"-"` // 移動速度（ユニット/秒）
}

// Player はゲーム内のプレイヤーを表す
type Player struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Snake           *Snake `json:"snake"`
	Score           int    `json:"score"`
	Conn            *websocket.Conn
	IsNPC           bool      `json:"-"` // NPCかどうかのフラグ
	LastDirectionChange time.Time `json:"-"` // 最後に方向を変えた時刻
}

// GameState はクライアントに送信される現在の状態を表す
type GameState struct {
	Players []PlayerState `json:"players"`
	Food    []Position    `json:"food"`
}

// PlayerState はクライアント同期用のプレイヤーデータを表す
type PlayerState struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Snake *Snake `json:"snake"`
	Score int    `json:"score"`
}

// GameConfig はゲームの設定を表す
type GameConfig struct {
	FieldWidth      float64 `json:"fieldWidth"`
	FieldHeight     float64 `json:"fieldHeight"`
	SnakeRadius     float64 `json:"snakeRadius"`
	FoodRadius      float64 `json:"foodRadius"`
	CullingWidth    float64 `json:"cullingWidth"`
	CullingHeight   float64 `json:"cullingHeight"`
	CullingMargin   float64 `json:"cullingMargin"`
}
