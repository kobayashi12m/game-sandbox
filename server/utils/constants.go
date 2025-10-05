package utils

import "time"

const (
	FIELD_WIDTH  = 5000.0                // フィールドの幅（大幅拡大）
	FIELD_HEIGHT = 3000.0                // フィールドの高さ（大幅拡大）
	SNAKE_RADIUS = 15.0                  // 蛇の半径
	FOOD_RADIUS  = 10.0                  // 食べ物の半径
	SNAKE_SPEED  = 300.0                 // 蛇の速度（ユニット/秒）
	GAME_TICK    = 16 * time.Millisecond // ゲーム更新間隔（60FPS）
	NPC_COUNT    = 100                   // デフォルトNPC数
	
	// デバッグ設定
	DISABLE_COLLISION = true // trueで当たり判定を無効化
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
