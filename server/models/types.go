package models

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
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
	Position Position `json:"p"`  // position → p
	Radius   float64  `json:"r"`  // radius → r
}

// MarshalJSON はDroppedSatelliteを配列形式でJSON化 [position, radius]
func (d DroppedSatellite) MarshalJSON() ([]byte, error) {
	posJSON, err := d.Position.MarshalJSON()
	if err != nil {
		return nil, err
	}
	
	// 配列形式: [position, radius]
	result := fmt.Sprintf(`[%s,%g]`, string(posJSON), d.Radius)
	return []byte(result), nil
}

// Player はゲーム内のプレイヤーを表す
type Player struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	Celestial           *Celestial `json:"celestial"`
	Score               int        `json:"score"`
	Conn                *websocket.Conn
	IsNPC               bool       `json:"-"` // NPCかどうかのフラグ
	LastDirectionChange time.Time  `json:"-"` // 最後に方向を変えた時刻
	ConnMu              sync.Mutex `json:"-"` // WebSocket書き込み用mutex
}

// Projectile は射出された衛星を表す
type Projectile struct {
	ID       string   `json:"id"`
	Sphere   *Sphere  `json:"sph"`     // sphere → sph
	OwnerID  string   `json:"oid"`     // ownerId → oid
	Lifetime float64  `json:"-"` // 残り寿命（秒）
}

// MarshalJSON はProjectileを配列形式でJSON化 [id, sphere, ownerId]
func (p Projectile) MarshalJSON() ([]byte, error) {
	sphereJSON, err := p.Sphere.MarshalJSON()
	if err != nil {
		return nil, err
	}
	
	// 配列形式: [id, sphere, ownerId]
	// ID と OwnerID を適切にエスケープ
	escapedID := fmt.Sprintf("%q", p.ID)
	escapedOwnerID := fmt.Sprintf("%q", p.OwnerID)
	result := fmt.Sprintf(`[%s,%s,%s]`,
		escapedID, string(sphereJSON), escapedOwnerID)
	
	return []byte(result), nil
}

// GameState はクライアントに送信される現在の状態を表す
type GameState struct {
	Players           []PlayerState      `json:"pls"`      // players → pls
	DroppedSatellites []DroppedSatellite `json:"ds"`       // droppedSatellites → ds
	Projectiles       []Projectile       `json:"proj"`     // projectiles → proj
}

// GridLine はSpatialGridの可視化用の線を表す
type GridLine struct {
	StartX float64 `json:"startX"`
	StartY float64 `json:"startY"`
	EndX   float64 `json:"endX"`
	EndY   float64 `json:"endY"`
}

// PlayerState はクライアント同期用のプレイヤーデータを表す
type PlayerState struct {
	ID        string     `json:"id"`
	Name      string     `json:"nm"`          // name → nm
	Celestial *Celestial `json:"cel"`         // celestial → cel
	Score     int        `json:"sc"`          // score → sc
}

// MarshalJSON はPlayerStateを配列形式でJSON化 [id, name, celestial, score]
func (p PlayerState) MarshalJSON() ([]byte, error) {
	celestialJSON, err := p.Celestial.MarshalJSON()
	if err != nil {
		return nil, err
	}
	
	// 配列形式: [id, name, celestial, score]
	// ID と Name を適切にエスケープ
	escapedID := fmt.Sprintf("%q", p.ID)
	escapedName := fmt.Sprintf("%q", p.Name)
	result := fmt.Sprintf(`[%s,%s,%s,%d]`,
		escapedID, escapedName, string(celestialJSON), p.Score)
	
	return []byte(result), nil
}

// ScoreInfo はスコアボード用の軽量プレイヤー情報を表す
type ScoreInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
	Alive bool   `json:"alive"`
	Color string `json:"color"`
}

// ScoreUpdate はスコアボードの更新情報を表す（さらに軽量化）
type ScoreUpdate struct {
	Players []ScoreInfo `json:"players"`
}

// GameConfig はゲームの設定を表す
type GameConfig struct {
	FieldWidth    float64    `json:"fieldWidth"`
	FieldHeight   float64    `json:"fieldHeight"`
	SphereRadius  float64    `json:"sphereRadius"`
	CullingWidth  float64    `json:"cullingWidth"`
	CullingHeight float64    `json:"cullingHeight"`
	GridLines     []GridLine `json:"gridLines,omitempty"` // SpatialGrid可視化用
}
