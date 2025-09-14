package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	mathrand "math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	GRID_SIZE     = 40
	INITIAL_SPEED = 100 * time.Millisecond
)

type Direction struct {
	X int `json:"x"`
	Y int `json:"y"`
}

var (
	DIRECTIONS = map[string]Direction{
		"UP":    {X: 0, Y: -1},
		"DOWN":  {X: 0, Y: 1},
		"LEFT":  {X: -1, Y: 0},
		"RIGHT": {X: 1, Y: 0},
	}
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Snake struct {
	ID        string     `json:"id"`
	Body      []Position `json:"body"`
	Direction Direction  `json:"direction"`
	Color     string     `json:"color"`
	Alive     bool       `json:"alive"`
	Growing   int        `json:"-"`
}

type Player struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Snake *Snake `json:"snake"`
	Score int    `json:"score"`
	Conn  *websocket.Conn
}

type GameState struct {
	Players []PlayerState `json:"players"`
	Food    []Position    `json:"food"`
}

type PlayerState struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Snake *Snake `json:"snake"`
	Score int    `json:"score"`
}

type Game struct {
	ID      string
	Players map[string]*Player
	Food    []Position
	Running bool
	mu      sync.RWMutex
}

type Hub struct {
	games map[string]*Game
	mu    sync.RWMutex
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// すべてのオリジンを許可（開発環境用）
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	hub = &Hub{
		games: make(map[string]*Game),
	}
)

func (s *Snake) Reset() {
	startX := mathrand.Intn(GRID_SIZE-10) + 5
	startY := mathrand.Intn(GRID_SIZE-10) + 5
	s.Body = []Position{
		{X: startX, Y: startY},
		{X: startX - 1, Y: startY},
		{X: startX - 2, Y: startY},
	}
	s.Direction = DIRECTIONS["RIGHT"]
	s.Growing = 0
	s.Alive = true
}

func (s *Snake) Move() {
	if !s.Alive {
		return
	}

	head := s.Body[0]
	newHead := Position{
		X: head.X + s.Direction.X,
		Y: head.Y + s.Direction.Y,
	}

	// Wrap around edges
	if newHead.X < 0 {
		newHead.X = GRID_SIZE - 1
	} else if newHead.X >= GRID_SIZE {
		newHead.X = 0
	}
	if newHead.Y < 0 {
		newHead.Y = GRID_SIZE - 1
	} else if newHead.Y >= GRID_SIZE {
		newHead.Y = 0
	}

	s.Body = append([]Position{newHead}, s.Body...)

	if s.Growing > 0 {
		s.Growing--
	} else {
		s.Body = s.Body[:len(s.Body)-1]
	}
}

func (s *Snake) Grow(amount int) {
	s.Growing += amount
}

func (s *Snake) CheckSelfCollision() bool {
	head := s.Body[0]
	for i := 1; i < len(s.Body); i++ {
		if head.X == s.Body[i].X && head.Y == s.Body[i].Y {
			return true
		}
	}
	return false
}

func (s *Snake) CheckCollisionWith(other *Snake) bool {
	if !other.Alive {
		return false
	}
	head := s.Body[0]
	for _, segment := range other.Body {
		if head.X == segment.X && head.Y == segment.Y {
			return true
		}
	}
	return false
}

func (s *Snake) ChangeDirection(newDir Direction) {
	// Prevent reverse direction
	if s.Direction.X == -newDir.X && s.Direction.Y == -newDir.Y {
		return
	}
	s.Direction = newDir
}

func (g *Game) AddPlayer(id, name string, conn *websocket.Conn) {
	colors := []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#DDA0DD", "#F4A460"}
	color := colors[len(g.Players)%len(colors)]
	
	snake := &Snake{
		ID:    id,
		Color: color,
	}
	snake.Reset()
	
	g.Players[id] = &Player{
		ID:    id,
		Name:  name,
		Snake: snake,
		Score: 0,
		Conn:  conn,
	}
}

func (g *Game) RemovePlayer(id string) {
	delete(g.Players, id)
}

func (g *Game) GenerateFood() {
	g.Food = []Position{}
	foodCount := 3
	if len(g.Players) > 0 {
		foodCount = int(float64(len(g.Players)) * 1.5)
		if foodCount < 3 {
			foodCount = 3
		}
	}

	for i := 0; i < foodCount; i++ {
		var pos Position
		attempts := 0
		for {
			pos = Position{
				X: mathrand.Intn(GRID_SIZE),
				Y: mathrand.Intn(GRID_SIZE),
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

func (g *Game) IsPositionOccupied(pos Position) bool {
	for _, player := range g.Players {
		for _, segment := range player.Snake.Body {
			if segment.X == pos.X && segment.Y == pos.Y {
				return true
			}
		}
	}
	return false
}

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

	// Respawn dead snakes
	for _, player := range g.Players {
		if !player.Snake.Alive {
			go func(p *Player) {
				time.Sleep(3 * time.Second)
				g.mu.Lock()
				p.Snake.Reset()
				g.mu.Unlock()
			}(player)
		}
	}
}

func (g *Game) GetState() GameState {
	players := make([]PlayerState, 0, len(g.Players))
	for _, p := range g.Players {
		players = append(players, PlayerState{
			ID:    p.ID,
			Name:  p.Name,
			Snake: p.Snake,
			Score: p.Score,
		})
	}
	return GameState{
		Players: players,
		Food:    g.Food,
	}
}

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

func (g *Game) Start() {
	g.Running = true
	g.GenerateFood()
	go g.RunGameLoop()
}

func (g *Game) RunGameLoop() {
	ticker := time.NewTicker(INITIAL_SPEED)
	defer ticker.Stop()

	for g.Running {
		select {
		case <-ticker.C:
			g.mu.Lock()
			g.Update()
			state := g.GetState()
			g.mu.Unlock()

			message := map[string]interface{}{
				"type":  "gameState",
				"state": state,
			}
			g.Broadcast(message)
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	var player *Player
	var game *Game
	var playerID string

	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		msgType, ok := msg["type"].(string)
		if !ok {
			continue
		}

		switch msgType {
		case "join":
			roomID, _ := msg["roomId"].(string)
			playerName, _ := msg["playerName"].(string)
			if playerName == "" {
				playerName = "Player"
			}

			playerID = generateID()

			hub.mu.Lock()
			if _, exists := hub.games[roomID]; !exists {
				hub.games[roomID] = &Game{
					ID:      roomID,
					Players: make(map[string]*Player),
					Running: false,
				}
			}
			game = hub.games[roomID]
			hub.mu.Unlock()

			game.mu.Lock()
			game.AddPlayer(playerID, playerName, conn)
			player = game.Players[playerID]
			
			if len(game.Players) == 1 && !game.Running {
				game.Start()
			}
			game.mu.Unlock()

			// Send join confirmation
			response := map[string]interface{}{
				"type":     "gameJoined",
				"playerId": playerID,
			}
			conn.WriteJSON(response)

			// Send current game state
			game.mu.RLock()
			state := game.GetState()
			game.mu.RUnlock()
			
			stateMsg := map[string]interface{}{
				"type":  "gameState",
				"state": state,
			}
			conn.WriteJSON(stateMsg)

		case "changeDirection":
			if player == nil || game == nil {
				continue
			}
			
			direction, _ := msg["direction"].(string)
			if newDir, ok := DIRECTIONS[direction]; ok {
				game.mu.Lock()
				if player.Snake.Alive {
					player.Snake.ChangeDirection(newDir)
				}
				game.mu.Unlock()
			}
		}
	}

	// Clean up on disconnect
	if game != nil && playerID != "" {
		game.mu.Lock()
		game.RemovePlayer(playerID)
		if len(game.Players) == 0 {
			game.Running = false
			hub.mu.Lock()
			delete(hub.games, game.ID)
			hub.mu.Unlock()
		}
		game.mu.Unlock()
	}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func main() {
	mathrand.Seed(time.Now().UnixNano())

	http.HandleFunc("/ws", handleWebSocket)
	
	// Serve static files
	fs := http.FileServer(http.Dir("../client/dist"))
	http.Handle("/", fs)

	log.Println("Snake game server running on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}