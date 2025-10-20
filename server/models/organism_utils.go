package models

import "game-sandbox/server/utils"

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
