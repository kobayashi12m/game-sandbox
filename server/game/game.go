package game

import (
	"sync"
	"time"

	"game-sandbox/server/models"
	"game-sandbox/server/utils"

	"github.com/gorilla/websocket"
)

// GameCommand はゲームアクションの統一インターフェース
type GameCommand interface {
	Execute(*Game) error
	GetType() string
	GetPlayer() *models.Player
}

// AccelerationCommand はWebSocketからの加速度コマンド
type AccelerationCommand struct {
	Player *models.Player
	X, Y   float64
}

func (cmd AccelerationCommand) Execute(g *Game) error {
	if cmd.Player != nil && cmd.Player.Celestial != nil && cmd.Player.Celestial.Alive {
		cmd.Player.Celestial.SetAcceleration(cmd.X, cmd.Y)
	}
	return nil
}

func (cmd AccelerationCommand) GetType() string {
	return "acceleration"
}

func (cmd AccelerationCommand) GetPlayer() *models.Player {
	return cmd.Player
}

// ShootCommand は衛星射出コマンド
type ShootCommand struct {
	Player  *models.Player
	TargetX float64
	TargetY float64
}

func (cmd ShootCommand) Execute(g *Game) error {
	if cmd.Player != nil {
		g.EjectPlayerSatellite(cmd.Player, cmd.TargetX, cmd.TargetY)
	}
	return nil
}

func (cmd ShootCommand) GetType() string {
	return "shoot"
}

func (cmd ShootCommand) GetPlayer() *models.Player {
	return cmd.Player
}

// Game はゲームセッションを表す
type Game struct {
	ID                string
	Players           map[string]*models.Player
	clients           map[string]*Client         // WebSocket接続（ゲームロジックと分離）
	DroppedSatellites []*models.DroppedSatellite // 落ちた衛星
	Projectiles       []*models.Projectile       // 射出された衛星
	Running           bool
	spatialGrid       *SpatialGrid // 空間分割グリッド
	frameCount        int64        // フレームカウンター
	// 通信統計（シンプル版）
	totalBytesSent int64 // 送信バイト数の累計
	startTime      time.Time
	// コマンドキュー（デッドロック防止）
	commands chan GameCommand
	// コマンド統計
	commandsPerSecond      float64
	maxQueueLength         int
	commandsLast10Sec      int64 // 直近10秒間の処理数
	lastCommandLog         time.Time
	lastCommandStatUpdate  time.Time
	commandsSinceLastCheck int64
	mu                     sync.RWMutex
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
	}

	// WebSocket接続がない場合はNPCとして初期化
	if conn == nil {
		player.IsNPC = true
		player.LastDirectionChange = time.Now()
	}

	g.Players[id] = player

	// 人間プレイヤーの場合はClientを作成（Playerを渡す）
	if conn != nil {
		g.clients[id] = NewClient(conn, player)
	}

	// 初期スポーン処理（安全な位置でスポーン）
	g.SpawnPlayer(player)
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

	// Clientを閉じる（人間プレイヤーの場合）
	if client, ok := g.clients[id]; ok {
		client.Close()
		delete(g.clients, id)
	}

	// Playersマップから削除
	delete(g.Players, id)

	// 人間プレイヤーが減ったらNPCを補充
	if !player.IsNPC {
		go g.ReplenishNPCs()
	}
}

// GetPlayer はIDでプレイヤーを取得する
func (g *Game) GetPlayer(id string) (*models.Player, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	player, exists := g.Players[id]
	return player, exists
}

// GetClient はIDでClientを取得する
func (g *Game) GetClient(id string) (*Client, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	client, exists := g.clients[id]
	return client, exists
}

// humanPlayerCount はロック済み状態で人間プレイヤー数を返す（内部用）
func (g *Game) humanPlayerCount() int {
	return len(g.clients)
}

// GetHumanPlayerCount は人間プレイヤーの数を返す
func (g *Game) GetHumanPlayerCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.clients)
}

// ShouldStart はゲームを開始すべきかチェックし、必要なら開始する
func (g *Game) ShouldStart() bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	humanCount := g.humanPlayerCount()

	// 人間プレイヤーが1人以上いて、ゲームが開始されていない場合
	if humanCount >= 1 && !g.Running {
		g.Running = true
		g.startTime = time.Now()
		go g.RunGameLoop()
		utils.LogGameSessionEvent("game_start", g.ID, humanCount, len(g.Players), 0)
		return true
	}
	return false
}

// Stop はゲームを安全に停止する
func (g *Game) Stop() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Running = false

	// コマンドチャネルをクリーンアップ
	close(g.commands)
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

			// Updateメソッド内で必要に応じてロックを取得する
			g.Update(deltaTime)

			// 各クライアントに最適化されたデータを個別送信
			g.BroadcastOptimized()

			// スコアボードは1秒に1回送信（60フレーム = 60FPS * 1秒）
			if g.frameCount%60 == 0 {
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

// GetCommandChannel はGameCommandチャネルを返す（WebSocketハンドラー用）
func (g *Game) GetCommandChannel() chan GameCommand {
	return g.commands
}

// SendCommand はコマンドを安全に送信する
func (g *Game) SendCommand(cmd GameCommand) bool {
	select {
	case g.commands <- cmd:
		return true
	default:
		// キュー満杯の場合は失敗
		var playerID string
		if player := cmd.GetPlayer(); player != nil {
			playerID = player.ID
		}
		utils.Warn("Command queue full", map[string]interface{}{
			"command_type": cmd.GetType(),
			"player_id":    playerID,
		})
		return false
	}
}

// GetPlayers returns all players in the game
func (g *Game) GetPlayers() map[string]*models.Player {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Players
}
