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

// checkOrganismCollision は組織レベルでの統合衝突判定を行う
func (g *Game) checkOrganismCollision(player *models.Player) {
	// コア＋全ノードから衝突した相手プレイヤーを収集
	collidedPlayers := make(map[string]*models.Player)
	
	// コアの衝突チェック
	core := player.Organism.Core.Position
	collidedPlayer := g.spatialGrid.CheckCollisionAt(core, player)
	if collidedPlayer != nil {
		collidedPlayers[collidedPlayer.ID] = collidedPlayer
	}

	// 各ノードの衝突チェック
	for _, node := range player.Organism.Nodes {
		collidedPlayerFromNode := g.spatialGrid.CheckCollisionAt(node.Position, player)
		if collidedPlayerFromNode != nil {
			collidedPlayers[collidedPlayerFromNode.ID] = collidedPlayerFromNode
		}
	}

	// 衝突した各プレイヤーに対して1回だけ反発力を適用
	for _, collidedPlayer := range collidedPlayers {
		g.applyOrganismCollisionRepulsion(player, collidedPlayer)
	}
}

// applyOrganismCollisionRepulsion は組織間の物理的反発を処理する（1回のみ）
func (g *Game) applyOrganismCollisionRepulsion(player1, player2 *models.Player) {
	// 両組織のコア間の方向を基準にする
	core1 := player1.Organism.Core.Position
	core2 := player2.Organism.Core.Position

	// 衝突方向ベクトルを計算
	dx := core1.X - core2.X
	dy := core1.Y - core2.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance > 0 {
		// 正規化
		nx := dx / distance
		ny := dy / distance

		// 反発力の強さ（距離に応じて調整）
		minDistance := utils.ORGANISM_RADIUS * 2.0
		repulsionForce := 30.0 // 以前より弱く設定
		
		// 距離が近いほど強い反発
		if distance < minDistance {
			repulsionForce *= (minDistance - distance) / minDistance
			repulsionForce = math.Min(repulsionForce, 80.0) // 最大値制限
		}

		// プレイヤー1のコアに反発力を適用
		player1.Organism.Core.Velocity.X += nx * repulsionForce
		player1.Organism.Core.Velocity.Y += ny * repulsionForce

		// プレイヤー2のコアに逆方向の反発力を適用
		player2.Organism.Core.Velocity.X -= nx * repulsionForce
		player2.Organism.Core.Velocity.Y -= ny * repulsionForce
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
