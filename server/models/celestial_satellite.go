package models

import (
	"game-sandbox/server/utils"
	"math"
	"math/rand/v2"
)

// AddSatellite は新しい衛星を追加する（成長時）
func (c *Celestial) AddSatellite(color string, startPos Position) {
	// 利用可能な最も内側の軌道を取得し、存在を保証
	orbitIndex := c.GetAvailableOrbitForNewSatellite()
	c.EnsureOrbitExists(orbitIndex)

	// 挿入角度を決定
	var angle float64
	if len(c.Satellites[orbitIndex]) == 0 {
		angle = rand.Float64() * 2.0 * math.Pi
	} else {
		angle = c.findBestInsertionAngle(orbitIndex)
	}

	// 衛星を作成して追加
	orbitConfig := c.GetOrbitConfig(orbitIndex)
	satellite := c.createSatellite(startPos, angle, orbitConfig, color)
	c.Satellites[orbitIndex] = append(c.Satellites[orbitIndex], satellite)

	// 速度パラメータを更新
	c.updateSpeedParameters()
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

	// 衛星削除後に速度パラメータを更新
	c.updateSpeedParameters()
	return true
}

// EjectSatelliteWithReturn は指定された方向に最も近い最外殻の衛星を射出し、射出された衛星を返す
func (c *Celestial) EjectSatelliteWithReturn(targetX, targetY float64) *Sphere {
	// 最外殻の軌道と衛星を取得
	outermostOrbit, outermostSatellites := c.GetOutermostOrbitWithSatellites()
	if outermostOrbit < 0 {
		return nil
	}

	// クリック位置に最も近い衛星を見つける
	closestSatellite, closestSatIndex := FindClosestSatellite(outermostSatellites, targetX, targetY)
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
		Color:        closestSatellite.Sphere.Color,
		Mass:         closestSatellite.Sphere.Mass,
	}

	// 衛星リストから削除
	c.RemoveSatellite(outermostOrbit, closestSatIndex)

	return ejectedSphere
}

// createSatellite は指定位置・角度・軌道設定で衛星を作成する
func (c *Celestial) createSatellite(position Position, angle float64, orbitConfig *OrbitConfig, color string) *Satellite {
	// 接線方向の初期速度を計算
	tangentVelX := -orbitConfig.Radius * orbitConfig.Speed * math.Sin(angle)
	tangentVelY := orbitConfig.Radius * orbitConfig.Speed * math.Cos(angle)

	sphere := &Sphere{
		Position:     position,
		Velocity:     Position{X: tangentVelX, Y: tangentVelY},
		Acceleration: Position{X: 0, Y: 0},
		Radius:       utils.SPHERE_RADIUS,
		Color:        color,
		Mass:         0.5,
	}

	return &Satellite{
		Sphere: sphere,
		Angle:  angle,
	}
}
