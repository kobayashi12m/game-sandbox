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
	ID          string
	Players     map[string]*models.Player
	Food        []*models.Food // 食べ物をポインタで管理
	Running     bool
	NPCCount    int          // NPC数の設定
	spatialGrid *SpatialGrid // 空間分割グリッド
	frameCount  int64        // フレームカウンター
	mu          sync.RWMutex
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
	targetFoodCount := 5 // 最小数を増加
	if len(g.Players) > 0 {
		// プレイヤー数の2倍に増加
		targetFoodCount = max(int(float64(len(g.Players))*2.0), 5)
	}

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
			g.Food = append(g.Food, &models.Food{Position: pos})
		} else {
			log.Printf("Failed to place food after 100 attempts (current food: %d, target: %d)",
				len(g.Food), targetFoodCount)
			break // 無限ループを防ぐ
		}
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

// RemoveFood は食べ物をポインタで効率的に削除する
func (g *Game) RemoveFood(targetFood *models.Food) {
	for i, food := range g.Food {
		if food == targetFood {
			g.Food = append(g.Food[:i], g.Food[i+1:]...)
			return
		}
	}
}

// UpdateSpatialGrid は空間分割グリッドを更新する
func (g *Game) UpdateSpatialGrid() {
	// グリッドをクリア
	g.spatialGrid.Clear()

	// プレイヤーの全セグメントをグリッドに追加
	for playerID, player := range g.Players {
		if player.Snake.Alive && len(player.Snake.Body) > 0 {
			// 蛇の全セグメントをグリッドに登録
			g.spatialGrid.AddPlayerSegments(playerID, player.Snake.Body)
		}
	}

	// 食べ物をグリッドに追加
	for _, food := range g.Food {
		g.spatialGrid.AddFood(food)
	}
}

// Update はゲームの1ティックを処理する
func (g *Game) Update(deltaTime float64) {
	if !g.Running {
		return
	}

	// フレームカウンターを増加
	g.frameCount++

	// デバッグ用：詳細なゲーム状態をログ出力
	if g.frameCount%300 == 0 { // 5秒に1回
		totalSegments := 0
		humanPlayers := 0
		maxSnakeLength := 0
		minSnakeLength := 999999
		deadPlayers := 0

		for _, player := range g.Players {
			segments := len(player.Snake.Body)
			totalSegments += segments

			if !player.IsNPC {
				humanPlayers++
			}

			if segments > maxSnakeLength {
				maxSnakeLength = segments
			}
			if segments < minSnakeLength {
				minSnakeLength = segments
			}

			if !player.Snake.Alive {
				deadPlayers++
			}
		}

		log.Printf("🎮 SERVER STATE: Frame %d | Players: %d (Human: %d, Dead: %d) | Food: %d | Segments: %d (Max: %d, Min: %d)",
			g.frameCount, len(g.Players), humanPlayers, deadPlayers, len(g.Food), totalSegments, maxSnakeLength, minSnakeLength)
	}

	// 全ての蛇を移動
	for _, player := range g.Players {
		player.Snake.Move(deltaTime)
	}

	// 空間分割グリッドを3フレームに1回更新（負荷軽減）
	if g.frameCount%3 == 0 {
		// defer文でパニックをキャッチ
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("\033[35m🚨 PANIC_RECOVERED in UpdateSpatialGrid: %v, Frame: %d\033[0m", r, g.frameCount)
				}
			}()
			g.UpdateSpatialGrid()
		}()
	}

	// 衝突判定
	for _, player := range g.Players {
		// defer文でパニックをキャッチ
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("\033[35m🚨 PANIC_RECOVERED in collision detection for player %s: %v, Frame: %d\033[0m", player.Name, r, g.frameCount)
				}
			}()

			if !player.Snake.Alive {
				return
			}

			// プレイヤー（人間）の当たり判定をスキップ
			if !player.IsNPC && utils.DISABLE_COLLISION {
				// 食べ物との衝突判定のみ実行
				head := player.Snake.Body[0]
				nearbyFood := g.spatialGrid.GetNearbyFoodSafe(head)

				for _, food := range nearbyFood {
					// 蛇の頭と食べ物の距離をチェック
					dx := head.X - food.Position.X
					dy := head.Y - food.Position.Y
					dist := dx*dx + dy*dy

					if dist < (utils.SNAKE_RADIUS+utils.FOOD_RADIUS)*(utils.SNAKE_RADIUS+utils.FOOD_RADIUS) {
						// 食べ物をポインタで直接削除
						g.RemoveFood(food)
						// 蛇を成長させる
						player.Snake.Growing = 3
						player.Score += 10
						return
					}
				}
				return
			}

			// NPCは通常の当たり判定
			// 自己衝突
			if player.Snake.CheckSelfCollision() {
				player.Snake.Alive = false
				player.Score -= 10
				if player.Score < 0 {
					player.Score = 0
				}
				return
			}

			// 他の蛇との衝突（空間分割で最適化、フォールバック付き）
			head := player.Snake.Body[0]
			nearbyPlayerIDs := g.spatialGrid.GetNearbyPlayersUnique(head)

			// 空間分割で候補が見つからない場合は全体検索（安全性確保）
			if len(nearbyPlayerIDs) == 0 {
				for otherPlayerID := range g.Players {
					if otherPlayerID != player.ID {
						nearbyPlayerIDs = append(nearbyPlayerIDs, otherPlayerID)
					}
				}
			}

			for _, otherPlayerID := range nearbyPlayerIDs {
				if otherPlayer, exists := g.Players[otherPlayerID]; exists &&
					player.ID != otherPlayer.ID &&
					player.Snake.CheckCollisionWith(otherPlayer.Snake) {
					player.Snake.Alive = false
					player.Score -= 10
					if player.Score < 0 {
						player.Score = 0
					}
					otherPlayer.Score += 5
					return
				}
			}

			// 食べ物との衝突判定（空間分割で最適化、安全）
			nearbyFood := g.spatialGrid.GetNearbyFoodSafe(head)

			for _, food := range nearbyFood {
				// 蛇の頭と食べ物の距離をチェック
				dx := head.X - food.Position.X
				dy := head.Y - food.Position.Y
				dist := dx*dx + dy*dy

				if dist < (utils.SNAKE_RADIUS+utils.FOOD_RADIUS)*(utils.SNAKE_RADIUS+utils.FOOD_RADIUS) {
					// 食べ物をポインタで直接削除
					g.RemoveFood(food)
					// 蛇を成長させる
					player.Snake.Growing = 3
					player.Score += 10
					return
				}
			}
		}()
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
	// Food を Position に変換
	foodPositions := make([]models.Position, len(g.Food))
	for i, food := range g.Food {
		foodPositions[i] = food.Position
	}

	return models.GameState{
		Players: players,
		Food:    foodPositions,
	}
}

// GetOptimizedState はクライアント専用の最適化されたゲーム状態を返す
func (g *Game) GetOptimizedState(clientPlayerID string, clientX, clientY, viewWidth, viewHeight float64) models.GameState {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// クライアントの画面範囲計算
	margin := utils.CULLING_MARGIN // 余裕を持った範囲
	minX := clientX - viewWidth/2 - margin
	maxX := clientX + viewWidth/2 + margin
	minY := clientY - viewHeight/2 - margin
	maxY := clientY + viewHeight/2 + margin

	players := make([]models.PlayerState, 0, 30) // 通常画面内は30体程度
	for _, p := range g.Players {
		// プレイヤーが画面範囲内にいるかチェック（生死問わず）
		if len(p.Snake.Body) > 0 {
			// 蛇の任意のセグメントが画面範囲内にあるかチェック
			isVisible := false
			for _, segment := range p.Snake.Body {
				if segment.X >= minX && segment.X <= maxX && segment.Y >= minY && segment.Y <= maxY {
					isVisible = true
					break
				}
			}

			if isVisible {
				// 元のデータを変更しないよう蛇のコピーを作成
				snakeCopy := *p.Snake

				// 体の一部でも画面内にあれば全身を送信（セグメント削除無し）

				players = append(players, models.PlayerState{
					ID:    p.ID,
					Name:  p.Name,
					Snake: &snakeCopy,
					Score: p.Score,
				})
			}
		}
	}

	// 画面範囲内の食べ物のみ
	food := make([]models.Position, 0, 50)
	for _, f := range g.Food {
		if f.Position.X >= minX && f.Position.X <= maxX && f.Position.Y >= minY && f.Position.Y <= maxY {
			food = append(food, f.Position)
		}
	}

	return models.GameState{
		Players: players,
		Food:    food,
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
		// WebSocket書き込みを同期化
		func() {
			player.ConnMu.Lock()
			defer player.ConnMu.Unlock()
			if err := player.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("\033[31m❌ WS_ERROR: Error broadcasting to player %s: %v\033[0m", player.ID, err)
			}
		}()
	}
}

// BroadcastOptimized は各クライアントに最適化されたデータを個別送信
func (g *Game) BroadcastOptimized() {
	// プレイヤーリストのスナップショットを取得
	g.mu.RLock()
	playerList := make([]*models.Player, 0, len(g.Players))
	for _, player := range g.Players {
		// NPCプレイヤーにはメッセージを送信しない
		if player.IsNPC || player.Conn == nil {
			continue
		}

		// プレイヤーの位置を取得（死んでいても送信を続ける）
		if len(player.Snake.Body) == 0 {
			continue
		}

		playerList = append(playerList, player)
	}
	g.mu.RUnlock()

	// スナップショットを使って各プレイヤーに送信（デッドロック回避）
	for _, player := range playerList {
		head := player.Snake.Body[0]
		// constants.goからカリング範囲を取得
		viewWidth := utils.CULLING_WIDTH
		viewHeight := utils.CULLING_HEIGHT

		// このクライアント専用の最適化されたゲーム状態を取得
		optimizedState := g.GetOptimizedState(player.ID, head.X, head.Y, viewWidth, viewHeight)

		message := map[string]interface{}{
			"type":  "gameState",
			"state": optimizedState,
		}

		data, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshaling optimized state for player %s: %v", player.ID, err)
			continue
		}

		// WebSocket書き込みを同期化
		func() {
			player.ConnMu.Lock()
			defer player.ConnMu.Unlock()
			if err := player.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("\033[31m❌ WS_ERROR: Error broadcasting optimized state to player %s: %v\033[0m", player.ID, err)
			}
			// デバッグ：データサイズをログ出力（10秒に1回）
			if g.frameCount%600 == 0 {
				log.Printf("\033[34m📊 WS_DATA: size for player %s: %d bytes\033[0m", player.Name, len(data))
			}
		}()
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
	// パニックリカバリー
	defer func() {
		if r := recover(); r != nil {
			log.Printf("\033[35m🚨 PANIC_RECOVERED in RunGameLoop for game %s: %v\033[0m", g.ID, r)
		}
	}()

	ticker := time.NewTicker(utils.GAME_TICK)
	defer ticker.Stop()
	lastUpdate := time.Now()

	log.Printf("\033[32m✅ GAME_LOOP: Started for game %s\033[0m", g.ID)
	defer log.Printf("\033[33m⚠️ GAME_LOOP: Ended for game %s\033[0m", g.ID)

	for g.Running {
		<-ticker.C
		now := time.Now()
		deltaTime := now.Sub(lastUpdate).Seconds()
		lastUpdate = now

		// 更新処理をゴルーチン安全にラップ
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("\033[35m🚨 PANIC_RECOVERED in game update for game %s: %v\033[0m", g.ID, r)
				}
			}()

			g.mu.Lock()
			// NPCの方向を更新
			g.updateNPCDirections()
			g.Update(deltaTime)
			g.mu.Unlock()

			// 各クライアントに最適化されたデータを個別送信
			g.BroadcastOptimized()
		}()
	}
}
