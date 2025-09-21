package game

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"sync"
	"time"

	"chess-mmo/server/models"
	"chess-mmo/server/utils"

	"github.com/gorilla/websocket"
)

// Game はゲームセッションを表す
type Game struct {
	ID      string
	Players map[string]*models.Player
	Food    []models.Position
	Running bool
	mu      sync.RWMutex
}

// AddPlayer はゲームに新しいプレイヤーを追加する
func (g *Game) AddPlayer(id, name string, conn *websocket.Conn) {
	g.mu.Lock()
	defer g.mu.Unlock()

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

// RemovePlayer はゲームからプレイヤーを削除する
func (g *Game) RemovePlayer(id string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.Players, id)
}

// GenerateFood はゲームグリッドに食べ物を生成する
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

// IsPositionOccupied は指定された位置が蛇に占有されているかチェックする
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

// Update はゲームの1ティックを処理する
func (g *Game) Update() {
	if !g.Running {
		return
	}

	// 全ての蛇を移動
	for _, player := range g.Players {
		player.Snake.Move()
	}

	// 衝突判定
	for _, player := range g.Players {
		if !player.Snake.Alive {
			continue
		}

		// 自己衝突
		if player.Snake.CheckSelfCollision() {
			player.Snake.Alive = false
			player.Score -= 10
			if player.Score < 0 {
				player.Score = 0
			}
			continue
		}

		// 他の蛇との衝突
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

		// 食べ物との衝突判定
		head := player.Snake.Body[0]
		for i := len(g.Food) - 1; i >= 0; i-- {
			if g.Food[i].X == head.X && g.Food[i].Y == head.Y {
				player.Snake.Grow(3)
				player.Score += 10
				g.Food = append(g.Food[:i], g.Food[i+1:]...)
			}
		}
	}

	// 必要に応じて食べ物を再生成
	if len(g.Food) < 3 {
		g.GenerateFood()
	}

	// 死んだ蛇の処理
	g.handleDeadSnakes()
}

// handleDeadSnakes は死んだ蛇の復活処理を管理する
func (g *Game) handleDeadSnakes() {
	now := time.Now()
	for _, player := range g.Players {
		snake := player.Snake

		if !snake.Alive && !snake.Respawning {
			// 死亡時の初期化
			snake.Respawning = true
			snake.DeathTime = now
			snake.Body = []models.Position{} // 即座にクリア
		}

		if snake.Respawning && now.Sub(snake.DeathTime) >= 3*time.Second {
			// 復活
			snake.Reset()
			snake.Respawning = false
		}
	}
}

// GetState はクライアント用の現在のゲーム状態を返す
func (g *Game) GetState() models.GameState {
	g.mu.RLock()
	defer g.mu.RUnlock()

	players := make([]models.PlayerState, 0, len(g.Players))
	for _, p := range g.Players {
		// 元のデータを変更しないよう蛇のコピーを作成
		snakeCopy := *p.Snake

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

// Broadcast はゲーム内の全プレイヤーにメッセージを送信する
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

// Start はゲームループを開始する
func (g *Game) Start() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.Running = true
	g.GenerateFood()
	go g.RunGameLoop()
}

// GetPlayer はIDでプレイヤーを取得する
func (g *Game) GetPlayer(id string) (*models.Player, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	player, exists := g.Players[id]
	return player, exists
}

// GetPlayerCount はゲーム内のプレイヤー数を返す
func (g *Game) GetPlayerCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.Players)
}

// IsRunning はゲームが実行中かどうかを返す
func (g *Game) IsRunning() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Running
}

// ShouldStart はゲームを開始すべきかチェックし、必要なら開始する
func (g *Game) ShouldStart() bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(g.Players) == 1 && !g.Running {
		g.Running = true
		g.GenerateFood()
		go g.RunGameLoop()
		return true
	}
	return false
}

// ChangePlayerDirection はプレイヤーの蛇の方向を変更する
func (g *Game) ChangePlayerDirection(playerID string, direction utils.Direction) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if player, exists := g.Players[playerID]; exists && player.Snake.Alive {
		player.Snake.ChangeDirection(direction)
	}
}

// Stop はゲームを安全に停止する
func (g *Game) Stop() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Running = false
}

// RunGameLoop はメインゲームの更新ループを実行する
func (g *Game) RunGameLoop() {
	ticker := time.NewTicker(utils.INITIAL_SPEED)
	defer ticker.Stop()

	for g.Running {
		<-ticker.C
		g.mu.Lock()
		g.Update()
		g.mu.Unlock()

		state := g.GetState()

		message := map[string]interface{}{
			"type":  "gameState",
			"state": state,
		}
		g.Broadcast(message)
	}
}
