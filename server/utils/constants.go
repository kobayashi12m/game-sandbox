package utils

import "time"

const (
	FIELD_WIDTH     = 5000.0                // フィールドの幅（大幅拡大）
	FIELD_HEIGHT    = 3000.0                // フィールドの高さ（大幅拡大）
	ORGANISM_RADIUS = 15.0                  // オーガニズムの半径
	FOOD_RADIUS     = 10.0                  // 食べ物の半径
	ORGANISM_SPEED  = 300.0                 // オーガニズムの速度（ユニット/秒）
	GAME_TICK       = 16 * time.Millisecond // ゲーム更新間隔（60FPS）
	NPC_COUNT       = 100                   // デフォルトNPC数

	// カリング設定
	CULLING_WIDTH  = 1300.0 // カリング範囲の幅
	CULLING_HEIGHT = 800.0  // カリング範囲の高さ

	// デバッグ設定
	DISABLE_COLLISION = false // trueで当たり判定を無効化

	// 物理シミュレーション定数
	CONNECTION_STIFFNESS      = 50.0  // ばね定数（復元力の強さ）- 大幅強化
	CONNECTION_DAMPING        = 1.5   // ダンピング係数（振動抑制）- 強化
	CONNECTION_NATURAL_RATIO  = 4.0   // 自然長の倍率（半径の何倍か）
	CONNECTION_MAX_RATIO      = 16.0  // 最大長の倍率（半径の何倍か）
	RING_CONNECTION_STRENGTH  = 2.0   // 環状接続の強度 - 大幅強化
	NODE_REPULSION_FORCE      = 200.0 // ノード間反発力 - 大幅強化
	ANGULAR_RESTORATION_FORCE = 30.0  // 角度復元力（絡まり防止）
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
