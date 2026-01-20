package game

import (
	"math"
	"math/rand"
	"time"

	"game-sandbox/server/celestial"
	"game-sandbox/server/models"
	"game-sandbox/server/types"
	"game-sandbox/server/utils"
)

// NPCを追加
func (g *Game) AddNPC(count int) {
	for range count {
		npcID := utils.GenerateID()
		npcName := utils.GenerateRandomNickname()

		g.AddPlayer(npcID, npcName, nil)
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
	for _, npc := range g.Players {
		if !npc.IsNPC || !npc.Celestial.Alive {
			continue
		}

		// 行動更新
		g.updateNPCBehavior(npc)
	}
}

// NPCのターゲット情報
type npcTarget struct {
	position   types.Position
	velocity   types.Position
	targetType string // "satellite" or "enemy"
}

// 予測位置を計算
func (g *Game) predictPosition(position, velocity types.Position, deltaTime float64) types.Position {
	return types.Position{
		X: position.X + velocity.X*deltaTime,
		Y: position.Y + velocity.Y*deltaTime,
	}
}

// NPCの行動を更新
func (g *Game) updateNPCBehavior(npc *models.Player) {
	// グリッドセルサイズを取得
	cellSize := g.spatialGrid.cellSize

	// ターゲット探索（0.5秒ごと）
	if time.Since(npc.LastDirectionChange) > 500*time.Millisecond {
		// ターゲットを探す
		target := g.findBestTarget(npc, cellSize)

		// ターゲットに応じた行動を決定
		if target == nil {
			g.randomWander(npc, cellSize)
		} else {
			g.moveTowardTarget(npc, target)
		}

		npc.LastDirectionChange = time.Now()
	}

	// 加速度を設定
	if npc.TargetDirection != nil {
		g.SendCommand(AccelerationCommand{
			Player: npc,
			X:      npc.TargetDirection.X,
			Y:      npc.TargetDirection.Y,
		})
	}
}

// ターゲットへ向かって移動し、必要なら直接射撃
func (g *Game) moveTowardTarget(npc *models.Player, target *npcTarget) {
	// 現在位置への距離
	dx := target.position.X - npc.Celestial.Core.Position.X
	dy := target.position.Y - npc.Celestial.Core.Position.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist <= 0 {
		return
	}

	// 予測位置へ向かう方向を計算
	predictedPos := g.predictPosition(target.position, target.velocity, 0.3)
	dx = predictedPos.X - npc.Celestial.Core.Position.X
	dy = predictedPos.Y - npc.Celestial.Core.Position.Y
	norm := math.Sqrt(dx*dx + dy*dy)

	if norm <= 0 {
		return
	}

	// 移動方向を設定
	npc.TargetDirection = &struct{ X, Y float64 }{
		X: dx / norm,
		Y: dy / norm,
	}

	// 敵への射撃判定
	if target.targetType == "enemy" && g.shouldNPCShoot(npc) {
		g.SendCommand(ShootCommand{
			Player:  npc,
			TargetX: dx / norm,
			TargetY: dy / norm,
		})
		npc.LastShoot = time.Now()
	}
}

// 最適なターゲットを探す
func (g *Game) findBestTarget(npc *models.Player, cellSize float64) *npcTarget {
	var bestTarget *npcTarget
	minScore := math.MaxFloat64

	npcPos := npc.Celestial.Core.Position
	mySatellites := npc.Celestial.GetTotalSatelliteCount()
	searchRadius := cellSize * 7.0 // グリッド7個分の探索範囲

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
				bestTarget = &npcTarget{
					position:   satellite.Position,
					velocity:   types.Position{X: 0, Y: 0},
					targetType: "satellite",
				}
			}
		}
	}

	// 他のプレイヤー（NPCも含む）を探す
	if mySatellites >= 12 { // 衛星が12個以上あれば攻撃を検討
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
						bestTarget = &npcTarget{
							position:   player.Celestial.Core.Position,
							velocity:   player.Celestial.Core.Velocity,
							targetType: "enemy",
						}
					}
				}
			}
		}
	}

	return bestTarget
}

// NPCが射撃すべきか判定
func (g *Game) shouldNPCShoot(npc *models.Player) bool {
	// クールダウン中は撃たない
	if time.Since(npc.LastShoot) < 300*time.Millisecond { // 300ms
		return false
	}

	// 衛星がない場合は撃てない
	if npc.Celestial.GetTotalSatelliteCount() == 0 {
		return false
	}

	// 一定確率で射撃
	return rand.Float64() < 0.95
}

// ランダムな徘徊
func (g *Game) randomWander(npc *models.Player, cellSize float64) {
	// フィールドの端に近い場合は中央へ向かう（グリッド5個分のマージン）
	margin := cellSize * 5.0
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
func (g *Game) distance(a, b *celestial.Celestial) float64 {
	if a == nil || b == nil || a.Core == nil || b.Core == nil {
		return math.MaxFloat64
	}
	dx := b.Core.Position.X - a.Core.Position.X
	dy := b.Core.Position.Y - a.Core.Position.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// GetDesiredNPCCount は現在の人間プレイヤー数から必要なNPC数を計算する（ロック済み状態で呼ぶこと）
func (g *Game) GetDesiredNPCCount() int {
	humanCount := g.humanPlayerCount()
	desiredNPCCount := utils.MAX_NPC_COUNT - humanCount
	if desiredNPCCount < 0 {
		return 0
	}
	return desiredNPCCount
}

// ReplenishNPCs は不足したNPCを補充する
func (g *Game) ReplenishNPCs() {
	g.mu.Lock()
	humanCount := g.humanPlayerCount()
	desiredNPCCount := g.GetDesiredNPCCount()
	currentNPCCount := len(g.Players) - humanCount
	g.mu.Unlock()

	// NPCが不足している場合は追加
	if currentNPCCount < desiredNPCCount {
		toAdd := desiredNPCCount - currentNPCCount
		g.AddNPC(toAdd)
		utils.Info("NPCs added to maintain player count", map[string]interface{}{
			"event":       "npc_add_maintain",
			"game_id":     g.ID,
			"human_count": humanCount,
			"current_npc": currentNPCCount,
			"desired_npc": desiredNPCCount,
			"added":       toAdd,
		})
	}
}

// ShouldNPCRespawn はNPCがリスポーンすべきかを判定する（ロック済み状態で呼ぶこと）
func (g *Game) ShouldNPCRespawn(npc *models.Player) bool {
	if !npc.IsNPC {
		return true // 人間プレイヤーは常にリスポーン
	}

	desiredNPCCount := g.GetDesiredNPCCount()
	humanCount := g.humanPlayerCount()
	currentNPCCount := len(g.Players) - humanCount

	// NPC数が上限以下ならリスポーン
	if currentNPCCount <= desiredNPCCount {
		return true
	}

	// NPC数が上限を超えている場合、スコアランキングをチェック
	// 効率的な方法：自分のスコアより高いNPCの数を数える
	higherScoreCount := 0
	for _, player := range g.Players {
		if player.IsNPC && player.ID != npc.ID && player.Score > npc.Score {
			higherScoreCount++
			if higherScoreCount >= 10 {
				// 10人以上が自分より高スコアならNPCを削除
				delete(g.Players, npc.ID)
				utils.Info("NPC removed due to low rank", map[string]interface{}{
					"event":       "npc_removed_low_rank",
					"game_id":     g.ID,
					"npc_id":      npc.ID,
					"npc_name":    npc.Name,
					"npc_score":   npc.Score,
					"human_count": humanCount,
					"npc_count":   currentNPCCount - 1,
				})
				return false
			}
		}
	}

	// 上位10位以内ならリスポーン
	return true
}
