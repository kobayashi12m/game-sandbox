package models

import (
	"game-sandbox/server/utils"
	"math"
)

// --- 軌道設定 ---

// EnsureOrbitExists は指定された軌道が存在することを保証する
// OrbitConfig と Satellites 配列の両方を初期化する
func (c *Celestial) EnsureOrbitExists(orbitIndex int) {
	// 軌道設定が存在しない場合は作成
	if _, exists := c.OrbitConfigs[orbitIndex]; !exists {
		radius := utils.SPHERE_RADIUS * utils.ORBITAL_RADIUS_RATIO * float64(orbitIndex+1)
		speed := utils.ORBITAL_SPEED / math.Sqrt(float64(orbitIndex+1))
		c.OrbitConfigs[orbitIndex] = &OrbitConfig{
			Radius: radius,
			Speed:  speed,
		}
	}

	// Satellites 配列を拡張
	for len(c.Satellites) <= orbitIndex {
		c.Satellites = append(c.Satellites, []*Satellite{})
	}
}

// GetOrbitConfig は指定された軌道の設定を返す
func (c *Celestial) GetOrbitConfig(orbitIndex int) *OrbitConfig {
	config, exists := c.OrbitConfigs[orbitIndex]
	if !exists {
		// デフォルト設定を返す
		return &OrbitConfig{
			Radius: utils.SPHERE_RADIUS * utils.ORBITAL_RADIUS_RATIO,
			Speed:  utils.ORBITAL_SPEED,
		}
	}
	return config
}

// GetMaxSatellitesForOrbit は指定された軌道に配置可能な最大衛星数を返す
// 原子の電子殻モデルに基づく: 0層目=2, 1層目=8, 2層目=18, 3層目=32... (2n²)
func GetMaxSatellitesForOrbit(orbitIndex int) int {
	n := orbitIndex + 1
	return 2 * n * n
}

// --- 軌道クエリ ---

// IsOrbitFull は指定された軌道が満杯かどうかを返す
func (c *Celestial) IsOrbitFull(orbitIndex int) bool {
	if orbitIndex < 0 || orbitIndex >= len(c.Satellites) {
		return false
	}
	currentCount := len(c.Satellites[orbitIndex])
	maxCount := GetMaxSatellitesForOrbit(orbitIndex)
	return currentCount >= maxCount
}

// GetAvailableOrbitForNewSatellite は新しい衛星を追加可能な最も内側の軌道番号を返す
func (c *Celestial) GetAvailableOrbitForNewSatellite() int {
	orbitIndex := 0
	maxOrbits := 9 // 最大10層まで（安全装置） - 0から9まで
	for orbitIndex <= maxOrbits {
		if !c.IsOrbitFull(orbitIndex) {
			return orbitIndex
		}
		orbitIndex++
	}
	// 万が一すべて満杯の場合は最後の軌道を返す
	return maxOrbits
}

// AreAllOrbitsFullUpToLayer は指定した層まで全ての軌道が満杯かどうかをチェックする
func (c *Celestial) AreAllOrbitsFullUpToLayer(maxLayer int) bool {
	for i := 0; i <= maxLayer; i++ {
		if !c.IsOrbitFull(i) {
			return false
		}
	}
	return true
}

// GetOutermostOrbitWithSatellites は最外殻の軌道番号とその衛星を返す
// 衛星がない場合は -1 と空配列を返す
func (c *Celestial) GetOutermostOrbitWithSatellites() (int, []*Satellite) {
	for i := len(c.Satellites) - 1; i >= 0; i-- {
		if len(c.Satellites[i]) > 0 {
			return i, c.Satellites[i]
		}
	}
	return -1, []*Satellite{}
}

// GetOutermostOrbitRadius は最外殻軌道の半径を返す
// 衛星がない場合はコアの半径を返す
func (c *Celestial) GetOutermostOrbitRadius() float64 {
	outermostOrbit, _ := c.GetOutermostOrbitWithSatellites()
	if outermostOrbit < 0 {
		// 衛星がない場合はコアの半径を返す
		return c.Core.Radius
	}

	// 最外殻軌道の設定を取得
	orbitConfig := c.GetOrbitConfig(outermostOrbit)
	return orbitConfig.Radius
}

// --- 衛星クエリ ---

// GetAllSpheres はすべての衛星の球体を配列で返す
func (c *Celestial) GetAllSpheres() []*Sphere {
	var spheres []*Sphere
	for _, orbit := range c.Satellites {
		for _, satellite := range orbit {
			spheres = append(spheres, satellite.Sphere)
		}
	}
	return spheres
}

// GetTotalSatelliteCount はすべての衛星の総数を返す
func (c *Celestial) GetTotalSatelliteCount() int {
	count := 0
	for _, orbit := range c.Satellites {
		count += len(orbit)
	}
	return count
}

// FindClosestSatellite は指定位置に最も近い衛星とそのインデックスを返す
// 衛星がない場合は nil, -1 を返す
func FindClosestSatellite(satellites []*Satellite, targetX, targetY float64) (*Satellite, int) {
	if len(satellites) == 0 {
		return nil, -1
	}

	var closest *Satellite
	closestIndex := -1
	minDistSq := math.MaxFloat64

	for i, sat := range satellites {
		dx := sat.Sphere.Position.X - targetX
		dy := sat.Sphere.Position.Y - targetY
		distSq := dx*dx + dy*dy

		if distSq < minDistSq {
			minDistSq = distSq
			closest = sat
			closestIndex = i
		}
	}

	return closest, closestIndex
}

// IsCore は指定した球体がこのCelestialのコアかどうかを判定する
func (c *Celestial) IsCore(sphere *Sphere) bool {
	return sphere == c.Core
}

// --- 角度計算 ---

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

// findBestInsertionAngle は既存の衛星の間で最大の空きスペースの中心角度を返す
func (c *Celestial) findBestInsertionAngle(orbitIndex int) float64 {
	if orbitIndex < 0 || orbitIndex >= len(c.Satellites) {
		return 0
	}
	satellites := c.Satellites[orbitIndex]
	if len(satellites) == 0 {
		return 0
	}

	// 衛星の角度をソート
	angles := make([]float64, len(satellites))
	for i, sat := range satellites {
		angles[i] = c.normalizeAngle(sat.Angle)
	}

	// ソート
	for i := 0; i < len(angles)-1; i++ {
		for j := i + 1; j < len(angles); j++ {
			if angles[i] > angles[j] {
				angles[i], angles[j] = angles[j], angles[i]
			}
		}
	}

	// 最大の空きスペースを見つける
	maxGap := 0.0
	bestAngle := 0.0

	for i := 0; i < len(angles); i++ {
		var gap float64
		if i == len(angles)-1 {
			// 最後の衛星から最初の衛星へのギャップ
			gap = (2.0*math.Pi - angles[i]) + angles[0]
		} else {
			gap = angles[i+1] - angles[i]
		}

		if gap > maxGap {
			maxGap = gap
			bestAngle = c.normalizeAngle(angles[i] + gap/2.0)
		}
	}

	return bestAngle
}

// --- 角度補正 ---

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
