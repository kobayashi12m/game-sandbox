package game

import (
	"sync"
	"time"

	"game-sandbox/server/models"
	"game-sandbox/server/utils"

	"github.com/gorilla/websocket"
)

// Game はゲームセッションを表す
type Game struct {
	ID                string
	Players           map[string]*models.Player
	DroppedSatellites []*models.DroppedSatellite // 落ちた衛星
	Projectiles       []*models.Projectile       // 射出された衛星
	Running           bool
	NPCCount          int              // NPC数の設定
	spatialGrid       *SpatialGrid     // 空間分割グリッド
	frameCount        int64            // フレームカウンター
	humanPlayers      []*models.Player // WebSocket接続する人間プレイヤーのキャッシュ
	// 通信統計（シンプル版）
	totalBytesSent int64 // 送信バイト数の累計
	startTime      time.Time
	mu             sync.RWMutex
}

// AddPlayer はゲームに新しいプレイヤーを追加する
func (g *Game) AddPlayer(id, name string, conn *websocket.Conn) {
	g.mu.Lock()
	defer g.mu.Unlock()

	colors := []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#DDA0DD", "#F4A460"}
	color := colors[len(g.Players)%len(colors)]

	celestial := &models.Celestial{
		Color: color,
	}

	player := &models.Player{
		ID:        id,
		Name:      name,
		Celestial: celestial,
		Score:     0,
		Conn:      conn,
	}

	// WebSocket接続がない場合はNPCとして初期化
	if conn == nil {
		player.IsNPC = true
		player.LastDirectionChange = time.Now()
	}

	g.Players[id] = player

	// 初期スポーン処理（安全な位置でスポーン）
	g.SpawnPlayer(player)

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

	humanPlayers := 0
	for _, player := range g.Players {
		if !player.IsNPC {
			humanPlayers++
		}
	}

	// 人間プレイヤーが1人以上いて、ゲームが開始されていない場合
	if humanPlayers >= 1 && !g.Running {
		g.Running = true
		g.startTime = time.Now()
		go g.RunGameLoop()
		utils.LogGameSessionEvent("game_start", g.ID, humanPlayers, len(g.Players), 0)
		return true
	}
	return false
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
			utils.LogPanicRecovery("RunGameLoop", g.ID, r)
		}
	}()

	ticker := time.NewTicker(utils.GAME_TICK)
	defer ticker.Stop()
	lastUpdate := time.Now()

	utils.Info("Game loop started", map[string]interface{}{
		"game_id": g.ID,
		"event":   "game_loop_start",
	})
	defer utils.Info("Game loop ended", map[string]interface{}{
		"game_id": g.ID,
		"event":   "game_loop_end",
	})

	for g.Running {
		<-ticker.C
		now := time.Now()
		deltaTime := now.Sub(lastUpdate).Seconds()
		lastUpdate = now

		// 更新処理をゴルーチン安全にラップ
		func() {
			defer func() {
				if r := recover(); r != nil {
					utils.LogPanicRecovery("game_update", g.ID, r)
				}
			}()

			g.mu.Lock()
			// NPCの方向を更新
			// g.updateNPCDirections()
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

// GetStartTime returns the game start time
func (g *Game) GetStartTime() time.Time {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.startTime
}

// GetPlayers returns all players in the game
func (g *Game) GetPlayers() map[string]*models.Player {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Players
}

// EjectPlayerSatellite はプレイヤーの衛星を射出する
func (g *Game) EjectPlayerSatellite(player *models.Player, targetX, targetY float64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if player == nil || player.Celestial == nil || !player.Celestial.Alive {
		return
	}

	// 衛星を射出して、射出された衛星を取得
	ejectedSphere := player.Celestial.EjectSatelliteWithReturn(targetX, targetY)
	if ejectedSphere != nil {
		// 射出物として追加
		projectile := &models.Projectile{
			ID:       utils.GenerateID(),
			Sphere:   ejectedSphere,
			Owner:    player,
			Lifetime: 5.0, // 5秒間存在
		}
		g.Projectiles = append(g.Projectiles, projectile)

		// 衛星が減った場合は自動補充タイマーをリセット
		player.ResetAutoSatelliteTimerIfNeeded()

		utils.Debug("Satellite ejected", map[string]interface{}{
			"player_id":            player.ID,
			"player_name":          player.Name,
			"remaining_satellites": len(player.Celestial.Satellites),
			"event":                "satellite_eject",
		})
	}
}
