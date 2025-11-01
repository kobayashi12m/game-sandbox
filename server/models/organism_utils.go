package models

import (
	"game-sandbox/server/utils"
	"math"
)

// GetAllSpheres はすべての衛星の球体を配列で返す
func (c *Celestial) GetAllSpheres() []*Sphere {
	var spheres []*Sphere
	for _, satellite := range c.Satellites {
		spheres = append(spheres, satellite.Sphere)
	}
	return spheres
}

// GetTotalSatelliteCount はすべての衛星の総数を返す
func (c *Celestial) GetTotalSatelliteCount() int {
	return len(c.Satellites)
}

// RemoveSatellite は指定されたインデックスの衛星を削除する
func (c *Celestial) RemoveSatellite(targetIndex int) bool {
	if targetIndex < 0 || targetIndex >= len(c.Satellites) {
		return false
	}
	c.Satellites = append(c.Satellites[:targetIndex], c.Satellites[targetIndex+1:]...)
	return true
}

// GetOrbitConfig は指定された軌道の設定を返す
func (c *Celestial) GetOrbitConfig(orbitType int) *OrbitConfig {
	config, exists := c.OrbitConfigs[orbitType]
	if !exists {
		// デフォルト設定を返す
		return &OrbitConfig{
			Radius: utils.SPHERE_RADIUS * utils.ORBITAL_RADIUS_RATIO,
			Speed:  utils.ORBITAL_SPEED,
		}
	}
	return config
}

// AddOrbitType は新しい軌道タイプを追加する
func (c *Celestial) AddOrbitType(orbitType int, radius, speed float64) {
	c.OrbitConfigs[orbitType] = &OrbitConfig{
		Radius: radius,
		Speed:  speed,
	}
}

// GetMaxSatellitesForOrbit は指定された軌道に配置可能な最大衛星数を返す
// 原子の電子殻モデルに基づく: 1層目=2, 2層目=8, 3層目=18, 4層目=32... (2n²)
func GetMaxSatellitesForOrbit(orbitType int) int {
	return 2 * orbitType * orbitType
}

// GetSatellitesInOrbit は指定された軌道にある衛星のリストを返す
func (c *Celestial) GetSatellitesInOrbit(orbitType int) []*Satellite {
	var satellites []*Satellite
	for _, sat := range c.Satellites {
		if sat.OrbitType == orbitType {
			satellites = append(satellites, sat)
		}
	}
	return satellites
}

// GetHighestOrbitType は現在使用中の最高軌道番号を返す
func (c *Celestial) GetHighestOrbitType() int {
	maxOrbit := 0
	for _, sat := range c.Satellites {
		if sat.OrbitType > maxOrbit {
			maxOrbit = sat.OrbitType
		}
	}
	return maxOrbit
}

// IsOrbitFull は指定された軌道が満杯かどうかを返す
func (c *Celestial) IsOrbitFull(orbitType int) bool {
	currentCount := len(c.GetSatellitesInOrbit(orbitType))
	maxCount := GetMaxSatellitesForOrbit(orbitType)
	return currentCount >= maxCount
}

// GetAvailableOrbitForNewSatellite は新しい衛星を追加可能な最も内側の軌道番号を返す
func (c *Celestial) GetAvailableOrbitForNewSatellite() int {
	orbitType := 1
	maxOrbits := 10 // 最大10層まで（安全装置）
	for orbitType <= maxOrbits {
		if !c.IsOrbitFull(orbitType) {
			return orbitType
		}
		orbitType++
	}
	// 万が一すべて満杯の場合は最後の軌道を返す
	return maxOrbits
}

// RebalanceSatellitesInOrbit は指定された軌道の衛星を等間隔に再配置する
func (c *Celestial) RebalanceSatellitesInOrbit(orbitType int) {
	satellites := c.GetSatellitesInOrbit(orbitType)
	count := len(satellites)
	if count == 0 {
		return
	}
	
	for i, sat := range satellites {
		sat.Angle = float64(i) * 2.0 * math.Pi / float64(count)
	}
}

// findBestInsertionAngle は既存の衛星の間で最大の空きスペースの中心角度を返す
func (c *Celestial) findBestInsertionAngle(orbitType int) float64 {
	satellites := c.GetSatellitesInOrbit(orbitType)
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
