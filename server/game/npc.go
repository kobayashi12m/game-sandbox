package game

import (
	"game-sandbox/server/models"
	"game-sandbox/server/utils"
	"math"
	"math/rand"
	"time"
)

// NPCを追加
func (g *Game) AddNPC(count int) {
	names := []string{"Bot Alpha", "Bot Beta", "Bot Gamma", "Bot Delta", "Bot Epsilon",
		"Bot Zeta", "Bot Eta", "Bot Theta", "Bot Iota", "Bot Kappa"}

	for i := range count {
		npcID := utils.GenerateID()
		npcName := names[i%len(names)]
		if i >= len(names) {
			npcName = "Bot " + string(rune('A'+i))
		}

		// 既存のAddPlayer関数を使ってNPCを追加（WebSocket接続はnil）
		g.AddPlayer(npcID, npcName, nil)

		utils.LogConnectionEvent("npc_joined", npcID, npcName, true)
	}

	utils.Info("NPCs added to game", map[string]interface{}{
		"event":         "npc_batch_add",
		"game_id":       g.ID,
		"npc_count":     count,
		"total_players": len(g.Players),
	})
}

// NPCのAIを更新 - 毎フレーム加速度を設定
func (g *Game) UpdateNPCAI() {
	// 射撃予定のNPCを記録
	type shootAction struct {
		npc  *models.Player
		dirX float64
		dirY float64
	}
	var pendingShoots []shootAction

	for _, npc := range g.Players {
		if !npc.IsNPC || !npc.Celestial.Alive {
			continue
		}

		// 行動更新と射撃判定
		if shoot := g.updateNPCBehavior(npc); shoot != nil {
			pendingShoots = append(pendingShoots, shootAction{
				npc:  shoot.npc,
				dirX: shoot.dirX,
				dirY: shoot.dirY,
			})
		}
	}

	// 射撃処理を後でまとめて実行（ロックを取らない内部メソッドを使用）
	for _, shoot := range pendingShoots {
		g.ejectPlayerSatelliteNoLock(shoot.npc, shoot.dirX, shoot.dirY)
		shoot.npc.LastShoot = time.Now()
	}
}

type shootAction struct {
	npc  *models.Player
	dirX float64
	dirY float64
}

// NPCの行動を更新
func (g *Game) updateNPCBehavior(npc *models.Player) *shootAction {
	// ターゲット探索（0.5秒ごと）
	if time.Since(npc.LastDirectionChange) > 500*time.Millisecond {
		target, targetType := g.findBestTarget(npc)

		if target != nil {
			// ターゲットへ向かう
			dx := target.Core.Position.X - npc.Celestial.Core.Position.X
			dy := target.Core.Position.Y - npc.Celestial.Core.Position.Y
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist > 0 {
				// 予測位置を計算
				predictedX := target.Core.Position.X + target.Core.Velocity.X*0.3
				predictedY := target.Core.Position.Y + target.Core.Velocity.Y*0.3

				dx = predictedX - npc.Celestial.Core.Position.X
				dy = predictedY - npc.Celestial.Core.Position.Y
				norm := math.Sqrt(dx*dx + dy*dy)

				if norm > 0 {
					npc.TargetDirection = &struct{ X, Y float64 }{
						X: dx / norm,
						Y: dy / norm,
					}

					// 敵プレイヤーが近い場合、射撃を検討
					if targetType == "enemy" && dist < 300 { // 射程
						if g.shouldNPCShoot(npc) {
							return &shootAction{
								npc:  npc,
								dirX: dx / norm,
								dirY: dy / norm,
							}
						}
					}
				}
			}
		} else {
			// ターゲットがない場合はランダム移動
			g.randomWander(npc)
		}

		npc.LastDirectionChange = time.Now()
	}

	// 加速度を設定
	if npc.TargetDirection != nil {
		npc.Celestial.SetAcceleration(npc.TargetDirection.X, npc.TargetDirection.Y)
	}

	return nil
}

// 最適なターゲットを探す
func (g *Game) findBestTarget(npc *models.Player) (*models.Celestial, string) {
	var bestTarget *models.Celestial
	var targetType string
	minScore := math.MaxFloat64

	npcPos := npc.Celestial.Core.Position
	mySatellites := npc.Celestial.GetTotalSatelliteCount()
	searchRadius := 600.0 // 探索範囲

	// spatial gridで周囲のオブジェクトを取得
	minX := npcPos.X - searchRadius
	maxX := npcPos.X + searchRadius
	minY := npcPos.Y - searchRadius
	maxY := npcPos.Y + searchRadius
	nearbyObjects := g.spatialGrid.GetObjectsInArea(minX, maxX, minY, maxY)

	// 落ちている衛星を探す（最優先）
	for _, satellite := range nearbyObjects.DroppedSatellites {
		dx := satellite.Position.X - npcPos.X
		dy := satellite.Position.Y - npcPos.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < searchRadius {
			// 距離に基づくスコア（近いほど良い）
			score := dist

			if score < minScore {
				minScore = score
				bestTarget = &models.Celestial{
					Core: &models.Sphere{
						Position: satellite.Position,
						Velocity: models.Position{X: 0, Y: 0},
					},
					Alive: true,
				}
				targetType = "satellite"
			}
		}
	}

	// 他のプレイヤー（NPCも含む）を探す
	if mySatellites >= 1 { // 衛星が1個以上あれば攻撃を検討
		for _, player := range nearbyObjects.Players {
			if player.ID == npc.ID || !player.Celestial.Alive || player.IsInvulnerable() {
				continue
			}

			dist := g.distance(npc.Celestial, player.Celestial)
			if dist < searchRadius {
				theirSatellites := player.Celestial.GetTotalSatelliteCount()

				// 自分より弱い相手、または同等くらいの相手も狙う
				if mySatellites+2 >= theirSatellites { // 相手が自分+2個以下なら攻撃
					// 距離と相手の弱さを考慮したスコア
					score := dist + float64(theirSatellites)*50

					if score < minScore {
						minScore = score
						bestTarget = player.Celestial
						targetType = "enemy"
					}
				}
			}
		}
	}

	return bestTarget, targetType
}

// NPCが射撃すべきか判定
func (g *Game) shouldNPCShoot(npc *models.Player) bool {
	// クールダウン中は撃たない
	if time.Since(npc.LastShoot) < 300*time.Millisecond { // 800ms→300msに短縮
		return false
	}

	// 衛星がない場合は撃てない
	if npc.Celestial.GetTotalSatelliteCount() == 0 {
		return false
	}

	// 一定確率で射撃（常に撃つと不自然なので）
	return rand.Float64() < 0.95 // 70%→95%に大幅上昇
}

// ランダムな徘徊
func (g *Game) randomWander(npc *models.Player) {
	// フィールドの端に近い場合は中央へ向かう
	margin := 300.0
	if npc.Celestial.Core.Position.X < margin || npc.Celestial.Core.Position.X > utils.FIELD_WIDTH-margin ||
		npc.Celestial.Core.Position.Y < margin || npc.Celestial.Core.Position.Y > utils.FIELD_HEIGHT-margin {
		centerX := utils.FIELD_WIDTH / 2
		centerY := utils.FIELD_HEIGHT / 2
		dx := centerX - npc.Celestial.Core.Position.X
		dy := centerY - npc.Celestial.Core.Position.Y
		norm := math.Sqrt(dx*dx + dy*dy)
		if norm > 0 {
			npc.TargetDirection = &struct{ X, Y float64 }{
				X: dx / norm,
				Y: dy / norm,
			}
			return
		}
	}

	// ランダムな方向へ移動
	angle := rand.Float64() * 2 * math.Pi
	npc.TargetDirection = &struct{ X, Y float64 }{
		X: math.Cos(angle),
		Y: math.Sin(angle),
	}
}

// 2つの天体間の距離を計算
func (g *Game) distance(a, b *models.Celestial) float64 {
	if a == nil || b == nil || a.Core == nil || b.Core == nil {
		return math.MaxFloat64
	}
	dx := b.Core.Position.X - a.Core.Position.X
	dy := b.Core.Position.Y - a.Core.Position.Y
	return math.Sqrt(dx*dx + dy*dy)
}
