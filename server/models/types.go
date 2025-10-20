package models

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Position はゲームフィールド上の座標を表す（浮動小数点）
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Food はゲーム内の食べ物を表す
type Food struct {
	Position Position `json:"position"`
}

// Player はゲーム内のプレイヤーを表す
type Player struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	Celestial           *Celestial `json:"celestial"`
	Score               int        `json:"score"`
	Conn                *websocket.Conn
	IsNPC               bool       `json:"-"` // NPCかどうかのフラグ
	LastDirectionChange time.Time  `json:"-"` // 最後に方向を変えた時刻
	ConnMu              sync.Mutex `json:"-"` // WebSocket書き込み用mutex
}

// GameState はクライアントに送信される現在の状態を表す
type GameState struct {
	Players []PlayerState `json:"players"`
	Food    []Position    `json:"food"`
}

// GridLine はSpatialGridの可視化用の線を表す
type GridLine struct {
	StartX float64 `json:"startX"`
	StartY float64 `json:"startY"`
	EndX   float64 `json:"endX"`
	EndY   float64 `json:"endY"`
}

// PlayerState はクライアント同期用のプレイヤーデータを表す
type PlayerState struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Celestial *Celestial `json:"celestial"`
	Score     int        `json:"score"`
}

// ScoreInfo はスコアボード用の軽量プレイヤー情報を表す
type ScoreInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
	Alive bool   `json:"alive"`
	Color string `json:"color"`
}

// ScoreUpdate はスコアボードの更新情報を表す（さらに軽量化）
type ScoreUpdate struct {
	Players []ScoreInfo `json:"players"`
}

// GameConfig はゲームの設定を表す
type GameConfig struct {
	FieldWidth    float64    `json:"fieldWidth"`
	FieldHeight   float64    `json:"fieldHeight"`
	SphereRadius  float64    `json:"sphereRadius"`
	FoodRadius    float64    `json:"foodRadius"`
	CullingWidth  float64    `json:"cullingWidth"`
	CullingHeight float64    `json:"cullingHeight"`
	GridLines     []GridLine `json:"gridLines,omitempty"` // SpatialGrid可視化用
}
