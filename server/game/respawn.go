package game

import (
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
	maxAttempts := 50

	// Spatial Gridを使った高速チェック
	for range maxAttempts {
		pos := generateRandomPosition()

		// Spatial Gridでその位置のセルを取得
		cellX, cellY := g.spatialGrid.GetCellCoords(pos.X, pos.Y)

		// そのセルにプレイヤーがいなければOK
		if len(g.spatialGrid.cells[cellY][cellX].playerSpheres) == 0 {
			return pos
		}
	}

	// 見つからない場合は、ランダムな位置を返す
	return generateRandomPosition()
}

// spawnPlayerInternal はプレイヤーを安全な位置でスポーンさせる（内部共通処理）
func (g *Game) spawnPlayerInternal(player *models.Player, isRespawn bool) {
	// 安全な位置を見つける
	safePos := g.FindSafeSpawnPosition()

	// Celestialを再初期化（安全な位置で）
	player.Celestial.ResetAtPosition(safePos.X, safePos.Y)

	// スポーン時刻を記録（無敵時間の開始）
	player.RespawnTime = time.Now()

	// 自動衛星タイマーを初期化
	player.LastAutoSatellite = time.Now()

}

// SpawnPlayer は新規プレイヤーを初期スポーンさせる
func (g *Game) SpawnPlayer(player *models.Player) {
	g.spawnPlayerInternal(player, false)
}

// RespawnPlayer はプレイヤーをリスポーンさせる（死後の処理用）
func (g *Game) RespawnPlayer(player *models.Player) {
	g.spawnPlayerInternal(player, true)
}
