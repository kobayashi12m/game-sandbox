package game

import (
	"log"
	"math"
	"time"

	"chess-mmo/server/models"
	"chess-mmo/server/utils"
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
			segments := len(player.Organism.Nodes) + 1 // コア + ノード
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

			if !player.Organism.Alive {
				deadPlayers++
			}
		}

		log.Printf("🎮 SERVER STATE: Frame %d | Players: %d (Human: %d, Dead: %d) | Food: %d | Segments: %d (Max: %d, Min: %d)",
			g.frameCount, len(g.Players), humanPlayers, deadPlayers, len(g.Food), totalSegments, maxOrganismLength, minOrganismLength)
	}

	// 全ての球体構造を移動
	for _, player := range g.Players {
		player.Organism.Move(deltaTime)
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

			if !player.Organism.Alive {
				return
			}

			// プレイヤー（人間）の当たり判定をスキップ
			if !player.IsNPC && utils.DISABLE_COLLISION {
				// 食べ物との衝突判定のみ実行
				core := player.Organism.Core.Position
				nearbyFood := g.spatialGrid.GetNearbyFoodSafe(core)

				for _, food := range nearbyFood {
					// コアと食べ物の距離をチェック
					dx := core.X - food.Position.X
					dy := core.Y - food.Position.Y
					dist := dx*dx + dy*dy

					if dist < (utils.ORGANISM_RADIUS+utils.FOOD_RADIUS)*(utils.ORGANISM_RADIUS+utils.FOOD_RADIUS) {
						// 食べ物をポインタで直接削除
						g.RemoveFood(food)
						// 球体構造を成長させる
						player.Organism.Growing = 3
						player.Score += 10
						return
					}
				}
				return
			}

			// 他の球体構造との衝突（組織レベルでの統合衝突判定）
			g.checkOrganismCollision(player)

			// 食べ物との衝突判定（空間分割で直接チェック）
			core := player.Organism.Core.Position
			collidedFood := g.spatialGrid.CheckFoodCollisionAt(core)

			if collidedFood != nil {
				// 食べ物をポインタで直接削除
				g.RemoveFood(collidedFood)
				// 球体構造を成長させる
				player.Organism.Growing = 3
				player.Score += 10
				return
			}
		}()
	}

	// 死んだ球体構造のリスポーン処理
	for _, player := range g.Players {
		if !player.Organism.Alive && !player.Organism.Respawning {
			player.Organism.Respawning = true
			player.Organism.DeathTime = time.Now()
		}

		if player.Organism.Respawning && time.Since(player.Organism.DeathTime) > 3*time.Second {
			player.Organism.Reset()
			player.Organism.Respawning = false
		}
	}

	// 食べ物の補充
	g.GenerateFood()
}

// checkOrganismCollision は球体レベルでの個別衝突判定を行う
func (g *Game) checkOrganismCollision(player *models.Player) {
	// プレイヤーの全球体（Core + Nodes）
	var playerSpheres []*models.PhysicsNode
	playerSpheres = append(playerSpheres, player.Organism.Core)
	for i := range player.Organism.Nodes {
		playerSpheres = append(playerSpheres, player.Organism.Nodes[i])
	}

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
func (g *Game) findCollidedSphere(sphere *models.PhysicsNode, targetPlayer *models.Player) *models.PhysicsNode {
	// Core との衝突をチェック
	dx := sphere.Position.X - targetPlayer.Organism.Core.Position.X
	dy := sphere.Position.Y - targetPlayer.Organism.Core.Position.Y
	dist := dx*dx + dy*dy
	collisionDist := (sphere.Radius + targetPlayer.Organism.Core.Radius) * (sphere.Radius + targetPlayer.Organism.Core.Radius)
	
	if dist < collisionDist {
		return targetPlayer.Organism.Core
	}

	// Nodes との衝突をチェック
	for i := range targetPlayer.Organism.Nodes {
		node := targetPlayer.Organism.Nodes[i]
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
func (g *Game) applySphereCollision(sphere1, sphere2 *models.PhysicsNode) {
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

// applyOrganismCollisionRepulsion は組織間の物理的反発を処理する（物理法則に基づく）
func (g *Game) applyOrganismCollisionRepulsion(player1, player2 *models.Player) {
	// 両組織のコア間の方向を基準にする
	core1 := player1.Organism.Core.Position
	core2 := player2.Organism.Core.Position

	// 衝突方向ベクトルを計算
	dx := core1.X - core2.X
	dy := core1.Y - core2.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	// 最小衝突距離をチェック
	minDistance := utils.ORGANISM_RADIUS * utils.COLLISION_MIN_DISTANCE
	if distance > 0 && distance < minDistance {
		// 正規化された衝突方向
		nx := dx / distance
		ny := dy / distance

		// 相対速度を計算
		relVelX := player1.Organism.Core.Velocity.X - player2.Organism.Core.Velocity.X
		relVelY := player1.Organism.Core.Velocity.Y - player2.Organism.Core.Velocity.Y

		// 法線方向の相対速度
		relVelNormal := relVelX*nx + relVelY*ny

		// 接近している場合のみ衝突処理を適用
		if relVelNormal > 0 {
			// 質量（簡単のため同じ質量とする）
			mass1 := player1.Organism.Core.Mass
			mass2 := player2.Organism.Core.Mass

			// 衝突インパルスを計算（物理法則に基づく）
			impulse := -(1 + utils.COLLISION_RESTITUTION) * relVelNormal / (1/mass1 + 1/mass2)

			// 各プレイヤーの速度変化を計算
			deltaV1X := impulse * nx / mass1
			deltaV1Y := impulse * ny / mass1
			deltaV2X := -impulse * nx / mass2
			deltaV2Y := -impulse * ny / mass2

			// 速度を更新（加算ではなく物理的な速度変化）
			player1.Organism.Core.Velocity.X += deltaV1X
			player1.Organism.Core.Velocity.Y += deltaV1Y
			player2.Organism.Core.Velocity.X += deltaV2X
			player2.Organism.Core.Velocity.Y += deltaV2Y

			// 位置分離（重なりを解消）
			overlap := minDistance - distance
			separationX := nx * overlap * 0.5
			separationY := ny * overlap * 0.5

			player1.Organism.Core.Position.X += separationX
			player1.Organism.Core.Position.Y += separationY
			player2.Organism.Core.Position.X -= separationX
			player2.Organism.Core.Position.Y -= separationY
		}
	}
}

// UpdateSpatialGrid は空間分割グリッドを更新する
func (g *Game) UpdateSpatialGrid() {
	// グリッドをクリア
	g.spatialGrid.Clear()

	// プレイヤーの全球体をグリッドに追加
	for _, player := range g.Players {
		if player.Organism.Core != nil {
			// 球体構造の全ノードをグリッドに登録
			var positions []models.Position
			positions = append(positions, player.Organism.Core.Position)
			for _, node := range player.Organism.Nodes {
				positions = append(positions, node.Position)
			}
			g.spatialGrid.AddPlayerSegments(player, positions)
		}
	}

	// 食べ物をグリッドに追加
	for _, food := range g.Food {
		g.spatialGrid.AddFood(food)
	}
}
