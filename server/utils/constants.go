package utils

import "time"

const (
	FIELD_WIDTH           = 5000.0 // フィールドの幅
	FIELD_HEIGHT          = 3000.0 // フィールドの高さ
	SPHERE_RADIUS         = 15.0   // 球体の半径
	CELESTIAL_SPEED       = 700.0  // 天体システムの基本速度（ユニット/秒）
	CELESTIAL_ACCEL_FORCE = 700.0  // 天体システムの基本加速力

	// 衛星による速度減少設定
	SPEED_REDUCTION_PER_SATELLITE = 3.0                   // 衛星1個につき減少する速度（ユニット/秒）
	ACCEL_REDUCTION_PER_SATELLITE = 3.0                   // 衛星1個につき減少する加速力
	AIR_RESISTANCE                = 0.98                  // 空気抵抗（非アクティブ時の減衰）
	STOP_THRESHOLD_RATIO          = 0.02                  // 停止閾値（最大速度の2%）
	GAME_TICK                     = 16 * time.Millisecond // ゲーム更新間隔（60FPS）
	MAX_NPC_COUNT                 = 50                    // NPC数上限

	// カメラ・表示設定
	CAMERA_ZOOM_SCALE = 0.85 // カメラの固定ズーム倍率（0.85=少し引いた視点、物が少し小さく見える）

	// カリング設定
	CULLING_WIDTH  = 1920.0 // カリング範囲の幅
	CULLING_HEIGHT = 1080.0 // カリング範囲の高さ

	// デバッグ設定
	DISABLE_COLLISION = false // trueで当たり判定を無効化

	// 軌道物理定数
	ORBITAL_RADIUS_RATIO         = 3.0 // 軌道半径の基本倍率（コア半径の何倍か）
	ORBITAL_SPEED                = 2.0 // 軌道速度（ラジアン/秒）
	ORBITAL_CORRECTION_STRENGTH  = 0.5 // 軌道補正の強さ（0-1、大きいほど硬い軌道）
	ORBITAL_VELOCITY_INHERITANCE = 0.8 // 核の速度継承率（0-1）
	ANGLE_CORRECTION_SPEED       = 0.5 // 角度補正速度（ラジアン/秒）

	// 衝突物理定数
	COLLISION_RESTITUTION  = 0.3 // 反発係数（0-1、0は完全非弾性）
	COLLISION_MIN_DISTANCE = 2.0 // 最小衝突距離（半径の倍数）

	// 衛星物理定数
	SATELLITE_EJECT_SPEED = 1200.0 // 衛星の射出速度（ユニット/秒）

	// 自動衛星追加設定
	AUTO_SATELLITE_INTERVAL = 5 * time.Second // 自動衛星追加間隔
	MAX_AUTO_SATELLITES     = 10              // 自動追加の上限（2層目まで：第0軌道2個+第1軌道8個）

	// 落ちた衛星設定
	MIN_FALLEN_SATELLITES        = 10  // 落ちた衛星の最低数
	FALLEN_SATELLITES_PER_PLAYER = 3.0 // プレイヤー1人あたりの落ちた衛星数倍率

	// リスポーン設定
	RESPAWN_INVULNERABILITY_TIME = 5 * time.Second // リスポーン後の無敵時間
	RESPAWN_SAFE_DISTANCE        = 300.0           // リスポーン時の他プレイヤーからの最小距離

	// スコア設定
	SCORE_PICKUP_SATELLITE    = 10  // 落ちた衛星を拾った時のスコア
	SCORE_PER_SATELLITE_KILL  = 20  // 敵を倒した時の衛星1個あたりのスコア
	SCORE_DEATH_PENALTY_RATIO = 0.5 // 死亡時のスコア減少率（50%）
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
