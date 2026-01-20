package types

import "fmt"

// Position はゲームフィールド上の座標を表す（浮動小数点）
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// MarshalJSON は座標を配列形式でJSONシリアライズする [x, y]
func (p Position) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`[%d,%d]`, int(p.X), int(p.Y))), nil
}

// GridLine はSpatialGridの可視化用の線を表す
type GridLine struct {
	StartX float64 `json:"startX"`
	StartY float64 `json:"startY"`
	EndX   float64 `json:"endX"`
	EndY   float64 `json:"endY"`
}

// GameConfig はゲームの設定を表す
type GameConfig struct {
	FieldWidth      float64    `json:"fieldWidth"`
	FieldHeight     float64    `json:"fieldHeight"`
	SphereRadius    float64    `json:"sphereRadius"`
	CullingWidth    float64    `json:"cullingWidth"`
	CullingHeight   float64    `json:"cullingHeight"`
	CameraZoomScale float64    `json:"cameraZoomScale"`     // カメラの固定ズーム倍率
	GridLines       []GridLine `json:"gridLines,omitempty"` // SpatialGrid可視化用
}

// NPCDebugStats はNPCの状態表示用
type NPCDebugStats struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	VelocityX  float64 `json:"velX"`
	VelocityY  float64 `json:"velY"`
	AccelX     float64 `json:"accelX"`
	AccelY     float64 `json:"accelY"`
	AccelForce float64 `json:"accelForce"`
	MaxSpeed   float64 `json:"maxSpeed"`
	Satellites int     `json:"satellites"`
}
