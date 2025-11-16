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
				g.checkOrganismCollision(player)
			}

			// 落ちた衛星との衝突判定
			core := player.Celestial.Core.Position
			collidedSatellite := g.checkDroppedSatelliteCollision(core)

			if collidedSatellite != nil {
				g.removeDroppedSatellite(collidedSatellite)
				player.Celestial.Growing = 3
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

	// 成長処理
	for _, player := range g.Players {
		if player.Celestial.Growing > 0 {
			player.Celestial.Growing--
			if player.Celestial.Growing == 0 {
				// 衛星を追加
				player.Celestial.AddSatellite()
			}
		}
	}

	// 射出物の更新
	g.updateProjectiles(deltaTime)

	// 射出物とプレイヤーの衝突判定
	g.checkProjectileCollisions()

	// 落ちた衛星の補充
	g.GenerateDroppedSatellites()
}

// checkOrganismCollision は球体レベルでの個別衝突判定を行う
func (g *Game) checkOrganismCollision(player *models.Player) {
	// プレイヤーの全球体（Core + Satellites）
	var playerSpheres []*models.Sphere
	playerSpheres = append(playerSpheres, player.Celestial.Core)
	playerSpheres = append(playerSpheres, player.Celestial.GetAllSpheres()...)

	// 各球体について衝突をチェック
	for _, sphere := range playerSpheres {
		collidedPlayer := g.spatialGrid.CheckCollisionAt(sphere.Position, player)
		if collidedPlayer != nil {
			// 衝突した相手プレイヤーの球体を特定
			targetSphere := g.findCollidedSphere(sphere, collidedPlayer)
			if targetSphere != nil {
				// 個別の球体間で衝突処理
				g.applySphereCollision(sphere, targetSphere)
			}
		}
	}
}

// findCollidedSphere は衝突している相手の球体を特定する
func (g *Game) findCollidedSphere(sphere *models.Sphere, targetPlayer *models.Player) *models.Sphere {
	// Core との衝突をチェック
	dx := sphere.Position.X - targetPlayer.Celestial.Core.Position.X
	dy := sphere.Position.Y - targetPlayer.Celestial.Core.Position.Y
	dist := dx*dx + dy*dy
	collisionDist := (sphere.Radius + targetPlayer.Celestial.Core.Radius) * (sphere.Radius + targetPlayer.Celestial.Core.Radius)

	if dist < collisionDist {
		return targetPlayer.Celestial.Core
	}

	// Satellites との衝突をチェック
	satellites := targetPlayer.Celestial.GetAllSpheres()
	for _, node := range satellites {
		dx = sphere.Position.X - node.Position.X
		dy = sphere.Position.Y - node.Position.Y
		dist = dx*dx + dy*dy
		collisionDist = (sphere.Radius + node.Radius) * (sphere.Radius + node.Radius)

		if dist < collisionDist {
			return node
		}
	}

	return nil
}

// applySphereCollision は個別の球体間の衝突を処理する
func (g *Game) applySphereCollision(sphere1, sphere2 *models.Sphere) {
	// 衝突方向ベクトルを計算
	dx := sphere1.Position.X - sphere2.Position.X
	dy := sphere1.Position.Y - sphere2.Position.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	// 最小衝突距離をチェック
	minDistance := sphere1.Radius + sphere2.Radius
	if distance > 0 && distance < minDistance {
		// 正規化された衝突方向
		nx := dx / distance
		ny := dy / distance

		// 相対速度を計算
		relVelX := sphere1.Velocity.X - sphere2.Velocity.X
		relVelY := sphere1.Velocity.Y - sphere2.Velocity.Y

		// 法線方向の相対速度
		relVelNormal := relVelX*nx + relVelY*ny

		// 接近している場合のみ衝突処理を適用
		if relVelNormal > 0 {
			// 衝突インパルスを計算
			impulse := -(1 + utils.COLLISION_RESTITUTION) * relVelNormal / (1/sphere1.Mass + 1/sphere2.Mass)

			// 各球体の速度変化を計算
			deltaV1X := impulse * nx / sphere1.Mass
			deltaV1Y := impulse * ny / sphere1.Mass
			deltaV2X := -impulse * nx / sphere2.Mass
			deltaV2Y := -impulse * ny / sphere2.Mass

			// 速度を更新
			sphere1.Velocity.X += deltaV1X
			sphere1.Velocity.Y += deltaV1Y
			sphere2.Velocity.X += deltaV2X
			sphere2.Velocity.Y += deltaV2Y

			// 位置分離（重なりを解消）
			overlap := minDistance - distance
			separationX := nx * overlap * 0.5
			separationY := ny * overlap * 0.5

			sphere1.Position.X += separationX
			sphere1.Position.Y += separationY
			sphere2.Position.X -= separationX
			sphere2.Position.Y -= separationY
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
		hit := false

		// 全プレイヤーとの衝突をチェック
		for _, player := range g.Players {
			// 自分の射出物はスキップ
			if player.ID == proj.OwnerID {
				continue
			}

			if !player.Celestial.Alive {
				continue
			}

			// コアとの衝突をチェック
			if g.checkSphereCollision(proj.Sphere, player.Celestial.Core) {
				// コアに当たった場合、プレイヤーを破壊（射出物も消滅）
				log.Printf("Projectile hit core: %s destroyed, projectile destroyed", player.Name)
				g.destroyPlayer(player)
				hit = true
				break
			}

			// 衛星との衝突をチェック
			var hitOrbitIndex, hitSatIndex int = -1, -1
		outerLoop:
			for oi, orbit := range player.Celestial.Satellites {
				for si, sat := range orbit {
					if g.checkSphereCollision(proj.Sphere, sat.Sphere) {
						hitOrbitIndex = oi
						hitSatIndex = si
						break outerLoop
					}
				}
			}

			if hitOrbitIndex >= 0 && hitSatIndex >= 0 {
				// 衛星に当たった場合、その衛星を完全消滅（射出物も消滅、落ちた衛星は作らない）
				log.Printf("Projectile hit satellite: %s satellite destroyed, projectile destroyed (both vanish)", player.Name)
				g.destroySatelliteCompletely(player, hitOrbitIndex, hitSatIndex)
				hit = true
				break
			}
		}

		// 当たらなかった射出物は残す
		if !hit {
			activeProjectiles = append(activeProjectiles, proj)
		}
	}

	g.Projectiles = activeProjectiles
}

// checkSphereCollision は二つの球体が衝突しているかチェックする
func (g *Game) checkSphereCollision(sphere1, sphere2 *models.Sphere) bool {
	dx := sphere1.Position.X - sphere2.Position.X
	dy := sphere1.Position.Y - sphere2.Position.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	minDist := sphere1.Radius + sphere2.Radius
	return dist < minDist
}

// destroyPlayer はプレイヤーを破壊し、衛星を落とす
func (g *Game) destroyPlayer(player *models.Player) {
	// 全ての衛星を落とす
	satelliteCount := 0
	for _, orbit := range player.Celestial.Satellites {
		for _, sat := range orbit {
			droppedSat := &models.DroppedSatellite{
				Position: sat.Sphere.Position,
				Radius:   sat.Sphere.Radius,
			}
			g.DroppedSatellites = append(g.DroppedSatellites, droppedSat)
			satelliteCount++
		}
	}

	// プレイヤーを死亡状態にする
	player.Celestial.Alive = false
	player.Celestial.Satellites = [][]*models.Satellite{}

	log.Printf("💥 Player %s core destroyed, %d satellites dropped at their locations", player.Name, satelliteCount)
}

// destroySatellite は指定された衛星を破壊する
func (g *Game) destroySatellite(player *models.Player, orbitIndex, satIndex int) {
	if orbitIndex < 0 || orbitIndex >= len(player.Celestial.Satellites) {
		return
	}
	if satIndex < 0 || satIndex >= len(player.Celestial.Satellites[orbitIndex]) {
		return
	}

	// 衛星を落とす
	sat := player.Celestial.Satellites[orbitIndex][satIndex]
	droppedSat := &models.DroppedSatellite{
		Position: sat.Sphere.Position,
		Radius:   sat.Sphere.Radius,
	}
	g.DroppedSatellites = append(g.DroppedSatellites, droppedSat)

	// 衛星を削除
	player.Celestial.RemoveSatellite(orbitIndex, satIndex)

	log.Printf("💫 Satellite destroyed for player %s, dropped satellite at position", player.Name)
}

// destroySatelliteCompletely は射出物衝突時に衛星を完全消滅させる（落ちた衛星は作らない）
func (g *Game) destroySatelliteCompletely(player *models.Player, orbitIndex, satIndex int) {
	if orbitIndex < 0 || orbitIndex >= len(player.Celestial.Satellites) {
		return
	}
	if satIndex < 0 || satIndex >= len(player.Celestial.Satellites[orbitIndex]) {
		return
	}

	// 衛星を削除（落ちた衛星は作らない）
	player.Celestial.RemoveSatellite(orbitIndex, satIndex)

	log.Printf("💥 Satellite completely destroyed for player %s (no dropped satellite)", player.Name)
}

// checkDroppedSatelliteCollision はコアと落ちた衛星の衝突をチェックする
func (g *Game) checkDroppedSatelliteCollision(corePos models.Position) *models.DroppedSatellite {
	for _, droppedSat := range g.DroppedSatellites {
		dx := corePos.X - droppedSat.Position.X
		dy := corePos.Y - droppedSat.Position.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		collisionDist := utils.SPHERE_RADIUS + droppedSat.Radius

		if dist < collisionDist {
			return droppedSat
		}
	}
	return nil
}

// removeDroppedSatellite は落ちた衛星を削除する
func (g *Game) removeDroppedSatellite(target *models.DroppedSatellite) {
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
			var positions []models.Position
			positions = append(positions, player.Celestial.Core.Position)
			for _, node := range player.Celestial.GetAllSpheres() {
				positions = append(positions, node.Position)
			}
			g.spatialGrid.AddPlayerSegments(player, positions)
		}
	}

	// 落ちた衛星はグリッドに追加しない（衝突判定が簡単なため）
}
