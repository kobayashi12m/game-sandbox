package models

import (
	"game-sandbox/server/utils"
	"math"
	"math/rand/v2"
	"time"
)

// Sphere は物理演算される球体を表す
type Sphere struct {
	Position     Position `json:"position"`
	Velocity     Position `json:"velocity,omitempty"`
	Acceleration Position `json:"acceleration,omitempty"` // 加速度
	Radius       float64  `json:"radius"`
	Mass         float64  `json:"-"` // 質量
}

// Satellite は衛星を表す
type Satellite struct {
	Sphere    *Sphere `json:"sphere"`    // 球体
	OrbitType int     `json:"orbitType"` // 軌道番号（1=第1軌道、2=第2軌道...）
	Angle     float64 `json:"angle"`     // 軌道上の角度
}

// OrbitConfig は軌道の設定
type OrbitConfig struct {
	Radius float64 // 軌道半径
	Speed  float64 // 回転速度（ラジアン/秒）
}

// Celestial は核と衛星からなる天体システムを表す
type Celestial struct {
	Core       *Sphere   `json:"core"`  // 中心球
	Nodes      []*Sphere `json:"nodes"` // クライアント用（JSON送信時のみ）
	Color      string    `json:"color"`
	Alive      bool      `json:"alive"`
	Growing    int       `json:"-"`
	Respawning bool      `json:"-"`
	DeathTime  time.Time `json:"-"`

	// 内部管理用（JSON送信されない）
	Satellites   []*Satellite         `json:"-"` // 衛星
	MaxSpeed     float64              `json:"-"` // 最大速度
	AccelForce   float64              `json:"-"` // 加速力
	OrbitConfigs map[int]*OrbitConfig `json:"-"` // 軌道設定
}

// UpdateNodes はJSON送信前にNodesフィールドを更新する
func (c *Celestial) UpdateNodes() {
	c.Nodes = c.GetAllSpheres()
}

// Reset は天体システムを初期状態に初期化する
func (c *Celestial) Reset() {
	// フィールド内のランダムな位置にスポーン
	startX := rand.Float64()*(utils.FIELD_WIDTH-100) + 50
	startY := rand.Float64()*(utils.FIELD_HEIGHT-100) + 50

	// コア（中心球）を初期化
	c.Core = &Sphere{
		Position:     Position{X: startX, Y: startY},
		Velocity:     Position{X: 0, Y: 0},
		Acceleration: Position{X: 0, Y: 0},
		Radius:       utils.SPHERE_RADIUS,
		Mass:         1.0,
	}

	// 軌道設定を初期化
	c.OrbitConfigs = map[int]*OrbitConfig{
		1: {
			Radius: utils.SPHERE_RADIUS * utils.ORBITAL_RADIUS_RATIO,
			Speed:  utils.ORBITAL_SPEED,
		},
	}

	// 初期衛星を配置（第1軌道に4個）
	c.Satellites = []*Satellite{}
	nodeCount := 4

	for i := 0; i < nodeCount; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(nodeCount) // 等間隔で配置
		orbitConfig := c.OrbitConfigs[1]

		nodeX := startX + orbitConfig.Radius*math.Cos(angle)
		nodeY := startY + orbitConfig.Radius*math.Sin(angle)

		// 軌道接線方向の初期速度を計算
		tangentVelX := -orbitConfig.Radius * orbitConfig.Speed * math.Sin(angle)
		tangentVelY := orbitConfig.Radius * orbitConfig.Speed * math.Cos(angle)

		sphere := &Sphere{
			Position:     Position{X: nodeX, Y: nodeY},
			Velocity:     Position{X: tangentVelX, Y: tangentVelY},
			Acceleration: Position{X: 0, Y: 0},
			Radius:       utils.SPHERE_RADIUS,
			Mass:         0.5, // 衛星はコアより軽い
		}

		satellite := &Satellite{
			Sphere:    sphere,
			OrbitType: 1, // 第1軌道
			Angle:     angle,
		}

		c.Satellites = append(c.Satellites, satellite)
	}

	c.Growing = 0
	c.Alive = true
	c.Respawning = false

	// 天体システムの移動パラメータ
	c.MaxSpeed = utils.CELESTIAL_SPEED
	c.AccelForce = utils.CELESTIAL_ACCEL_FORCE // 加速力

	// JSON用のNodesを更新
	c.UpdateNodes()
}

// UpdateMotion は天体システムの運動を更新する
func (c *Celestial) UpdateMotion(deltaTime float64) {
	if !c.Alive {
		return
	}

	// 1. コアの運動更新
	c.updateCoreMotion(deltaTime)

	// 2. 衛星の軌道更新
	c.updateSatelliteOrbits(deltaTime)

	// 3. JSON用のNodesを更新
	c.UpdateNodes()

	// 4. 衝突処理
	c.handleSphereCollisions(deltaTime)
	c.applyBoundaryCollision()
}

// updateCoreMotion はコアの運動のみを更新する
func (c *Celestial) updateCoreMotion(deltaTime float64) {
	// 加速度を速度に適用
	c.Core.Velocity.X += c.Core.Acceleration.X * deltaTime
	c.Core.Velocity.Y += c.Core.Acceleration.Y * deltaTime

	// 最大速度制限
	speed := math.Sqrt(c.Core.Velocity.X*c.Core.Velocity.X + c.Core.Velocity.Y*c.Core.Velocity.Y)
	if speed > c.MaxSpeed {
		c.Core.Velocity.X = (c.Core.Velocity.X / speed) * c.MaxSpeed
		c.Core.Velocity.Y = (c.Core.Velocity.Y / speed) * c.MaxSpeed
	}

	// 位置を更新
	c.Core.Position.X += c.Core.Velocity.X * deltaTime
	c.Core.Position.Y += c.Core.Velocity.Y * deltaTime

	// 空気抵抗を適用
	c.Core.Velocity.X *= utils.AIR_RESISTANCE
	c.Core.Velocity.Y *= utils.AIR_RESISTANCE

	// 低速時の停止判定（入力がない時のみ）
	hasInput := c.Core.Acceleration.X != 0 || c.Core.Acceleration.Y != 0
	if !hasInput {
		speed := math.Sqrt(c.Core.Velocity.X*c.Core.Velocity.X + c.Core.Velocity.Y*c.Core.Velocity.Y)
		if speed < c.MaxSpeed*utils.STOP_THRESHOLD_RATIO {
			c.Core.Velocity.X = 0
			c.Core.Velocity.Y = 0
		}
	}
}

// updateSatelliteOrbits は衛星の軌道運動のみを更新する
func (c *Celestial) updateSatelliteOrbits(deltaTime float64) {
	for _, satellite := range c.Satellites {
		// 軌道設定を取得
		orbitConfig := c.GetOrbitConfig(satellite.OrbitType)

		// 角度を更新
		satellite.Angle += orbitConfig.Speed * deltaTime

		// 理想的な軌道位置を計算
		idealX := c.Core.Position.X + orbitConfig.Radius*math.Cos(satellite.Angle)
		idealY := c.Core.Position.Y + orbitConfig.Radius*math.Sin(satellite.Angle)

		// 現在位置から理想位置への差
		dx := idealX - satellite.Sphere.Position.X
		dy := idealY - satellite.Sphere.Position.Y

		// スムーズに理想位置に移動（強めの補正）
		satellite.Sphere.Position.X += dx * utils.ORBITAL_CORRECTION_STRENGTH
		satellite.Sphere.Position.Y += dy * utils.ORBITAL_CORRECTION_STRENGTH

		// 軌道速度を計算（接線方向）
		tangentX := -math.Sin(satellite.Angle) * orbitConfig.Radius * orbitConfig.Speed
		tangentY := math.Cos(satellite.Angle) * orbitConfig.Radius * orbitConfig.Speed

		// 核の速度を継承（核と一緒に移動する感じを出す）
		satellite.Sphere.Velocity.X = tangentX + c.Core.Velocity.X*utils.ORBITAL_VELOCITY_INHERITANCE
		satellite.Sphere.Velocity.Y = tangentY + c.Core.Velocity.Y*utils.ORBITAL_VELOCITY_INHERITANCE
	}

}

// applyBoundaryCollision はフィールド境界での衝突処理を適用
func (c *Celestial) applyBoundaryCollision() {
	// コアの境界衝突処理
	if c.Core.Position.X-c.Core.Radius < 0 {
		c.Core.Position.X = c.Core.Radius
		c.Core.Velocity.X = -c.Core.Velocity.X * 0.5 // 反発係数0.5
	} else if c.Core.Position.X+c.Core.Radius >= utils.FIELD_WIDTH {
		c.Core.Position.X = utils.FIELD_WIDTH - c.Core.Radius
		c.Core.Velocity.X = -c.Core.Velocity.X * 0.5
	}
	if c.Core.Position.Y-c.Core.Radius < 0 {
		c.Core.Position.Y = c.Core.Radius
		c.Core.Velocity.Y = -c.Core.Velocity.Y * 0.5
	} else if c.Core.Position.Y+c.Core.Radius >= utils.FIELD_HEIGHT {
		c.Core.Position.Y = utils.FIELD_HEIGHT - c.Core.Radius
		c.Core.Velocity.Y = -c.Core.Velocity.Y * 0.5
	}

	// ノードは軌道上で自動的に動くため、境界衝突処理は不要
}

// SetAcceleration は加速度を直接設定する（360度自由移動用）
func (c *Celestial) SetAcceleration(x, y float64) {
	// 入力値を-1〜1の範囲に制限
	if x > 1.0 {
		x = 1.0
	} else if x < -1.0 {
		x = -1.0
	}
	if y > 1.0 {
		y = 1.0
	} else if y < -1.0 {
		y = -1.0
	}

	// ベクトルの大きさが1を超えないように正規化
	magnitude := math.Sqrt(x*x + y*y)
	if magnitude > 1.0 {
		x /= magnitude
		y /= magnitude
	}

	c.Core.Acceleration.X = x * c.AccelForce
	c.Core.Acceleration.Y = y * c.AccelForce
}

// AddSatellite は新しい衛星を追加する（成長時）
func (c *Celestial) AddSatellite() {
	// 第1軌道に追加（将来的には軌道選択ロジックを追加）
	orbitType := 1

	// 軌道設定が存在しない場合は作成
	if _, exists := c.OrbitConfigs[orbitType]; !exists {
		c.AddOrbitType(orbitType, utils.SPHERE_RADIUS*utils.ORBITAL_RADIUS_RATIO, utils.ORBITAL_SPEED)
	}

	orbitConfig := c.GetOrbitConfig(orbitType)

	// 新しい衛星を作成
	angle := rand.Float64() * 2.0 * math.Pi // ランダムな角度

	coreX := c.Core.Position.X
	coreY := c.Core.Position.Y
	nodeX := coreX + orbitConfig.Radius*math.Cos(angle)
	nodeY := coreY + orbitConfig.Radius*math.Sin(angle)

	// 接線方向の初期速度を計算
	tangentVelX := -orbitConfig.Radius * orbitConfig.Speed * math.Sin(angle)
	tangentVelY := orbitConfig.Radius * orbitConfig.Speed * math.Cos(angle)

	sphere := &Sphere{
		Position:     Position{X: nodeX, Y: nodeY},
		Velocity:     Position{X: tangentVelX, Y: tangentVelY},
		Acceleration: Position{X: 0, Y: 0},
		Radius:       utils.SPHERE_RADIUS,
		Mass:         0.5,
	}

	satellite := &Satellite{
		Sphere:    sphere,
		OrbitType: orbitType,
		Angle:     angle,
	}

	c.Satellites = append(c.Satellites, satellite)

	// JSON用のNodesを更新
	c.UpdateNodes()
}

// handleSphereCollisions は球体間の衝突処理を行う
func (c *Celestial) handleSphereCollisions(deltaTime float64) {
	minDistance := utils.SPHERE_RADIUS * 2.0 // 衝突距離（球同士が接触する距離）

	// 全衛星を取得
	allSatellites := c.GetAllSpheres()

	// 全ノードペアについて衝突をチェック
	for i := 0; i < len(allSatellites); i++ {
		for j := i + 1; j < len(allSatellites); j++ {
			nodeA := allSatellites[i]
			nodeB := allSatellites[j]

			// 距離を計算
			dx := nodeB.Position.X - nodeA.Position.X
			dy := nodeB.Position.Y - nodeA.Position.Y
			distance := math.Sqrt(dx*dx + dy*dy)

			// 衝突している場合
			if distance < minDistance && distance > 0.01 {
				// 正規化された衝突方向
				nx := dx / distance
				ny := dy / distance

				// 相対速度を計算
				dvx := nodeB.Velocity.X - nodeA.Velocity.X
				dvy := nodeB.Velocity.Y - nodeA.Velocity.Y

				// 衝突方向の相対速度
				relativeSpeed := dvx*nx + dvy*ny

				// 離れている場合は衝突処理不要
				if relativeSpeed >= 0 {
					continue
				}

				// 反発係数を適用した衝突応答
				impulse := 2 * relativeSpeed / (nodeA.Mass + nodeB.Mass)

				// 速度を更新（運動量保存）
				nodeA.Velocity.X += impulse * nodeB.Mass * nx * utils.COLLISION_RESTITUTION
				nodeA.Velocity.Y += impulse * nodeB.Mass * ny * utils.COLLISION_RESTITUTION
				nodeB.Velocity.X -= impulse * nodeA.Mass * nx * utils.COLLISION_RESTITUTION
				nodeB.Velocity.Y -= impulse * nodeA.Mass * ny * utils.COLLISION_RESTITUTION

				// 位置の重なりを解消
				overlap := minDistance - distance
				separationRatio := nodeA.Mass / (nodeA.Mass + nodeB.Mass)
				nodeA.Position.X -= nx * overlap * (1 - separationRatio)
				nodeA.Position.Y -= ny * overlap * (1 - separationRatio)
				nodeB.Position.X += nx * overlap * separationRatio
				nodeB.Position.Y += ny * overlap * separationRatio
			}
		}
	}
}
