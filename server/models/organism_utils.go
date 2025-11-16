package models

import (
	"game-sandbox/server/utils"
	"math"
)

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

// RemoveSatellite は指定された軌道とインデックスの衛星を削除する
func (c *Celestial) RemoveSatellite(orbitIndex, satIndex int) bool {
	if orbitIndex < 0 || orbitIndex >= len(c.Satellites) {
		return false
	}
	if satIndex < 0 || satIndex >= len(c.Satellites[orbitIndex]) {
		return false
	}
	c.Satellites[orbitIndex] = append(c.Satellites[orbitIndex][:satIndex], c.Satellites[orbitIndex][satIndex+1:]...)
	return true
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

// AddorbitIndex は新しい軌道タイプを追加する
func (c *Celestial) AddorbitIndex(orbitIndex int, radius, speed float64) {
	c.OrbitConfigs[orbitIndex] = &OrbitConfig{
		Radius: radius,
		Speed:  speed,
	}
}

// GetMaxSatellitesForOrbit は指定された軌道に配置可能な最大衛星数を返す
// 原子の電子殻モデルに基づく: 0層目=2, 1層目=8, 2層目=18, 3層目=32... (2n²)
func GetMaxSatellitesForOrbit(orbitIndex int) int {
	n := orbitIndex + 1
	return 2 * n * n
}

// GetSatellitesInOrbit は指定された軌道にある衛星のリストを返す
// orbitIndexは0から始まる（第0軌道=0、第1軌道=1...）
func (c *Celestial) GetSatellitesInOrbit(orbitIndex int) []*Satellite {
	if orbitIndex < 0 || orbitIndex >= len(c.Satellites) {
		return []*Satellite{}
	}
	return c.Satellites[orbitIndex]
}

// GetHighestorbitIndex は現在使用中の最高軌道番号を返す
func (c *Celestial) GetHighestorbitIndex() int {
	// 最外側から内側に向かって、衛星が存在する軌道を探す
	for i := len(c.Satellites) - 1; i >= 0; i-- {
		if len(c.Satellites[i]) > 0 {
			return i
		}
	}
	return -1 // 衛星が存在しない場合
}

// IsOrbitFull は指定された軌道が満杯かどうかを返す
func (c *Celestial) IsOrbitFull(orbitIndex int) bool {
	currentCount := len(c.GetSatellitesInOrbit(orbitIndex))
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

// RebalanceSatellitesInOrbit は指定された軌道の衛星を等間隔に再配置する
func (c *Celestial) RebalanceSatellitesInOrbit(orbitIndex int) {
	satellites := c.GetSatellitesInOrbit(orbitIndex)
	count := len(satellites)
	if count == 0 {
		return
	}

	for i, sat := range satellites {
		sat.Angle = float64(i) * 2.0 * math.Pi / float64(count)
	}
}

// findBestInsertionAngle は既存の衛星の間で最大の空きスペースの中心角度を返す
func (c *Celestial) findBestInsertionAngle(orbitIndex int) float64 {
	satellites := c.GetSatellitesInOrbit(orbitIndex)
	if len(satellites) == 0 {
		return 0
	}

	// 衛星の角度をソート
	angles := make([]float64, len(satellites))
	for i, sat := range satellites {
		angles[i] = sat.Angle
		// 角度を0-2πの範囲に正規化
		for angles[i] < 0 {
			angles[i] += 2.0 * math.Pi
		}
		for angles[i] >= 2.0*math.Pi {
			angles[i] -= 2.0 * math.Pi
		}
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
		nextIndex := (i + 1) % len(angles)
		var gap float64
		if i == len(angles)-1 {
			// 最後の衛星から最初の衛星へのギャップ
			gap = (2.0*math.Pi - angles[i]) + angles[0]
		} else {
			gap = angles[nextIndex] - angles[i]
		}

		if gap > maxGap {
			maxGap = gap
			if i == len(angles)-1 {
				// 最後の衛星から最初の衛星へのギャップの中心
				bestAngle = angles[i] + gap/2.0
				if bestAngle >= 2.0*math.Pi {
					bestAngle -= 2.0 * math.Pi
				}
			} else {
				bestAngle = angles[i] + gap/2.0
			}
		}
	}

	return bestAngle
}
