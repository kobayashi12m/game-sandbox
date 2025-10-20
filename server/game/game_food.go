package game

import (
	"log"
	"math/rand/v2"

	"game-sandbox/server/models"
	"game-sandbox/server/utils"
)

// GenerateFood はゲームフィールドに食べ物を生成する
func (g *Game) GenerateFood() {
	targetFoodCount := 5 // 最小数を増加
	if len(g.Players) > 0 {
		// プレイヤー数の2倍に増加
		targetFoodCount = max(int(float64(len(g.Players))*2.0), 5)
	}

	for len(g.Food) < targetFoodCount {
		var pos models.Position
		attempts := 0
		for {
			pos = models.Position{
				X: rand.Float64() * utils.FIELD_WIDTH,
				Y: rand.Float64() * utils.FIELD_HEIGHT,
			}
			if !g.spatialGrid.IsPositionOccupiedOptimized(pos) || attempts > 100 {
				break
			}
			attempts++
		}
		if attempts <= 100 {
			g.Food = append(g.Food, &models.Food{Position: pos})
		} else {
			log.Printf("Failed to place food after 100 attempts (current food: %d, target: %d)",
				len(g.Food), targetFoodCount)
			break // 無限ループを防ぐ
		}
	}
}

// RemoveFood は食べ物をポインタで効率的に削除する
func (g *Game) RemoveFood(targetFood *models.Food) {
	for i, food := range g.Food {
		if food == targetFood {
			g.Food = append(g.Food[:i], g.Food[i+1:]...)
			return
		}
	}
}
