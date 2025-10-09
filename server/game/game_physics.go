package game

import (
	"log"
	"time"

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
		maxSnakeLength := 0
		minSnakeLength := 999999
		deadPlayers := 0

		for _, player := range g.Players {
			segments := len(player.Snake.Body)
			totalSegments += segments

			if !player.IsNPC {
				humanPlayers++
			}

			if segments > maxSnakeLength {
				maxSnakeLength = segments
			}
			if segments < minSnakeLength {
				minSnakeLength = segments
			}

			if !player.Snake.Alive {
				deadPlayers++
			}
		}

		log.Printf("🎮 SERVER STATE: Frame %d | Players: %d (Human: %d, Dead: %d) | Food: %d | Segments: %d (Max: %d, Min: %d)",
			g.frameCount, len(g.Players), humanPlayers, deadPlayers, len(g.Food), totalSegments, maxSnakeLength, minSnakeLength)
	}

	// 全ての蛇を移動
	for _, player := range g.Players {
		player.Snake.Move(deltaTime)
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

			if !player.Snake.Alive {
				return
			}

			// プレイヤー（人間）の当たり判定をスキップ
			if !player.IsNPC && utils.DISABLE_COLLISION {
				// 食べ物との衝突判定のみ実行
				head := player.Snake.Body[0]
				nearbyFood := g.spatialGrid.GetNearbyFoodSafe(head)

				for _, food := range nearbyFood {
					// 蛇の頭と食べ物の距離をチェック
					dx := head.X - food.Position.X
					dy := head.Y - food.Position.Y
					dist := dx*dx + dy*dy

					if dist < (utils.SNAKE_RADIUS+utils.FOOD_RADIUS)*(utils.SNAKE_RADIUS+utils.FOOD_RADIUS) {
						// 食べ物をポインタで直接削除
						g.RemoveFood(food)
						// 蛇を成長させる
						player.Snake.Growing = 3
						player.Score += 10
						return
					}
				}
				return
			}

			// 他の蛇との衝突（空間分割で最適化、セグメント直接チェック）
			head := player.Snake.Body[0]
			collidedPlayer := g.spatialGrid.CheckCollisionAt(head, player)

			if collidedPlayer != nil {
				player.Snake.Alive = false
				player.Score -= 10
				if player.Score < 0 {
					player.Score = 0
				}
				collidedPlayer.Score += 5
				return
			}

			// 食べ物との衝突判定（空間分割で直接チェック）
			collidedFood := g.spatialGrid.CheckFoodCollisionAt(head)

			if collidedFood != nil {
				// 食べ物をポインタで直接削除
				g.RemoveFood(collidedFood)
				// 蛇を成長させる
				player.Snake.Growing = 3
				player.Score += 10
				return
			}
		}()
	}

	// 死んだ蛇のリスポーン処理
	for _, player := range g.Players {
		if !player.Snake.Alive && !player.Snake.Respawning {
			player.Snake.Respawning = true
			player.Snake.DeathTime = time.Now()
		}

		if player.Snake.Respawning && time.Since(player.Snake.DeathTime) > 3*time.Second {
			player.Snake.Reset()
			player.Snake.Respawning = false
		}
	}

	// 食べ物の補充
	g.GenerateFood()
}

// UpdateSpatialGrid は空間分割グリッドを更新する
func (g *Game) UpdateSpatialGrid() {
	// グリッドをクリア
	g.spatialGrid.Clear()

	// プレイヤーの全セグメントをグリッドに追加
	for _, player := range g.Players {
		if len(player.Snake.Body) > 0 {
			// 蛇の全セグメントをグリッドに登録
			g.spatialGrid.AddPlayerSegments(player, player.Snake.Body)
		}
	}

	// 食べ物をグリッドに追加
	for _, food := range g.Food {
		g.spatialGrid.AddFood(food)
	}
}
