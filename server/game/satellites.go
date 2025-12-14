package game

import (
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
		// FindSafeSpawnPosition を使用して安全な位置を取得
		pos := g.FindSafeSpawnPosition()

		newSatellite := &models.DroppedSatellite{
			Position:       pos,
			Radius:         utils.SPHERE_RADIUS,
			Color:          "#FFFFFF",
			IsOriginalCore: false,
		}
		g.DroppedSatellites = append(g.DroppedSatellites, newSatellite)
		// spatial gridに追加
		g.spatialGrid.AddDroppedSatellite(newSatellite)
	}
}
