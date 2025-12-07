package game

import (
	"game-sandbox/server/models"
	"game-sandbox/server/utils"
	"time"
)

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

// destroyPlayer はプレイヤーを破壊し、コアと衛星を落とす
func (g *Game) destroyPlayer(player *models.Player) {
	// コアを落ちた衛星として追加（元コアとして記録）
	if player.Celestial.Core != nil {
		droppedCore := &models.DroppedSatellite{
			Position:       player.Celestial.Core.Position,
			Radius:         player.Celestial.Core.Radius,
			Color:          player.Celestial.Core.Color, // 元の色を維持
			IsOriginalCore: true,
		}
		g.DroppedSatellites = append(g.DroppedSatellites, droppedCore)
		g.spatialGrid.AddDroppedSatellite(droppedCore)
	}

	// 全ての衛星を落とす（元衛星として記録）
	satelliteCount := 0
	for _, orbit := range player.Celestial.Satellites {
		for _, sat := range orbit {
			droppedSat := &models.DroppedSatellite{
				Position:       sat.Sphere.Position,
				Radius:         sat.Sphere.Radius,
				Color:          "#FFFFFF", // 落ちた時は白色
				IsOriginalCore: false,
			}
			g.DroppedSatellites = append(g.DroppedSatellites, droppedSat)
			// spatial gridに追加
			g.spatialGrid.AddDroppedSatellite(droppedSat)
			satelliteCount++
		}
	}

	// プレイヤーを死亡状態にする
	player.Celestial.Alive = false
	player.Celestial.Satellites = [][]*models.Satellite{}

	utils.Info("Player destroyed", map[string]interface{}{
		"event":              "player_destroyed",
		"game_id":            g.ID,
		"player_id":          player.ID,
		"player_name":        player.Name,
		"is_npc":             player.IsNPC,
		"satellites_dropped": satelliteCount,
		"metric":             "game_event",
	})
}

// destroyTargetSatellite は指定した位置の衛星を完全消滅させる
func (g *Game) destroyTargetSatellite(player *models.Player, sphere *models.Sphere) {
	for oi, orbit := range player.Celestial.Satellites {
		for si, sat := range orbit {
			if sat.Sphere == sphere {
				// 衛星を完全消滅
				player.Celestial.RemoveSatellite(oi, si)

				// 衛星が減った場合は自動補充タイマーをリセット
				player.ResetAutoSatelliteTimerIfNeeded()

				return
			}
		}
	}
}

// removeDroppedSatellite は落ちた衛星を削除する
func (g *Game) removeDroppedSatellite(target *models.DroppedSatellite) {
	// spatial gridから削除
	g.spatialGrid.RemoveDroppedSatellite(target)

	// スライスから削除
	for i, droppedSat := range g.DroppedSatellites {
		if droppedSat == target {
			g.DroppedSatellites = append(
				g.DroppedSatellites[:i],
				g.DroppedSatellites[i+1:]...,
			)
			return
		}
	}
}

// updateAutoSatellites は各プレイヤーに定期的に衛星を自動追加する
func (g *Game) updateAutoSatellites() {
	for _, player := range g.Players {
		// 生きているプレイヤーのみ対象
		if !player.Celestial.Alive {
			continue
		}

		// 最後の自動追加から一定時間経過したかチェック
		if time.Since(player.LastAutoSatellite) < utils.AUTO_SATELLITE_INTERVAL {
			continue
		}

		// 現在の衛星数を取得
		currentSatelliteCount := player.Celestial.GetTotalSatelliteCount()

		// 上限に達していたら追加しない
		if currentSatelliteCount >= utils.MAX_AUTO_SATELLITES {
			continue
		}

		// コア位置
		corePos := player.Celestial.Core.Position

		startPos := models.Position{
			X: corePos.X,
			Y: corePos.Y,
		}

		// 自動追加の衛星はコアと同じ色
		player.Celestial.AddSatellite(player.Celestial.Core.Color, startPos)
		player.LastAutoSatellite = time.Now()

		utils.Debug("Auto satellite added", map[string]interface{}{
			"event":           "auto_satellite_added",
			"player_id":       player.ID,
			"player_name":     player.Name,
			"satellite_count": currentSatelliteCount + 1,
			"max_satellites":  utils.MAX_AUTO_SATELLITES,
		})
	}
}
