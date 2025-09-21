package models

import (
	"github.com/gorilla/websocket"
	"chess-mmo/server/utils"
)

// Position はゲームグリッド上の座標を表す
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Snake はゲーム内の蛇を表す
type Snake struct {
	ID        string          `json:"id"`
	Body      []Position      `json:"body"`
	Direction utils.Direction `json:"direction"`
	Color     string          `json:"color"`
	Alive     bool            `json:"alive"`
	Growing   int             `json:"-"`
}

// Player はゲーム内のプレイヤーを表す
type Player struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Snake *Snake `json:"snake"`
	Score int    `json:"score"`
	Conn  *websocket.Conn
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