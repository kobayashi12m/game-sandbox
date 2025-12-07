package models

import (
	"fmt"
	"sync"
	"time"

	"game-sandbox/server/utils"

	"github.com/gorilla/websocket"
)

// Player はゲーム内のプレイヤーを表す
type Player struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	Celestial           *Celestial `json:"celestial"`
	Score               int        `json:"score"`
	Conn                *websocket.Conn
	IsNPC               bool       `json:"-"` // NPCかどうかのフラグ
	LastDirectionChange time.Time  `json:"-"` // 最後に方向を変えた時刻
	LastAutoSatellite   time.Time  `json:"-"` // 最後に自動衛星を追加した時刻
	RespawnTime         time.Time  `json:"-"` // リスポーンした時刻（無敵時間の計算用）
	ConnMu              sync.Mutex `json:"-"` // WebSocket書き込み用mutex
}

// IsInvulnerable はプレイヤーが無敵状態かどうかを返す
func (p *Player) IsInvulnerable() bool {
	return time.Since(p.RespawnTime) < utils.RESPAWN_INVULNERABILITY_TIME
}

// ResetAutoSatelliteTimerIfNeeded は衛星数が上限未満になった場合にタイマーをリセットする
func (p *Player) ResetAutoSatelliteTimerIfNeeded() {
	if !p.Celestial.Alive {
		return
	}

	currentSatelliteCount := p.Celestial.GetTotalSatelliteCount()
	if currentSatelliteCount < utils.MAX_AUTO_SATELLITES {
		p.LastAutoSatellite = time.Now()
		utils.Debug("Auto satellite timer reset", map[string]interface{}{
			"event":           "auto_satellite_timer_reset",
			"player_id":       p.ID,
			"player_name":     p.Name,
			"satellite_count": currentSatelliteCount,
			"max_satellites":  utils.MAX_AUTO_SATELLITES,
		})
	}
}

// PlayerState はクライアント同期用のプレイヤーデータを表す
type PlayerState struct {
	ID           string     `json:"id"`
	Name         string     `json:"nm"`  // name → nm
	Celestial    *Celestial `json:"cel"` // celestial → cel
	Score        int        `json:"sc"`  // score → sc
	Invulnerable bool       `json:"inv"` // invulnerable → inv
}

// MarshalJSON はPlayerStateを配列形式でJSON化 [id, name, celestial, score, invulnerable]
func (p PlayerState) MarshalJSON() ([]byte, error) {
	celestialJSON, err := p.Celestial.MarshalJSON()
	if err != nil {
		return nil, err
	}

	// 配列形式: [id, name, celestial, score, invulnerable]
	// ID と Name を適切にエスケープ
	escapedID := fmt.Sprintf("%q", p.ID)
	escapedName := fmt.Sprintf("%q", p.Name)
	result := fmt.Sprintf(`[%s,%s,%s,%d,%t]`,
		escapedID, escapedName, string(celestialJSON), p.Score, p.Invulnerable)

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
