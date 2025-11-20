package models

import (
	"fmt"
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

// MarshalJSON は配列形式でJSONサイズを最大削減する [[x,y], radius, [vx,vy], [ax,ay]]
func (s Sphere) MarshalJSON() ([]byte, error) {
	// 基本形式: [position, radius]
	result := fmt.Sprintf(`[[%d,%d],%d`,
		int(s.Position.X), int(s.Position.Y), int(s.Radius))

	// velocityがゼロでない場合のみ追加
	vx, vy := int(s.Velocity.X), int(s.Velocity.Y)
	if vx != 0 || vy != 0 {
		result += fmt.Sprintf(`,[%d,%d]`, vx, vy)
	} else {
		// velocityがゼロでもaccelerationがある場合はnullを追加
		ax, ay := int(s.Acceleration.X), int(s.Acceleration.Y)
		if ax != 0 || ay != 0 {
			result += `,null`
		}
	}

	// accelerationがゼロでない場合のみ追加
	ax, ay := int(s.Acceleration.X), int(s.Acceleration.Y)
	if ax != 0 || ay != 0 {
		result += fmt.Sprintf(`,[%d,%d]`, ax, ay)
	}

	result += "]"
	return []byte(result), nil
}

// Satellite は衛星を表す
type Satellite struct {
	Sphere *Sphere `json:"sphere"` // 球体
	Angle  float64 `json:"angle"`  // 軌道上の角度
}

// OrbitConfig は軌道の設定
type OrbitConfig struct {
	Radius float64 // 軌道半径
	Speed  float64 // 回転速度（ラジアン/秒）
}

// Celestial は核と衛星からなる天体システムを表す
type Celestial struct {
	Core       *Sphere   `json:"core"` // 中心球
	Color      string    `json:"color"`
	Alive      bool      `json:"alive"`
	Growing    int       `json:"-"`
	Respawning bool      `json:"-"`
	DeathTime  time.Time `json:"-"`

	// 内部管理用（JSON送信されない）
	Satellites   [][]*Satellite       `json:"-"` // 衛星（インデックス0が最内側軌道）
	MaxSpeed     float64              `json:"-"` // 最大速度
	AccelForce   float64              `json:"-"` // 加速力
	OrbitConfigs map[int]*OrbitConfig `json:"-"` // 軌道設定
}

// MarshalJSON はCelestialを配列形式でJSON化 [core, color, alive, nodes]
func (c *Celestial) MarshalJSON() ([]byte, error) {
	// コアのJSONを手動で生成（Sphereのカスタムマーシャリングを使用）
	coreJSON, err := c.Core.MarshalJSON()
	if err != nil {
		return nil, err
	}

	// ノード配列のJSONを手動で生成
	nodes := c.GetAllSpheres()
	nodesJSON := "["
	for i, node := range nodes {
		if i > 0 {
			nodesJSON += ","
		}
		nodeJSON, err := node.MarshalJSON()
		if err != nil {
			return nil, err
		}
		nodesJSON += string(nodeJSON)
	}
	nodesJSON += "]"

	// 配列形式: [core, color, alive, nodes]
	// カラー文字列を適切にエスケープ
	escapedColor := fmt.Sprintf("%q", c.Color)
	result := fmt.Sprintf(`[%s,%s,%t,%s]`,
		string(coreJSON), escapedColor, c.Alive, nodesJSON)

	return []byte(result), nil
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
		0: {
			Radius: utils.SPHERE_RADIUS * utils.ORBITAL_RADIUS_RATIO,
			Speed:  utils.ORBITAL_SPEED,
		},
	}

	// 初期衛星を配置（第0軌道に2個）
	c.Satellites = [][]*Satellite{}
	firstOrbit := []*Satellite{}
	nodeCount := 2 // 第0軌道は最大2個

	for i := 0; i < nodeCount; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(nodeCount) // 等間隔で配置
		orbitConfig := c.OrbitConfigs[0]

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
			Sphere: sphere,
			Angle:  angle,
		}

		firstOrbit = append(firstOrbit, satellite)
	}
	c.Satellites = append(c.Satellites, firstOrbit)

	c.Growing = 0
	c.Alive = true
	c.Respawning = false

	// 天体システムの移動パラメータ
	c.MaxSpeed = utils.CELESTIAL_SPEED
	c.AccelForce = utils.CELESTIAL_ACCEL_FORCE // 加速力

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

// AddSatellite は新しい衛星を追加する（成長時）
func (c *Celestial) AddSatellite() {
	// 利用可能な最も内側の軌道を取得
	orbitIndex := c.GetAvailableOrbitForNewSatellite()

	// 軌道設定が存在しない場合は作成
	if _, exists := c.OrbitConfigs[orbitIndex]; !exists {
		// 各軌道の半径と速度を計算（外側ほど半径が大きく、速度は遅くなる）
		radius := utils.SPHERE_RADIUS * utils.ORBITAL_RADIUS_RATIO * float64(orbitIndex+1)
		speed := utils.ORBITAL_SPEED / math.Sqrt(float64(orbitIndex+1))
		c.AddorbitIndex(orbitIndex, radius, speed)
	}

	orbitConfig := c.GetOrbitConfig(orbitIndex)

	// 新しい衛星を作成
	// 既存の軌道の流れに合わせて自然に配置
	var existingSatellites []*Satellite
	if orbitIndex >= 0 && orbitIndex < len(c.Satellites) {
		existingSatellites = c.Satellites[orbitIndex]
	}

	var angle float64
	if len(existingSatellites) == 0 {
		// 最初の衛星の場合はランダムな角度
		angle = rand.Float64() * 2.0 * math.Pi
	} else {
		// 既存の衛星の間で最大の空きスペースを見つける
		angle = c.findBestInsertionAngle(orbitIndex)
	}

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
		Sphere: sphere,
		Angle:  angle,
	}

	// 必要に応じて軌道を拡張
	for len(c.Satellites) <= orbitIndex {
		c.Satellites = append(c.Satellites, []*Satellite{})
	}

	c.Satellites[orbitIndex] = append(c.Satellites[orbitIndex], satellite)
}

// EjectSatelliteWithReturn は指定された方向に最も近い最外殻の衛星を射出し、射出された衛星を返す
func (c *Celestial) EjectSatelliteWithReturn(targetX, targetY float64) *Sphere {
	// 最外殻の軌道と衛星を取得
	outermostOrbit, outermostSatellites := c.GetOutermostOrbitWithSatellites()
	if outermostOrbit < 0 || len(outermostSatellites) == 0 {
		return nil
	}

	// クリック位置に最も近い衛星を見つける
	var closestSatellite *Satellite
	var closestSatIndex int
	minDistance := math.MaxFloat64

	for i, sat := range outermostSatellites {
		dx := sat.Sphere.Position.X - targetX
		dy := sat.Sphere.Position.Y - targetY
		dist := dx*dx + dy*dy

		if dist < minDistance {
			minDistance = dist
			closestSatellite = sat
			closestSatIndex = i
		}
	}

	if closestSatellite == nil {
		return nil
	}

	// 射出方向を計算（コアからクリック位置への方向）
	dx := targetX - c.Core.Position.X
	dy := targetY - c.Core.Position.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < 0.001 {
		return nil
	}

	// 方向を正規化
	dirX := dx / dist
	dirY := dy / dist

	// 射出速度を設定
	ejectSpeed := utils.SATELLITE_EJECT_SPEED
	closestSatellite.Sphere.Velocity.X = dirX * ejectSpeed
	closestSatellite.Sphere.Velocity.Y = dirY * ejectSpeed

	// 射出する衛星のコピーを作成
	ejectedSphere := &Sphere{
		Position:     closestSatellite.Sphere.Position,
		Velocity:     closestSatellite.Sphere.Velocity,
		Acceleration: Position{X: 0, Y: 0},
		Radius:       closestSatellite.Sphere.Radius,
		Mass:         closestSatellite.Sphere.Mass,
	}

	// 衛星リストから削除
	c.RemoveSatellite(outermostOrbit, closestSatIndex)

	return ejectedSphere
}

// correctSatelliteAngles は各軌道の衛星を理想的な正多角形の角度に向けて微調整する
func (c *Celestial) correctSatelliteAngles(deltaTime float64) {
	// 各軌道ごとに処理
	for _, orbit := range c.Satellites {
		if len(orbit) > 0 {
			c.correctOrbitAngles(orbit, deltaTime)
		}
	}
}

// correctOrbitAngles は指定された軌道の衛星を理想的な角度に向けて補正する
func (c *Celestial) correctOrbitAngles(satellites []*Satellite, deltaTime float64) {
	count := len(satellites)
	if count <= 1 {
		return
	}

	// 衛星を角度順にソート
	for i := 0; i < len(satellites)-1; i++ {
		for j := i + 1; j < len(satellites); j++ {
			angle1 := c.normalizeAngle(satellites[i].Angle)
			angle2 := c.normalizeAngle(satellites[j].Angle)
			if angle1 > angle2 {
				satellites[i], satellites[j] = satellites[j], satellites[i]
			}
		}
	}

	// 理想的な角度間隔
	idealStep := 2.0 * math.Pi / float64(count)

	// 各衛星を理想的な位置に向けて微調整
	for i, sat := range satellites {
		// 理想的な角度（最初の衛星の位置を基準に等間隔）
		baseAngle := c.normalizeAngle(satellites[0].Angle)
		idealAngle := baseAngle + float64(i)*idealStep
		idealAngle = c.normalizeAngle(idealAngle)

		// 現在の角度との差
		currentAngle := c.normalizeAngle(sat.Angle)
		angleDiff := c.shortestAngleDifference(currentAngle, idealAngle)

		// 補正速度を適用（距離に応じて調整）
		correctionSpeed := utils.ANGLE_CORRECTION_SPEED * deltaTime

		// 距離が小さい場合は補正しない（振動を防ぐ）
		if math.Abs(angleDiff) < 0.01 {
			continue
		}

		// 距離に比例した補正（近づくほど緩やか）
		correction := angleDiff * correctionSpeed
		sat.Angle += correction
	}
}

// normalizeAngle は角度を0-2πの範囲に正規化する
func (c *Celestial) normalizeAngle(angle float64) float64 {
	for angle < 0 {
		angle += 2.0 * math.Pi
	}
	for angle >= 2.0*math.Pi {
		angle -= 2.0 * math.Pi
	}
	return angle
}

// shortestAngleDifference は二つの角度間の最短距離を計算する
func (c *Celestial) shortestAngleDifference(from, to float64) float64 {
	diff := to - from

	// -π から π の範囲に正規化
	for diff > math.Pi {
		diff -= 2.0 * math.Pi
	}
	for diff < -math.Pi {
		diff += 2.0 * math.Pi
	}

	return diff
}
