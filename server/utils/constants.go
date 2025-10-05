package utils

import "time"

const (
	FIELD_WIDTH  = 1800.0                // フィールドの幅
	FIELD_HEIGHT = 1000.0                // フィールドの高さ
	SNAKE_RADIUS = 12.0                  // 蛇の半径
	FOOD_RADIUS  = 8.0                   // 食べ物の半径
	SNAKE_SPEED  = 250.0                 // 蛇の速度（ユニット/秒）
	GAME_TICK    = 16 * time.Millisecond // ゲーム更新間隔（60FPS）
)

// Direction represents movement direction (normalized vector)
type Direction struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

var (
	DIRECTIONS = map[string]Direction{
		"UP":    {X: 0, Y: -1},
		"DOWN":  {X: 0, Y: 1},
		"LEFT":  {X: -1, Y: 0},
		"RIGHT": {X: 1, Y: 0},
	}
)
