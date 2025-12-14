package models

import (
	"fmt"
)

// Position はゲームフィールド上の座標を表す（浮動小数点）
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// MarshalJSON は座標を配列形式でJSONシリアライズする [x, y]
func (p Position) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`[%d,%d]`, int(p.X), int(p.Y))), nil
}

// DroppedSatellite は落ちた衛星を表す
type DroppedSatellite struct {
	Position       Position `json:"p"` // position → p
	Radius         float64  `json:"r"` // radius → r
	Color          string   `json:"c"` // color → c
	IsOriginalCore bool     `json:"-"` // 元コアかどうか（JSONに含めない）
}

// MarshalJSON はDroppedSatelliteを配列形式でJSON化 [position, radius, color]
func (d DroppedSatellite) MarshalJSON() ([]byte, error) {
	posJSON, err := d.Position.MarshalJSON()
	if err != nil {
		return nil, err
	}

	// 配列形式: [position, radius, color]
	escapedColor := fmt.Sprintf("%q", d.Color)
	result := fmt.Sprintf(`[%s,%g,%s]`, string(posJSON), d.Radius, escapedColor)
	return []byte(result), nil
}

// Projectile は射出された衛星を表す
type Projectile struct {
	ID       string  `json:"id"`
	Sphere   *Sphere `json:"sph"` // sphere → sph
	Owner    *Player `json:"-"`   // オーナーへの参照
	Lifetime float64 `json:"-"`   // 残り寿命（秒）
}

// MarshalJSON はProjectileを配列形式でJSON化 [id, sphere]
func (p Projectile) MarshalJSON() ([]byte, error) {
	sphereJSON, err := p.Sphere.MarshalJSON()
	if err != nil {
		return nil, err
	}

	// 配列形式: [id, sphere]
	escapedID := fmt.Sprintf("%q", p.ID)
	result := fmt.Sprintf(`[%s,%s]`, escapedID, string(sphereJSON))

	return []byte(result), nil
}

// GameState はクライアントに送信される現在の状態を表す
type GameState struct {
	Players           []PlayerState      `json:"pls"`                // players → pls
	DroppedSatellites []DroppedSatellite `json:"ds"`                 // droppedSatellites → ds
	Projectiles       []Projectile       `json:"proj"`               // projectiles → proj
	NPCDebug          *NPCDebugStats     `json:"npcDebug,omitempty"` // NPCデバッグ情報
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
