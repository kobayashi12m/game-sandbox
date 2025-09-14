package game

import (
	"encoding/json"
	"log"
	"sync"
	"time"
	"math/rand/v2"

	"chess-mmo/server/models"
	"chess-mmo/server/utils"
	"github.com/gorilla/websocket"
)

// Game represents a game session
type Game struct {
	ID      string
	Players map[string]*models.Player
	Food    []models.Position
	Running bool
	Mu      sync.RWMutex
}

// AddPlayer adds a new player to the game
func (g *Game) AddPlayer(id, name string, conn *websocket.Conn) {
	colors := []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#DDA0DD", "#F4A460"}
	color := colors[len(g.Players)%len(colors)]
	
	snake := &models.Snake{
		ID:    id,
		Color: color,
	}
	snake.Reset()
	
	g.Players[id] = &models.Player{
		ID:    id,
		Name:  name,
		Snake: snake,
		Score: 0,
		Conn:  conn,
	}
}

// RemovePlayer removes a player from the game
func (g *Game) RemovePlayer(id string) {
	delete(g.Players, id)
}

// GenerateFood creates food items on the game grid
func (g *Game) GenerateFood() {
	g.Food = []models.Position{}
	foodCount := 3
	if len(g.Players) > 0 {
		foodCount = int(float64(len(g.Players)) * 1.5)
		if foodCount < 3 {
			foodCount = 3
		}
	}

	for i := 0; i < foodCount; i++ {
		var pos models.Position
		attempts := 0
		for {
			pos = models.Position{
				X: rand.IntN(utils.GRID_SIZE),
				Y: rand.IntN(utils.GRID_SIZE),
			}
			if !g.IsPositionOccupied(pos) || attempts > 100 {
				break
			}
			attempts++
		}
		if attempts <= 100 {
			g.Food = append(g.Food, pos)
		}
	}
}

// IsPositionOccupied checks if a position is occupied by any snake
func (g *Game) IsPositionOccupied(pos models.Position) bool {
	for _, player := range g.Players {
		for _, segment := range player.Snake.Body {
			if segment.X == pos.X && segment.Y == pos.Y {
				return true
			}
		}
	}
	return false
}

// Update processes one game tick
func (g *Game) Update() {
	if !g.Running {
		return
	}

	// Move all snakes
	for _, player := range g.Players {
		player.Snake.Move()
	}

	// Check collisions
	for _, player := range g.Players {
		if !player.Snake.Alive {
			continue
		}

		// Self collision
		if player.Snake.CheckSelfCollision() {
			player.Snake.Alive = false
			player.Score -= 10
			if player.Score < 0 {
				player.Score = 0
			}
			continue
		}

		// Collision with other snakes
		for _, otherPlayer := range g.Players {
			if player.ID != otherPlayer.ID && player.Snake.CheckCollisionWith(otherPlayer.Snake) {
				player.Snake.Alive = false
				player.Score -= 10
				if player.Score < 0 {
					player.Score = 0
				}
				otherPlayer.Score += 5
				break
			}
		}

		// Check food collision
		head := player.Snake.Body[0]
		for i := len(g.Food) - 1; i >= 0; i-- {
			if g.Food[i].X == head.X && g.Food[i].Y == head.Y {
				player.Snake.Grow(3)
				player.Score += 10
				g.Food = append(g.Food[:i], g.Food[i+1:]...)
			}
		}
	}

	// Regenerate food if needed
	if len(g.Food) < 3 {
		g.GenerateFood()
	}

	// Respawn dead snakes (only if not already respawning)
	for _, player := range g.Players {
		if !player.Snake.Alive && player.Snake.Growing != -1 {
			// Use Growing=-1 as a flag to prevent multiple respawn goroutines
			player.Snake.Growing = -1
			go func(p *models.Player) {
				time.Sleep(3 * time.Second)
				g.Mu.Lock()
				if !p.Snake.Alive { // Double check in case player was manually revived
					p.Snake.Reset()
				}
				g.Mu.Unlock()
			}(player)
		}
	}
}

// GetState returns the current game state for clients
func (g *Game) GetState() models.GameState {
	players := make([]models.PlayerState, 0, len(g.Players))
	for _, p := range g.Players {
		// Create a copy of the snake to avoid modifying original
		snakeCopy := *p.Snake
		
		// If snake is dead, clear its body to prevent visual glitches
		if !snakeCopy.Alive {
			snakeCopy.Body = []models.Position{}
		}
		
		players = append(players, models.PlayerState{
			ID:    p.ID,
			Name:  p.Name,
			Snake: &snakeCopy,
			Score: p.Score,
		})
	}
	return models.GameState{
		Players: players,
		Food:    g.Food,
	}
}

// Broadcast sends a message to all players in the game
func (g *Game) Broadcast(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		return
	}

	for _, player := range g.Players {
		if err := player.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Error broadcasting to player %s: %v", player.ID, err)
		}
	}
}

// Start begins the game loop
func (g *Game) Start() {
	g.Running = true
	g.GenerateFood()
	go g.RunGameLoop()
}

// RunGameLoop runs the main game update loop
func (g *Game) RunGameLoop() {
	ticker := time.NewTicker(utils.INITIAL_SPEED)
	defer ticker.Stop()

	for g.Running {
		select {
		case <-ticker.C:
			g.Mu.Lock()
			g.Update()
			state := g.GetState()
			g.Mu.Unlock()

			message := map[string]interface{}{
				"type":  "gameState",
				"state": state,
			}
			g.Broadcast(message)
		}
	}
}