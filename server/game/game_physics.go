package game

import (
	"log"
	"math"
	"time"

	"game-sandbox/server/models"
	"game-sandbox/server/utils"
)

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
		maxOrganismLength := 0
		minOrganismLength := 999999
		deadPlayers := 0

		for _, player := range g.Players {
			segments := player.Celestial.GetTotalSatelliteCount() + 1 // コア + ノード
			totalSegments += segments

			if !player.IsNPC {
				humanPlayers++
			}

			if segments > maxOrganismLength {
				maxOrganismLength = segments
			}
			if segments < minOrganismLength {
				minOrganismLength = segments
			}

			if !player.Celestial.Alive {
				deadPlayers++
			}
		}

		log.Printf("🎮 SERVER STATE: Frame %d | Players: %d (Human: %d, Dead: %d) | Dropped Satellites: %d | Segments: %d (Max: %d, Min: %d)",
			g.frameCount, len(g.Players), humanPlayers, deadPlayers, len(g.DroppedSatellites), totalSegments, maxOrganismLength, minOrganismLength)
		log.Printf("📡 TEST: Network monitoring test")
	}

	// 全ての天体システムの運動を更新
	for _, player := range g.Players {
		player.Celestial.UpdateMotion(deltaTime)
	}

	// 空間分割グリッドを毎フレーム更新
	// defer文でパニックをキャッチ
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("\033[35m🚨 PANIC_RECOVERED in UpdateSpatialGrid: %v, Frame: %d\033[0m", r, g.frameCount)
			}
		}()
		g.UpdateSpatialGrid()
	}()

	// 衝突判定
	for _, player := range g.Players {
		// defer文でパニックをキャッチ
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("\033[35m🚨 PANIC_RECOVERED in collision detection for player %s: %v, Frame: %d\033[0m", player.Name, r, g.frameCount)
				}
			}()

			if !player.Celestial.Alive {
				return
			}

			// プレイヤーとの衝突判定をスキップするかどうか
			skipPlayerCollision := !player.IsNPC && utils.DISABLE_COLLISION

			// 他プレイヤーとの衝突判定
			if !skipPlayerCollision {
				g.checkCelestialCollision(player)
			}

			// 落ちた衛星との衝突判定（0-3層が全て満杯の場合は衝突判定しない）
			var collidedSatellite *models.DroppedSatellite
			if !player.Celestial.AreAllOrbitsFullUpToLayer(3) {
				collidedSatellite = g.spatialGrid.CheckDroppedSatelliteCollision(player)
			}

			if collidedSatellite != nil {
				// 拾った衛星の位置から新しい衛星を追加
				var satelliteColor string
				if collidedSatellite.IsOriginalCore {
					// 元コアの場合は元の色を維持
					satelliteColor = collidedSatellite.Color
				} else {
					// 元衛星の場合は拾ったプレイヤーのコア色
					satelliteColor = player.Celestial.Core.Color
				}
				player.Celestial.AddSatellite(satelliteColor, collidedSatellite.Position)
				g.removeDroppedSatellite(collidedSatellite)
				player.Score += 10
			}
		}()
	}

	// 死んだ球体構造のリスポーン処理
	for _, player := range g.Players {
		if !player.Celestial.Alive && !player.Celestial.Respawning {
			player.Celestial.Respawning = true
			player.Celestial.DeathTime = time.Now()
		}

		if player.Celestial.Respawning && time.Since(player.Celestial.DeathTime) > 3*time.Second {
			player.Celestial.Reset()
			player.Celestial.Respawning = false
		}
	}

	// 射出物の更新
	g.updateProjectiles(deltaTime)

	// 射出物とプレイヤーの衝突判定
	g.checkProjectileCollisions()

	// 落ちた衛星の補充
	g.GenerateDroppedSatellites()

	// 自動衛星追加
	g.updateAutoSatellites()
}

// checkCelestialCollision は球体レベルでの個別衝突判定を行う
func (g *Game) checkCelestialCollision(player *models.Player) {
	// プレイヤーの全球体（Core + Satellites）
	var playerSpheres []*models.Sphere
	playerSpheres = append(playerSpheres, player.Celestial.Core)
	playerSpheres = append(playerSpheres, player.Celestial.GetAllSpheres()...)

	// 各球体について衝突をチェック
	for _, sphere := range playerSpheres {
		hitPlayer, hitSphere := g.spatialGrid.CheckCollisionAt(sphere.Position, sphere.Radius, player)
		if hitSphere != nil {
			g.applySphereCollision(sphere, hitSphere, player, hitPlayer)
		}
	}
}

// applySphereCollision は個別の球体間の衝突を処理する
func (g *Game) applySphereCollision(sphere1, sphere2 *models.Sphere, player1, player2 *models.Player) {
	// 衝突方向ベクトルを計算
	dx := sphere1.Position.X - sphere2.Position.X
	dy := sphere1.Position.Y - sphere2.Position.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	// 最小衝突距離をチェック
	minDistance := sphere1.Radius + sphere2.Radius
	if distance > 0 && distance < minDistance {
		// 球体の種類を判定
		isCore1 := player1.Celestial.IsCore(sphere1)
		isCore2 := player2.Celestial.IsCore(sphere2)

		// 衝突ルール
		if isCore1 && isCore2 {
			// コア同士：何も起きない
			return
		} else if !isCore1 && !isCore2 {
			// 衛星同士：両方消滅
			log.Printf("💥 Satellite collision: both satellites destroyed")
			g.destroyTargetSatellite(player1, sphere1)
			g.destroyTargetSatellite(player2, sphere2)
		} else {
			// コアと衛星：両方消滅
			if isCore1 {
				log.Printf("💥 Core-Satellite collision: %s core destroyed, satellite destroyed", player1.Name)
				g.destroyPlayer(player1)
				g.destroyTargetSatellite(player2, sphere2)
			} else {
				log.Printf("💥 Core-Satellite collision: %s core destroyed, satellite destroyed", player2.Name)
				g.destroyPlayer(player2)
				g.destroyTargetSatellite(player1, sphere1)
			}
		}
	}
}

// updateProjectiles は射出物の更新とライフサイクル管理を行う
func (g *Game) updateProjectiles(deltaTime float64) {
	var activeProjectiles []*models.Projectile

	for _, proj := range g.Projectiles {
		// 寿命を減らす
		proj.Lifetime -= deltaTime

		// 寿命が残っている場合のみ保持
		if proj.Lifetime > 0 {
			// 位置を更新
			proj.Sphere.Position.X += proj.Sphere.Velocity.X * deltaTime
			proj.Sphere.Position.Y += proj.Sphere.Velocity.Y * deltaTime

			// フィールド境界チェック
			if proj.Sphere.Position.X < 0 || proj.Sphere.Position.X > utils.FIELD_WIDTH ||
				proj.Sphere.Position.Y < 0 || proj.Sphere.Position.Y > utils.FIELD_HEIGHT {
				continue // 境界外の射出物は削除
			}

			activeProjectiles = append(activeProjectiles, proj)
		}
	}

	g.Projectiles = activeProjectiles
}

// checkProjectileCollisions は射出物とプレイヤーの衝突判定を行う
func (g *Game) checkProjectileCollisions() {
	var activeProjectiles []*models.Projectile

	for _, proj := range g.Projectiles {
		// 衝突検出
		hitPlayer, hitSphere := g.spatialGrid.CheckCollisionAt(proj.Sphere.Position, proj.Sphere.Radius, proj.Owner)
		if hitSphere != nil && hitPlayer != nil {
			// 衝突時の処理
			if hitSphere == hitPlayer.Celestial.Core {
				log.Printf("Projectile hit core: %s destroyed", hitPlayer.Name)
				g.destroyPlayer(hitPlayer)
			} else {
				log.Printf("Projectile hit satellite: %s satellite destroyed", hitPlayer.Name)
				g.destroyTargetSatellite(hitPlayer, hitSphere)
			}
			continue // 射出物は消滅
		}

		// 当たらなかった射出物は残す
		activeProjectiles = append(activeProjectiles, proj)
	}

	g.Projectiles = activeProjectiles
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

	log.Printf("💥 Player %s core destroyed, core + %d satellites dropped at their locations", player.Name, satelliteCount)
}

// destroyTargetSatellite は指定した位置の衛星を完全消滅させる
func (g *Game) destroyTargetSatellite(player *models.Player, sphere *models.Sphere) {
	for oi, orbit := range player.Celestial.Satellites {
		for si, sat := range orbit {
			if sat.Sphere == sphere {
				// 衛星を完全消滅
				player.Celestial.RemoveSatellite(oi, si)

				// 衛星が減った場合は自動補充タイマーをリセット
				g.resetAutoSatelliteTimerIfNeeded(player)

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

// UpdateSpatialGrid は空間分割グリッドを更新する
func (g *Game) UpdateSpatialGrid() {
	// グリッドをクリア
	g.spatialGrid.Clear()

	// プレイヤーの全球体をグリッドに追加
	for _, player := range g.Players {
		if player.Celestial.Core != nil {
			// 球体構造の全ノードをグリッドに登録
			var spheres []*models.Sphere
			spheres = append(spheres, player.Celestial.Core)
			spheres = append(spheres, player.Celestial.GetAllSpheres()...)
			g.spatialGrid.AddPlayerSpheres(player, spheres)
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

		log.Printf("🌟 Auto satellite added to %s (total: %d/%d)",
			player.Name, currentSatelliteCount+1, utils.MAX_AUTO_SATELLITES)
	}
}

// resetAutoSatelliteTimerIfNeeded は衛星数が上限未満になった場合にタイマーをリセットする
func (g *Game) resetAutoSatelliteTimerIfNeeded(player *models.Player) {
	if !player.Celestial.Alive {
		return
	}

	currentSatelliteCount := player.Celestial.GetTotalSatelliteCount()
	if currentSatelliteCount < utils.MAX_AUTO_SATELLITES {
		player.LastAutoSatellite = time.Now()
		log.Printf("⏰ Auto satellite timer reset for %s (satellites: %d/%d)",
			player.Name, currentSatelliteCount, utils.MAX_AUTO_SATELLITES)
	}
}
