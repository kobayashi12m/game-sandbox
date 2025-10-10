package game

import (
	"log"
	"sync"
	"time"

	"chess-mmo/server/models"
	"chess-mmo/server/utils"

	"github.com/gorilla/websocket"
)

// Game はゲームセッションを表す
type Game struct {
	ID           string
	Players      map[string]*models.Player
	Food         []*models.Food // 食べ物をポインタで管理
	Running      bool
	NPCCount     int              // NPC数の設定
	spatialGrid  *SpatialGrid     // 空間分割グリッド
	frameCount   int64            // フレームカウンター
	humanPlayers []*models.Player // WebSocket接続する人間プレイヤーのキャッシュ
	mu           sync.RWMutex
}

// AddPlayer はゲームに新しいプレイヤーを追加する
func (g *Game) AddPlayer(id, name string, conn *websocket.Conn) {
	g.mu.Lock()
	defer g.mu.Unlock()

	colors := []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#DDA0DD", "#F4A460"}
	color := colors[len(g.Players)%len(colors)]

	organism := &models.OrganismBody{
		Color: color,
	}
	organism.Reset()

	player := &models.Player{
		ID:       id,
		Name:     name,
		Organism: organism,
		Score:    0,
		Conn:     conn,
	}

	g.Players[id] = player

	// 人間プレイヤーの場合はキャッシュに追加
	if !player.IsNPC && player.Conn != nil {
		g.humanPlayers = append(g.humanPlayers, player)
	}
}

// RemovePlayer はゲームからプレイヤーを削除する
func (g *Game) RemovePlayer(id string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// プレイヤーを取得
	player, exists := g.Players[id]
	if !exists {
		return
	}

	// Playersマップから削除
	delete(g.Players, id)

	// 人間プレイヤーキャッシュからも削除
	if !player.IsNPC {
		for i, cachedPlayer := range g.humanPlayers {
			if cachedPlayer == player {
				g.humanPlayers = append(g.humanPlayers[:i], g.humanPlayers[i+1:]...)
				break
			}
		}
	}
}

// GetPlayer はIDでプレイヤーを取得する
func (g *Game) GetPlayer(id string) (*models.Player, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	player, exists := g.Players[id]
	return player, exists
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

	if player, exists := g.Players[playerID]; exists && player.Organism.Alive {
		player.Organism.ChangeDirection(direction)
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

			// スコアボードは3秒に1回送信（180フレーム = 60FPS * 3秒）
			if g.frameCount%180 == 0 {
				g.BroadcastScoreboard()
			}
		}()
	}
}

// GetSpatialGridLines はSpatialGridの分割線を取得する
func (g *Game) GetSpatialGridLines() []models.GridLine {
	return g.spatialGrid.GetGridLines()
}
