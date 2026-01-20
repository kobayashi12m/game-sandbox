package celestial

import (
	"fmt"
	"game-sandbox/server/types"
	"game-sandbox/server/utils"
	"time"
)

// Sphere は物理演算される球体を表す
type Sphere struct {
	Position     types.Position `json:"position"`
	Velocity     types.Position `json:"velocity,omitempty"`
	Acceleration types.Position `json:"acceleration,omitempty"` // 加速度
	Radius       float64        `json:"radius"`
	Color        string         `json:"color"` // 球体の色
	Mass         float64        `json:"-"`     // 質量
}

// MarshalJSON は配列形式でJSONサイズを最大削減する [[x,y], radius, color, [vx,vy], [ax,ay]]
func (s Sphere) MarshalJSON() ([]byte, error) {
	// 基本形式: [position, radius, color]
	escapedColor := fmt.Sprintf("%q", s.Color)
	result := fmt.Sprintf(`[[%d,%d],%d,%s`,
		int(s.Position.X), int(s.Position.Y), int(s.Radius), escapedColor)

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

// ResetAtPosition は指定位置で天体システムを初期化する
func (c *Celestial) ResetAtPosition(x, y float64) {
	// コア（中心球）を初期化
	c.Core = &Sphere{
		Position:     types.Position{X: x, Y: y},
		Velocity:     types.Position{X: 0, Y: 0},
		Acceleration: types.Position{X: 0, Y: 0},
		Radius:       utils.SPHERE_RADIUS,
		Color:        c.Color,
		Mass:         1.0,
	}

	// 軌道設定を初期化
	c.OrbitConfigs = map[int]*OrbitConfig{
		0: {
			Radius: utils.SPHERE_RADIUS * utils.ORBITAL_RADIUS_RATIO,
			Speed:  utils.ORBITAL_SPEED,
		},
	}

	c.Satellites = [][]*Satellite{}
	c.Alive = true
	c.Respawning = false

	// 初期衛星を配置
	nodeCount := GetMaxSatellitesForOrbit(0)
	for i := 0; i < nodeCount; i++ {
		c.AddSatellite(c.Core.Color, c.Core.Position)
	}
}
