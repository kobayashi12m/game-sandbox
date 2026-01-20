package celestial

import (
	"game-sandbox/server/utils"
	"math"
)

// UpdateMotion は天体システムの運動を更新する
func (c *Celestial) UpdateMotion(deltaTime float64) {
	if !c.Alive {
		return
	}

	// 1. コアの運動更新
	c.updateCoreMotion(deltaTime)

	// 2. 衛星の軌道更新
	c.updateSatelliteOrbits(deltaTime)

	// 3. 衝突処理
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
		speed = c.MaxSpeed // 速度を更新
	}

	// 位置を更新
	c.Core.Position.X += c.Core.Velocity.X * deltaTime
	c.Core.Position.Y += c.Core.Velocity.Y * deltaTime

	// 空気抵抗を適用
	c.Core.Velocity.X *= utils.AIR_RESISTANCE
	c.Core.Velocity.Y *= utils.AIR_RESISTANCE

	// 低速時の停止判定（入力がない時のみ）
	hasInput := c.Core.Acceleration.X != 0 || c.Core.Acceleration.Y != 0
	if !hasInput && speed < c.MaxSpeed*utils.STOP_THRESHOLD_RATIO {
		c.Core.Velocity.X = 0
		c.Core.Velocity.Y = 0
	}
}

// updateSatelliteOrbits は衛星の軌道運動のみを更新する
func (c *Celestial) updateSatelliteOrbits(deltaTime float64) {
	// 各軌道ごとに理想的な角度配置に向けて補正
	c.correctSatelliteAngles(deltaTime)

	for orbitIndex, orbit := range c.Satellites {
		for _, satellite := range orbit {
			// 軌道設定を取得
			orbitConfig := c.GetOrbitConfig(orbitIndex)

			// 角度を更新して正規化（0-2πの範囲に収める）
			satellite.Angle = c.normalizeAngle(satellite.Angle + orbitConfig.Speed*deltaTime)

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

// updateSpeedParameters は衛星数に応じて速度と加速力を更新する
func (c *Celestial) updateSpeedParameters() {
	satelliteCount := c.GetTotalSatelliteCount()

	// 基本値から衛星数分を減算
	c.MaxSpeed = utils.CELESTIAL_SPEED - float64(satelliteCount)*utils.SPEED_REDUCTION_PER_SATELLITE
	c.AccelForce = utils.CELESTIAL_ACCEL_FORCE - float64(satelliteCount)*utils.ACCEL_REDUCTION_PER_SATELLITE

	// 最低限の値を保つ
	if c.MaxSpeed < 50.0 {
		c.MaxSpeed = 50.0
	}
	if c.AccelForce < 50.0 {
		c.AccelForce = 50.0
	}
}
