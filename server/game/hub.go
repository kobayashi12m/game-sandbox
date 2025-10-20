package game

import (
	"game-sandbox/server/models"
	"game-sandbox/server/utils"
	"sync"
)

// Hub は複数のゲームルームを管理する
type Hub struct {
	games map[string]*Game
	mu    sync.RWMutex
}

// NewHub は新しいゲームハブを作成する
func NewHub() *Hub {
	return &Hub{
		games: make(map[string]*Game),
	}
}

// GetOrCreateGame は既存のゲームを取得するか、新しいゲームを作成する
func (h *Hub) GetOrCreateGame(roomID string) *Game {
	h.mu.Lock()
	defer h.mu.Unlock()

	if game, exists := h.games[roomID]; exists {
		return game
	}

	game := &Game{
		ID:          roomID,
		Players:     make(map[string]*models.Player),
		Running:     false,
		NPCCount:    utils.NPC_COUNT,  // constants.goから取得
		spatialGrid: NewSpatialGrid(), // 空間分割グリッドを初期化
	}

	// NPCを追加
	game.AddNPC(game.NPCCount)

	h.games[roomID] = game
	return h.games[roomID]
}

// RemoveGame はハブからゲームを削除する
func (h *Hub) RemoveGame(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.games, roomID)
}
