package game

import (
	"encoding/json"
	"sort"
	"time"

	"game-sandbox/server/models"
	"game-sandbox/server/utils"

	"github.com/gorilla/websocket"
)

// GetState はクライアント用の現在のゲーム状態を返す
func (g *Game) GetState() models.GameState {
	g.mu.RLock()
	defer g.mu.RUnlock()

	players := make([]models.PlayerState, 0, len(g.Players))
	for _, p := range g.Players {
		// 元のデータを変更しないよう球体構造のコピーを作成
		celestialCopy := *p.Celestial

		players = append(players, models.PlayerState{
			ID:           p.ID,
			Name:         p.Name,
			Celestial:    &celestialCopy,
			Score:        p.Score,
			Invulnerable: p.IsInvulnerable(),
		})
	}

	// 射出物をコピー
	projectiles := make([]models.Projectile, len(g.Projectiles))
	for i, proj := range g.Projectiles {
		projectiles[i] = *proj
	}

	// 落ちた衛星をコピー
	droppedSatellites := make([]models.DroppedSatellite, len(g.DroppedSatellites))
	for i, sat := range g.DroppedSatellites {
		droppedSatellites[i] = *sat
	}

	// NPCデバッグ情報を追加
	npcDebug := g.createNPCDebugInfo()

	return models.GameState{
		Players:           players,
		DroppedSatellites: droppedSatellites,
		Projectiles:       projectiles,
		NPCDebug:          npcDebug,
	}
}

// GetOptimizedState はクライアント専用の最適化されたゲーム状態を返す
func (g *Game) GetOptimizedState(clientPlayerID string, clientX, clientY, viewWidth, viewHeight float64) models.GameState {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// クライアントの画面範囲計算
	minX := clientX - viewWidth/2
	maxX := clientX + viewWidth/2
	minY := clientY - viewHeight/2
	maxY := clientY + viewHeight/2

	// Spatial Gridで画面範囲内のプレイヤーと落ちた衛星を同時に取得
	areaResult := g.spatialGrid.GetObjectsInArea(minX, maxX, minY, maxY)

	players := make([]models.PlayerState, 0, len(areaResult.Players))
	for _, p := range areaResult.Players {
		if p.Celestial.Core != nil {
			// 元のデータを変更しないよう球体構造の深いコピーを作成
			celestialCopy := *p.Celestial
			coreCopy := *p.Celestial.Core
			celestialCopy.Core = &coreCopy

			// 自分以外のプレイヤーの速度・加速度情報をクリア
			if p.ID != clientPlayerID {
				celestialCopy.Core.Velocity = models.Position{}
				celestialCopy.Core.Acceleration = models.Position{}
			}

			players = append(players, models.PlayerState{
				ID:           p.ID,
				Name:         p.Name,
				Celestial:    &celestialCopy,
				Score:        p.Score,
				Invulnerable: p.IsInvulnerable(),
			})
		}
	}

	// 画面範囲内の射出物を取得
	projectiles := make([]models.Projectile, 0)
	for _, proj := range g.Projectiles {
		if proj.Sphere.Position.X >= minX && proj.Sphere.Position.X <= maxX &&
			proj.Sphere.Position.Y >= minY && proj.Sphere.Position.Y <= maxY {
			projectiles = append(projectiles, *proj)
		}
	}

	// spatial gridからの落ちた衛星をスライスに変換
	droppedSatellites := make([]models.DroppedSatellite, len(areaResult.DroppedSatellites))
	for i, sat := range areaResult.DroppedSatellites {
		droppedSatellites[i] = *sat
	}

	// NPCデバッグ情報を追加
	npcDebug := g.createNPCDebugInfo()

	return models.GameState{
		Players:           players,
		DroppedSatellites: droppedSatellites,
		Projectiles:       projectiles,
		NPCDebug:          npcDebug,
	}
}

// GetScoreboard は全プレイヤーのスコア情報をソート済みで返す
func (g *Game) GetScoreboard() []models.ScoreInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()

	scores := make([]models.ScoreInfo, 0, len(g.Players))
	for _, p := range g.Players {
		scores = append(scores, models.ScoreInfo{
			ID:    p.ID,
			Name:  p.Name,
			Score: p.Score,
			Alive: p.Celestial.Alive,
			Color: p.Celestial.Color,
		})
	}

	// サーバー側でスコア順にソート（高い順）
	// 同スコアの場合はIDでソート
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Score != scores[j].Score {
			return scores[i].Score > scores[j].Score // スコア高い順
		}
		return scores[i].ID < scores[j].ID // 同スコアならID昇順
	})

	// 上位10名に制限
	if len(scores) > 10 {
		scores = scores[:10]
	}

	return scores
}

// BroadcastScoreboard はスコアボード情報を全クライアントに送信
func (g *Game) BroadcastScoreboard() {
	g.mu.RLock()
	playerList := make([]*models.Player, 0)
	for _, player := range g.Players {
		if !player.IsNPC && player.Conn != nil {
			playerList = append(playerList, player)
		}
	}
	g.mu.RUnlock()

	if len(playerList) == 0 {
		return
	}

	// 上位10名のスコアボード情報を取得
	scoreboard := g.GetScoreboard()

	// 各プレイヤーに個別に送信（自分のスコア情報も含める）
	for _, player := range playerList {
		func() {
			// プレイヤー自身のスコア情報を作成
			myScore := models.ScoreInfo{
				ID:    player.ID,
				Name:  player.Name,
				Score: player.Score,
				Alive: player.Celestial.Alive,
				Color: player.Celestial.Color,
			}

			message := map[string]interface{}{
				"type":       "scoreboard",
				"scoreboard": scoreboard,
				"myScore":    myScore,
			}

			data, err := json.Marshal(message)
			if err != nil {
				utils.Error("Failed to marshal scoreboard", map[string]interface{}{
					"error":     err.Error(),
					"player_id": player.ID,
				})
				return
			}

			player.ConnMu.Lock()
			defer player.ConnMu.Unlock()
			if err := player.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				utils.LogWebSocketError(player.ID, "broadcast_scoreboard", err)
			}
		}()
	}
}

// BroadcastOptimized は各クライアントに最適化されたデータを個別送信
func (g *Game) BroadcastOptimized() {
	// 人間プレイヤーリストを取得
	g.mu.RLock()
	playerList := make([]*models.Player, 0)
	for _, player := range g.Players {
		if player.IsNPC {
			continue
		}
		// 接続が切断されていないかチェック
		if player.Conn == nil {
			continue
		}

		// プレイヤーの位置を取得（死んでいても送信を続ける）
		if player.Celestial.Core == nil {
			continue
		}

		playerList = append(playerList, player)
	}
	g.mu.RUnlock()

	// スナップショットを使って各プレイヤーに送信（デッドロック回避）
	for _, player := range playerList {
		core := player.Celestial.Core.Position
		// ズームレベルに応じてカリング範囲を調整
		zoomScale := utils.CAMERA_ZOOM_SCALE
		viewWidth := utils.CULLING_WIDTH / zoomScale
		viewHeight := utils.CULLING_HEIGHT / zoomScale

		// このクライアント専用の最適化されたゲーム状態を取得
		optimizedState := g.GetOptimizedState(player.ID, core.X, core.Y, viewWidth, viewHeight)

		message := map[string]interface{}{
			"type":  "gameState",
			"state": optimizedState,
		}

		data, err := json.Marshal(message)
		if err != nil {
			utils.Error("Failed to marshal optimized state", map[string]interface{}{
				"error":     err.Error(),
				"player_id": player.ID,
				"game_id":   g.ID,
			})
			continue
		}

		// WebSocket書き込みを同期化
		func() {
			player.ConnMu.Lock()
			defer player.ConnMu.Unlock()
			start := time.Now()
			if err := player.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				utils.LogWebSocketError(player.ID, "broadcast_state", err)
			} else {
				// パフォーマンス計測
				duration := time.Since(start)
				utils.LogPerformanceWarning("websocket_write", duration, 10*time.Millisecond)

				// 送信バイト数の追跡
				g.totalBytesSent += int64(len(data))
			}
		}()
	}
}
