package models

import (
	"chess-mmo/server/utils"
	"math"
	"math/rand/v2"
)

// Reset は蛇を初期状態に初期化する
func (s *Snake) Reset() {
	// フィールド内のランダムな位置にスポーン
	startX := rand.Float64()*(utils.FIELD_WIDTH-100) + 50
	startY := rand.Float64()*(utils.FIELD_HEIGHT-100) + 50

	// 初期体長を設定（連続したセグメント）
	s.Body = []Position{
		{X: startX, Y: startY},
		{X: startX - utils.SNAKE_RADIUS*2, Y: startY},
		{X: startX - utils.SNAKE_RADIUS*4, Y: startY},
	}
	s.Direction = utils.DIRECTIONS["RIGHT"]
	s.Growing = 0
	s.Alive = true
	s.Respawning = false
	s.Speed = utils.SNAKE_SPEED
}

// Move は蛇を現在の方向に移動させる（deltaTime: 秒単位）
func (s *Snake) Move(deltaTime float64) {
	if !s.Alive {
		return
	}

	head := s.Body[0]
	// 速度に基づいて移動距離を計算
	dist := s.Speed * deltaTime
	newHead := Position{
		X: head.X + s.Direction.X*dist,
		Y: head.Y + s.Direction.Y*dist,
	}

	// フィールドの端でのラップアラウンド
	if newHead.X < 0 {
		newHead.X += utils.FIELD_WIDTH
	} else if newHead.X >= utils.FIELD_WIDTH {
		newHead.X -= utils.FIELD_WIDTH
	}
	if newHead.Y < 0 {
		newHead.Y += utils.FIELD_HEIGHT
	} else if newHead.Y >= utils.FIELD_HEIGHT {
		newHead.Y -= utils.FIELD_HEIGHT
	}

	// 体のセグメントを更新
	s.updateBodySegments(newHead)
}

// Grow は蛇の成長カウンターを増加する
func (s *Snake) Grow(amount int) {
	s.Growing += amount
}

// CheckCollisionWith はこの蛇の頭が他の蛇と衝突した場合trueを返す
func (s *Snake) CheckCollisionWith(other *Snake) bool {
	if !other.Alive {
		return false
	}
	head := s.Body[0]
	for _, segment := range other.Body {
		if wrappedDistance(head, segment) < utils.SNAKE_RADIUS*2 {
			return true
		}
	}
	return false
}

// ChangeDirection は蛇の方向を更新し、逆方向を防ぐ
func (s *Snake) ChangeDirection(newDir utils.Direction) {
	// 逆方向を防ぐ
	if s.Direction.X == -newDir.X && s.Direction.Y == -newDir.Y {
		return
	}
	s.Direction = newDir
}

// updateBodySegments は蛇の体のセグメントを更新する
func (s *Snake) updateBodySegments(newHead Position) {
	if len(s.Body) == 0 {
		s.Body = []Position{newHead}
		return
	}

	// 新しい体を作成
	newBody := make([]Position, 0, len(s.Body)+1)
	newBody = append(newBody, newHead)

	// 各セグメントを前のセグメントの位置に向けて移動
	for i := 1; i < len(s.Body); i++ {
		prev := s.Body[i-1]
		curr := s.Body[i]

		// ラップアラウンドを考慮した距離を計算
		dist := wrappedDistance(prev, curr)
		if dist > utils.SNAKE_RADIUS*2 {
			// ラップアラウンドを考慮した方向ベクトルを取得
			dx, dy := wrappedDirection(curr, prev)
			length := math.Sqrt(dx*dx + dy*dy)
			if length > 0 {
				dx /= length
				dy /= length
				// セグメントを適切な距離に移動
				curr.X += dx * (dist - utils.SNAKE_RADIUS*2)
				curr.Y += dy * (dist - utils.SNAKE_RADIUS*2)

				// ラップアラウンドを適用
				if curr.X < 0 {
					curr.X += utils.FIELD_WIDTH
				} else if curr.X >= utils.FIELD_WIDTH {
					curr.X -= utils.FIELD_WIDTH
				}
				if curr.Y < 0 {
					curr.Y += utils.FIELD_HEIGHT
				} else if curr.Y >= utils.FIELD_HEIGHT {
					curr.Y -= utils.FIELD_HEIGHT
				}
			}
		}
		newBody = append(newBody, curr)
	}

	// 成長処理
	if s.Growing > 0 {
		s.Growing--
		// 最後のセグメントを保持
		if len(s.Body) > 0 {
			newBody = append(newBody, s.Body[len(s.Body)-1])
		}
	} else if len(newBody) > len(s.Body) {
		// 通常の移動では尻尾を削除
		newBody = newBody[:len(s.Body)]
	}

	s.Body = newBody
}

// distance は2点間の距離を計算する
func distance(p1, p2 Position) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// wrappedDistance はラップアラウンドを考慮した2点間の距離を計算する
func wrappedDistance(p1, p2 Position) float64 {
	// X軸の最短距離を計算
	dx := math.Abs(p1.X - p2.X)
	if dx > utils.FIELD_WIDTH/2 {
		dx = utils.FIELD_WIDTH - dx
	}

	// Y軸の最短距離を計算
	dy := math.Abs(p1.Y - p2.Y)
	if dy > utils.FIELD_HEIGHT/2 {
		dy = utils.FIELD_HEIGHT - dy
	}

	return math.Sqrt(dx*dx + dy*dy)
}

// wrappedDirection はラップアラウンドを考慮した方向ベクトルを計算する
func wrappedDirection(from, to Position) (float64, float64) {
	// X方向の計算
	dx := to.X - from.X
	if dx > utils.FIELD_WIDTH/2 {
		dx -= utils.FIELD_WIDTH
	} else if dx < -utils.FIELD_WIDTH/2 {
		dx += utils.FIELD_WIDTH
	}

	// Y方向の計算
	dy := to.Y - from.Y
	if dy > utils.FIELD_HEIGHT/2 {
		dy -= utils.FIELD_HEIGHT
	} else if dy < -utils.FIELD_HEIGHT/2 {
		dy += utils.FIELD_HEIGHT
	}

	return dx, dy
}
