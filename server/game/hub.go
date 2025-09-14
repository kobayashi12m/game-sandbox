package game

import (
	"sync"
	"chess-mmo/server/models"
)

// Hub manages multiple game rooms
type Hub struct {
	games map[string]*Game
	mu    sync.RWMutex
}

// NewHub creates a new game hub
func NewHub() *Hub {
	return &Hub{
		games: make(map[string]*Game),
	}
}

// GetOrCreateGame gets an existing game or creates a new one
func (h *Hub) GetOrCreateGame(roomID string) *Game {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if game, exists := h.games[roomID]; exists {
		return game
	}
	
	h.games[roomID] = &Game{
		ID:      roomID,
		Players: make(map[string]*models.Player),
		Running: false,
	}
	return h.games[roomID]
}

// RemoveGame removes a game from the hub
func (h *Hub) RemoveGame(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.games, roomID)
}