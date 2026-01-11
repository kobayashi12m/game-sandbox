package game

import (
	"encoding/json"
	"sort"

	"game-sandbox/server/models"
	"game-sandbox/server/utils"
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

// BroadcastScoreboard はスコアボード情報を全クライアントに送信（非同期）
func (g *Game) BroadcastScoreboard() {
	// クライアントのスナップショットを取得
	g.mu.RLock()
	clients := make([]*Client, 0, len(g.clients))
	for _, client := range g.clients {
		if !client.IsClosed() {
			clients = append(clients, client)
		}
	}
	g.mu.RUnlock()

	if len(clients) == 0 {
		return
	}

	// 上位10名のスコアボード情報を取得
	scoreboard := g.GetScoreboard()

	// 各クライアントに送信
	for _, client := range clients {
		player := client.Player
		if player == nil {
			continue
		}

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
			continue
		}

		// 非同期送信（ブロックしない）
		client.Send(data)
	}
}

// BroadcastOptimized は各クライアントに最適化されたデータを個別送信（非同期）
func (g *Game) BroadcastOptimized() {
	// クライアントのスナップショットを取得
	g.mu.RLock()
	clients := make([]*Client, 0, len(g.clients))
	for _, client := range g.clients {
		if client.IsClosed() {
			continue
		}
		if client.Player == nil || client.Player.Celestial.Core == nil {
			continue
		}
		clients = append(clients, client)
	}
	g.mu.RUnlock()

	// 各クライアントに送信
	for _, client := range clients {
		player := client.Player
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

		// 非同期送信
		client.Send(data)

		// 送信バイト数の追跡
		g.totalBytesSent += int64(len(data))
	}
}
