package game

import (
	"math"
	"math/rand"
	"time"

	"game-sandbox/server/models"
	"game-sandbox/server/utils"
)

// generateRandomPosition はランダムな位置を生成する
func generateRandomPosition() models.Position {
	x := rand.Float64()*(utils.FIELD_WIDTH-100) + 50
	y := rand.Float64()*(utils.FIELD_HEIGHT-100) + 50
	return models.Position{X: x, Y: y}
}

// FindSafeSpawnPosition は安全なリスポーン位置を見つける
func (g *Game) FindSafeSpawnPosition() models.Position {
	maxAttempts := 100
	minDistance := utils.RESPAWN_SAFE_DISTANCE

	for attempt := 0; attempt < maxAttempts; attempt++ {
		pos := generateRandomPosition()
		if g.isPositionSafe(pos, minDistance) {
			return pos
		}
	}

	// 安全な場所が見つからない場合は、ランダムな位置を返す
	return generateRandomPosition()
}

// isPositionSafe は指定位置が他のプレイヤーや落ちた衛星から十分離れているかチェック
func (g *Game) isPositionSafe(pos models.Position, minDistance float64) bool {
	// 他のプレイヤーとの距離をチェック
	for _, player := range g.Players {
		if !player.Celestial.Alive || player.Celestial.Core == nil {
			continue
		}

		// コアとの距離
		dx := pos.X - player.Celestial.Core.Position.X
		dy := pos.Y - player.Celestial.Core.Position.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance < minDistance {
			return false
		}

		// 衛星との距離もチェック
		for _, orbit := range player.Celestial.Satellites {
			for _, sat := range orbit {
				dx := pos.X - sat.Sphere.Position.X
				dy := pos.Y - sat.Sphere.Position.Y
				distance := math.Sqrt(dx*dx + dy*dy)

				if distance < minDistance/2 { // 衛星は少し近くても許容
					return false
				}
			}
		}
	}

	// 落ちた衛星との距離をチェック
	for _, droppedSat := range g.DroppedSatellites {
		dx := pos.X - droppedSat.Position.X
		dy := pos.Y - droppedSat.Position.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		// 落ちた衛星とは最低限の距離（半径の3倍）を保つ
		if distance < utils.SPHERE_RADIUS*3 {
			return false
		}
	}

	return true
}

// RespawnPlayer はプレイヤーを安全な位置でリスポーンさせる
func (g *Game) RespawnPlayer(player *models.Player) {
	// 安全な位置を見つける
	safePos := g.FindSafeSpawnPosition()

	// Celestialを再初期化（安全な位置で）
	player.Celestial.ResetAtPosition(safePos.X, safePos.Y)

	// リスポーン時刻を記録（無敵時間の開始）
	player.RespawnTime = time.Now()

	// ログ出力
	utils.LogPlayerAction("respawn", player.ID, player.Name, map[string]interface{}{
		"position_x":            safePos.X,
		"position_y":            safePos.Y,
		"invulnerable_duration": utils.RESPAWN_INVULNERABILITY_TIME.Seconds(),
	})
}
