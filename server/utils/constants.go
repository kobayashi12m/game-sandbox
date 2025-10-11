package utils

import "time"

const (
	FIELD_WIDTH          = 5000.0                // フィールドの幅
	FIELD_HEIGHT         = 3000.0                // フィールドの高さ
	ORGANISM_RADIUS      = 15.0                  // オーガニズムの半径
	FOOD_RADIUS          = 10.0                  // 食べ物の半径
	ORGANISM_SPEED       = 2000.0                // オーガニズムの速度（ユニット/秒）
	ORGANISM_ACCEL_FORCE = 1000.0                // オーガニズムの加速力
	AIR_RESISTANCE       = 0.98                  // 空気抵抗（非アクティブ時の減衰）
	STOP_THRESHOLD_RATIO = 0.02                  // 停止閾値（最大速度の2%）
	GAME_TICK            = 16 * time.Millisecond // ゲーム更新間隔（60FPS）
	NPC_COUNT            = 50                    // デフォルトNPC数

	// カリング設定
	CULLING_WIDTH  = 1300.0 // カリング範囲の幅
	CULLING_HEIGHT = 800.0  // カリング範囲の高さ

	// デバッグ設定
	DISABLE_COLLISION = false // trueで当たり判定を無効化

	// 物理シミュレーション定数
	CONNECTION_STIFFNESS      = 50.0  // ばね定数（復元力の強さ）
	CONNECTION_DAMPING        = 1.5   // ダンピング係数（振動抑制）
	CONNECTION_NATURAL_RATIO  = 4.0   // 自然長の倍率（半径の何倍か）
	CONNECTION_MAX_RATIO      = 16.0  // 最大長の倍率（半径の何倍か）
	RING_CONNECTION_STRENGTH  = 2.0   // 環状接続の強度
	NODE_REPULSION_FORCE      = 200.0 // ノード間反発力
	ANGULAR_RESTORATION_FORCE = 30.0  // 角度復元力（絡まり防止）
	
	// 衝突物理定数
	COLLISION_RESTITUTION     = 0.3   // 反発係数（0-1、0は完全非弾性）
	COLLISION_MIN_DISTANCE    = 2.0   // 最小衝突距離（半径の倍数）
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
