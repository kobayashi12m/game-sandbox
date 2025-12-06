package game

import (
	"log"
	"math/rand/v2"

	"game-sandbox/server/models"
	"game-sandbox/server/utils"
)

// GenerateDroppedSatellites は落ちた衛星をフィールドに生成する
func (g *Game) GenerateDroppedSatellites() {
	targetCount := utils.MIN_FALLEN_SATELLITES
	if len(g.Players) > 0 {
		// プレイヤー数の倍率分の落ちた衛星を維持
		targetCount = max(int(float64(len(g.Players))*utils.FALLEN_SATELLITES_PER_PLAYER), utils.MIN_FALLEN_SATELLITES)
	}

	for len(g.DroppedSatellites) < targetCount {
		var pos models.Position
		attempts := 0
		for {
			pos = models.Position{
				X: rand.Float64() * utils.FIELD_WIDTH,
				Y: rand.Float64() * utils.FIELD_HEIGHT,
			}
			// 簡単な重複チェック（プレイヤーコアから一定距離離れているか）
			occupied := false
			for _, player := range g.Players {
				if player.Celestial.Core != nil {
					dx := pos.X - player.Celestial.Core.Position.X
					dy := pos.Y - player.Celestial.Core.Position.Y
					dist := dx*dx + dy*dy
					if dist < (utils.SPHERE_RADIUS*4)*(utils.SPHERE_RADIUS*4) {
						occupied = true
						break
					}
				}
			}
			if !occupied || attempts > 100 {
				break
			}
			attempts++
		}
		if attempts <= 100 {
			newSatellite := &models.DroppedSatellite{
				Position:       pos,
				Radius:         utils.SPHERE_RADIUS,
				Color:          "#FFFFFF",
				IsOriginalCore: false,
			}
			g.DroppedSatellites = append(g.DroppedSatellites, newSatellite)
			// spatial gridに追加
			g.spatialGrid.AddDroppedSatellite(newSatellite)
		} else {
			log.Printf("Failed to place dropped satellite after 100 attempts (current: %d, target: %d)",
				len(g.DroppedSatellites), targetCount)
			break
		}
	}
}
