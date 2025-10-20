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
	Velocity     Position `json:"velocity"`
	Acceleration Position `json:"-"` // 加速度
	Radius       float64  `json:"radius"`
	Mass         float64  `json:"-"` // 質量
}

// Orbit は軌道を表す
type Orbit struct {
	Node         *Sphere `json:"-"`            // 球体（JSON送信不要）
	Angle        float64 `json:"angle"`        // 現在の角度（ラジアン）
	OrbitalSpeed float64 `json:"orbitalSpeed"` // 軌道速度（ラジアン/秒）
	Radius       float64 `json:"radius"`       // 軌道半径
}

// Celestial は核と衛星からなる天体システムを表す
type Celestial struct {
	Core       *Sphere   `json:"core"`       // 中心球
	Satellites []*Sphere `json:"nodes"`      // 周辺球
	Orbits     []*Orbit  `json:"satellites"` // 衛星情報
	Color      string    `json:"color"`
	Alive      bool      `json:"alive"`
	Growing    int       `json:"-"`
	Respawning bool      `json:"-"`
	DeathTime  time.Time `json:"-"`

	// 移動システム用
	MaxSpeed    float64 `json:"-"` // 最大速度
	AccelForce  float64 `json:"-"` // 加速力
	InputActive bool    `json:"-"` // 入力が有効かどうか
}

// Reset は天体システムを初期状態に初期化する
func (o *Celestial) Reset() {
	// フィールド内のランダムな位置にスポーン
	startX := rand.Float64()*(utils.FIELD_WIDTH-100) + 50
	startY := rand.Float64()*(utils.FIELD_HEIGHT-100) + 50

	// コア（中心球）を初期化
	o.Core = &Sphere{
		Position:     Position{X: startX, Y: startY},
		Velocity:     Position{X: 0, Y: 0},
		Acceleration: Position{X: 0, Y: 0},
		Radius:       utils.SPHERE_RADIUS,
		Mass:         1.0,
	}

	// 初期ノード（衛星構造：コアの周りに4個を軌道上に配置）
	o.Satellites = []*Sphere{}
	o.Orbits = []*Orbit{}

	// コアの周りに4個のノードを軌道上に配置
	nodeCount := 4
	orbitalRadius := utils.SPHERE_RADIUS * utils.ORBITAL_RADIUS_RATIO

	for i := 0; i < nodeCount; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(nodeCount) // 等間隔で配置

		nodeX := startX + orbitalRadius*math.Cos(angle)
		nodeY := startY + orbitalRadius*math.Sin(angle)

		// 軌道接線方向の初期速度を計算
		tangentVelX := -orbitalRadius * utils.ORBITAL_SPEED * math.Sin(angle)
		tangentVelY := orbitalRadius * utils.ORBITAL_SPEED * math.Cos(angle)

		node := &Sphere{
			Position:     Position{X: nodeX, Y: nodeY},
			Velocity:     Position{X: tangentVelX, Y: tangentVelY},
			Acceleration: Position{X: 0, Y: 0},
			Radius:       utils.SPHERE_RADIUS,
			Mass:         0.5, // 衛星はコアより軽い
		}

		// 軌道情報を作成
		satellite := &Orbit{
			Node:         node,
			Angle:        angle,
			OrbitalSpeed: utils.ORBITAL_SPEED,
			Radius:       orbitalRadius,
		}

		o.Satellites = append(o.Satellites, node)
		o.Orbits = append(o.Orbits, satellite)
	}

	o.Growing = 0
	o.Alive = true
	o.Respawning = false

	// 天体システムの移動パラメータ
	o.MaxSpeed = utils.CELESTIAL_SPEED
	o.AccelForce = utils.CELESTIAL_ACCEL_FORCE // 加速力
	o.InputActive = false
}

// UpdateMotion は天体システムの運動を更新する
func (o *Celestial) UpdateMotion(deltaTime float64) {
	if !o.Alive {
		return
	}

	// 加速度ベースの移動システム
	if o.InputActive {
		// キーが押されている間は加速度を適用
		o.Core.Velocity.X += o.Core.Acceleration.X * deltaTime
		o.Core.Velocity.Y += o.Core.Acceleration.Y * deltaTime
	}

	// 最大速度制限
	speed := math.Sqrt(o.Core.Velocity.X*o.Core.Velocity.X + o.Core.Velocity.Y*o.Core.Velocity.Y)
	if speed > o.MaxSpeed {
		o.Core.Velocity.X = (o.Core.Velocity.X / speed) * o.MaxSpeed
		o.Core.Velocity.Y = (o.Core.Velocity.Y / speed) * o.MaxSpeed
	}

	// 全衛星の軌道運動シミュレーション
	o.updateOrbitalMotion(deltaTime)

	// フィールド境界での衝突処理
	o.applyBoundaryCollision()
}

// updateOrbitalMotion は軌道運動シミュレーションを実行する
func (o *Celestial) updateOrbitalMotion(deltaTime float64) {
	// コアの速度と位置を更新
	o.Core.Velocity.X += o.Core.Acceleration.X * deltaTime
	o.Core.Velocity.Y += o.Core.Acceleration.Y * deltaTime
	o.Core.Position.X += o.Core.Velocity.X * deltaTime
	o.Core.Position.Y += o.Core.Velocity.Y * deltaTime

	// 各衛星の軌道を更新（核の動きとは独立）
	for i, satellite := range o.Orbits {
		node := satellite.Node

		// 軌道角度を更新（一定速度で回転）
		satellite.Angle += satellite.OrbitalSpeed * deltaTime

		// 理想的な軌道位置を計算
		idealX := o.Core.Position.X + satellite.Radius*math.Cos(satellite.Angle)
		idealY := o.Core.Position.Y + satellite.Radius*math.Sin(satellite.Angle)

		// 現在位置から理想位置への差
		dx := idealX - node.Position.X
		dy := idealY - node.Position.Y

		// スムーズに理想位置に移動（強めの補正）
		node.Position.X += dx * utils.ORBITAL_CORRECTION_STRENGTH
		node.Position.Y += dy * utils.ORBITAL_CORRECTION_STRENGTH

		// 軌道速度を計算（接線方向）
		tangentX := -math.Sin(satellite.Angle) * satellite.Radius * satellite.OrbitalSpeed
		tangentY := math.Cos(satellite.Angle) * satellite.Radius * satellite.OrbitalSpeed

		// 核の速度を継承（核と一緒に移動する感じを出す）
		node.Velocity.X = tangentX + o.Core.Velocity.X*utils.ORBITAL_VELOCITY_INHERITANCE
		node.Velocity.Y = tangentY + o.Core.Velocity.Y*utils.ORBITAL_VELOCITY_INHERITANCE

		o.Satellites[i] = node
	}

	// 球体間の衝突処理
	o.handleSphereCollisions(deltaTime)

	// 空気抵抗を適用
	o.Core.Velocity.X *= utils.AIR_RESISTANCE
	o.Core.Velocity.Y *= utils.AIR_RESISTANCE

	// 低速時の停止判定
	if !o.InputActive {
		speed := math.Sqrt(o.Core.Velocity.X*o.Core.Velocity.X + o.Core.Velocity.Y*o.Core.Velocity.Y)
		if speed < o.MaxSpeed*utils.STOP_THRESHOLD_RATIO {
			o.Core.Velocity.X = 0
			o.Core.Velocity.Y = 0
		}
	}
}

// applyBoundaryCollision はフィールド境界での衝突処理を適用
func (o *Celestial) applyBoundaryCollision() {
	// コアの境界衝突処理
	if o.Core.Position.X-o.Core.Radius < 0 {
		o.Core.Position.X = o.Core.Radius
		o.Core.Velocity.X = -o.Core.Velocity.X * 0.5 // 反発係数0.5
	} else if o.Core.Position.X+o.Core.Radius >= utils.FIELD_WIDTH {
		o.Core.Position.X = utils.FIELD_WIDTH - o.Core.Radius
		o.Core.Velocity.X = -o.Core.Velocity.X * 0.5
	}
	if o.Core.Position.Y-o.Core.Radius < 0 {
		o.Core.Position.Y = o.Core.Radius
		o.Core.Velocity.Y = -o.Core.Velocity.Y * 0.5
	} else if o.Core.Position.Y+o.Core.Radius >= utils.FIELD_HEIGHT {
		o.Core.Position.Y = utils.FIELD_HEIGHT - o.Core.Radius
		o.Core.Velocity.Y = -o.Core.Velocity.Y * 0.5
	}

	// ノードは軌道上で自動的に動くため、境界衝突処理は不要
}

// SetAcceleration は加速度を直接設定する（360度自由移動用）
func (o *Celestial) SetAcceleration(x, y float64) {
	o.Core.Acceleration.X = x * o.AccelForce
	o.Core.Acceleration.Y = y * o.AccelForce
	o.InputActive = (x != 0 || y != 0)
}

// AddSatellite は新しい衛星を追加する（成長時）
func (o *Celestial) AddSatellite() {
	// 軌道半径を取得
	orbitalRadius := utils.SPHERE_RADIUS * utils.ORBITAL_RADIUS_RATIO

	// ランダムな角度で新しいノードを追加
	angle := rand.Float64() * 2.0 * math.Pi

	coreX := o.Core.Position.X
	coreY := o.Core.Position.Y
	nodeX := coreX + orbitalRadius*math.Cos(angle)
	nodeY := coreY + orbitalRadius*math.Sin(angle)

	// 接線方向の初期速度を計算
	tangentVelX := -orbitalRadius * utils.ORBITAL_SPEED * math.Sin(angle)
	tangentVelY := orbitalRadius * utils.ORBITAL_SPEED * math.Cos(angle)

	newNode := &Sphere{
		Position:     Position{X: nodeX, Y: nodeY},
		Velocity:     Position{X: tangentVelX, Y: tangentVelY},
		Acceleration: Position{X: 0, Y: 0},
		Radius:       utils.SPHERE_RADIUS,
		Mass:         0.5,
	}

	// 新しい軌道ノードを作成
	newSatellite := &Orbit{
		Node:         newNode,
		Angle:        angle,
		OrbitalSpeed: utils.ORBITAL_SPEED,
		Radius:       orbitalRadius,
	}

	o.Satellites = append(o.Satellites, newNode)
	o.Orbits = append(o.Orbits, newSatellite)
}

// handleSphereCollisions は球体間の衝突処理を行う
func (o *Celestial) handleSphereCollisions(deltaTime float64) {
	minDistance := utils.SPHERE_RADIUS * 2.0 // 衝突距離（球同士が接触する距離）

	// 全ノードペアについて衝突をチェック
	for i := 0; i < len(o.Satellites); i++ {
		for j := i + 1; j < len(o.Satellites); j++ {
			nodeA := o.Satellites[i]
			nodeB := o.Satellites[j]

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
