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
	ID       string
	Players  map[string]*models.Player
	Food     []models.Position
	Running  bool
	NPCCount int // NPC数の設定
	mu       sync.RWMutex
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

// GetHumanPlayerCount は人間プレイヤーの数を返す
func (g *Game) GetHumanPlayerCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	count := 0
	for _, player := range g.Players {
		if !player.IsNPC {
			count++
		}
	}
	return count
}

// GenerateFood はゲームフィールドに食べ物を生成する
func (g *Game) GenerateFood() {
	targetFoodCount := 5  // 最小数を増加
	if len(g.Players) > 0 {
		targetFoodCount = int(float64(len(g.Players)) * 2.0)  // プレイヤー数の2倍に増加
		if targetFoodCount < 5 {
			targetFoodCount = 5
		}
	}

	initialFoodCount := len(g.Food)
	
	for len(g.Food) < targetFoodCount {
		var pos models.Position
		attempts := 0
		for {
			pos = models.Position{
				X: rand.Float64() * utils.FIELD_WIDTH,
				Y: rand.Float64() * utils.FIELD_HEIGHT,
			}
			if !g.IsPositionOccupied(pos) || attempts > 100 {
				break
			}
			attempts++
		}
		if attempts <= 100 {
			g.Food = append(g.Food, pos)
		} else {
			log.Printf("Failed to place food after 100 attempts (current food: %d, target: %d)", 
				len(g.Food), targetFoodCount)
			break  // 無限ループを防ぐ
		}
	}
	
	// 食べ物が生成された場合のみログ出力
	if len(g.Food) > initialFoodCount {
		log.Printf("Generated food: %d -> %d (target: %d, players: %d)", 
			initialFoodCount, len(g.Food), targetFoodCount, len(g.Players))
	}
}

// IsPositionOccupied は指定された位置が蛇に占有されているかチェックする
func (g *Game) IsPositionOccupied(pos models.Position) bool {
	for _, player := range g.Players {
		for _, segment := range player.Snake.Body {
			dx := segment.X - pos.X
			dy := segment.Y - pos.Y
			dist := dx*dx + dy*dy
			if dist < (utils.SNAKE_RADIUS+utils.FOOD_RADIUS)*(utils.SNAKE_RADIUS+utils.FOOD_RADIUS) {
				return true
			}
		}
	}
	return false
}

// Update はゲームの1ティックを処理する
func (g *Game) Update(deltaTime float64) {
	if !g.Running {
		return
	}

	// 全ての蛇を移動
	for _, player := range g.Players {
		player.Snake.Move(deltaTime)
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
			// 蛇の頭と食べ物の距離をチェック
			dx := head.X - g.Food[i].X
			dy := head.Y - g.Food[i].Y
			dist := dx*dx + dy*dy

			if dist < (utils.SNAKE_RADIUS+utils.FOOD_RADIUS)*(utils.SNAKE_RADIUS+utils.FOOD_RADIUS) {
				// 食べ物を除去
				g.Food = append(g.Food[:i], g.Food[i+1:]...)
				// 蛇を成長させる
				player.Snake.Growing = 3
				player.Score += 10
				log.Printf("Player %s ate food! Remaining food: %d", player.Name, len(g.Food))
				break
			}
		}
	}

	// 死んだ蛇のリスポーン処理
	for _, player := range g.Players {
		if !player.Snake.Alive && !player.Snake.Respawning {
			player.Snake.Respawning = true
			player.Snake.DeathTime = time.Now()
		}

		if player.Snake.Respawning && time.Since(player.Snake.DeathTime) > 3*time.Second {
			player.Snake.Reset()
			player.Snake.Respawning = false
		}
	}

	// 食べ物の補充
	g.GenerateFood()
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
		// NPCプレイヤーにはメッセージを送信しない
		if player.IsNPC || player.Conn == nil {
			continue
		}
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

	// 人間プレイヤーが1人以上いて、ゲームが開始されていない場合
	humanPlayers := 0
	for _, player := range g.Players {
		if !player.IsNPC {
			humanPlayers++
		}
	}

	if humanPlayers >= 1 && !g.Running {
		g.Running = true
		g.GenerateFood()
		go g.RunGameLoop()
		log.Printf("Game started with %d human players and %d total players", humanPlayers, len(g.Players))
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
	ticker := time.NewTicker(utils.GAME_TICK)
	defer ticker.Stop()
	lastUpdate := time.Now()

	for g.Running {
		<-ticker.C
		now := time.Now()
		deltaTime := now.Sub(lastUpdate).Seconds()
		lastUpdate = now

		g.mu.Lock()
		// NPCの方向を更新
		g.updateNPCDirections()
		g.Update(deltaTime)
		g.mu.Unlock()

		state := g.GetState()

		message := map[string]interface{}{
			"type":  "gameState",
			"state": state,
		}
		g.Broadcast(message)
	}
}