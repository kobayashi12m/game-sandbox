package models

import (
	"github.com/gorilla/websocket"
	"chess-mmo/server/utils"
)

// Position represents coordinates on the game grid
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Snake represents a snake in the game
type Snake struct {
	ID        string          `json:"id"`
	Body      []Position      `json:"body"`
	Direction utils.Direction `json:"direction"`
	Color     string          `json:"color"`
	Alive     bool            `json:"alive"`
	Growing   int             `json:"-"`
}

// Player represents a player in the game
type Player struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Snake *Snake `json:"snake"`
	Score int    `json:"score"`
	Conn  *websocket.Conn
}

// GameState represents the current state sent to clients
type GameState struct {
	Players []PlayerState `json:"players"`
	Food    []Position    `json:"food"`
}

// PlayerState represents player data for client synchronization
type PlayerState struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Snake *Snake `json:"snake"`
	Score int    `json:"score"`
}