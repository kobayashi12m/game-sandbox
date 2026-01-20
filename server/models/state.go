package models

import (
	"fmt"
	"game-sandbox/server/celestial"
	"game-sandbox/server/types"
)

// DroppedSatellite は落ちた衛星を表す
type DroppedSatellite struct {
	Position       types.Position `json:"p"` // position → p
	Radius         float64        `json:"r"` // radius → r
	Color          string         `json:"c"` // color → c
	IsOriginalCore bool           `json:"-"` // 元コアかどうか（JSONに含めない）
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
	ID       string            `json:"id"`
	Sphere   *celestial.Sphere `json:"sph"` // sphere → sph
	Owner    *Player           `json:"-"`   // オーナーへの参照
	Lifetime float64           `json:"-"`   // 残り寿命（秒）
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
	Players           []PlayerState        `json:"pls"`                // players → pls
	DroppedSatellites []DroppedSatellite   `json:"ds"`                 // droppedSatellites → ds
	Projectiles       []Projectile         `json:"proj"`               // projectiles → proj
	NPCDebug          *types.NPCDebugStats `json:"npcDebug,omitempty"` // NPCデバッグ情報
}
