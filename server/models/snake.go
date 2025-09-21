package models

import (
	"chess-mmo/server/utils"
	"math/rand/v2"
)

// Reset は蛇を初期状態に初期化する
func (s *Snake) Reset() {
	startX := rand.IntN(utils.GRID_SIZE-10) + 5
	startY := rand.IntN(utils.GRID_SIZE-10) + 5
	s.Body = []Position{
		{X: startX, Y: startY},
		{X: startX - 1, Y: startY},
		{X: startX - 2, Y: startY},
	}
	s.Direction = utils.DIRECTIONS["RIGHT"]
	s.Growing = 0
	s.Alive = true
	s.Respawning = false
}

// Move は蛇を現在の方向に1マス進める
func (s *Snake) Move() {
	if !s.Alive {
		return
	}

	head := s.Body[0]
	newHead := Position{
		X: head.X + s.Direction.X,
		Y: head.Y + s.Direction.Y,
	}

	// 端でのラップアラウンド
	if newHead.X < 0 {
		newHead.X = utils.GRID_SIZE - 1
	} else if newHead.X >= utils.GRID_SIZE {
		newHead.X = 0
	}
	if newHead.Y < 0 {
		newHead.Y = utils.GRID_SIZE - 1
	} else if newHead.Y >= utils.GRID_SIZE {
		newHead.Y = 0
	}

	s.Body = append([]Position{newHead}, s.Body...)

	if s.Growing > 0 {
		s.Growing--
	} else {
		s.Body = s.Body[:len(s.Body)-1]
	}
}

// Grow は蛇の成長カウンターを増加する
func (s *Snake) Grow(amount int) {
	s.Growing += amount
}

// CheckSelfCollision は蛇の頭が体と衝突した場合trueを返す
func (s *Snake) CheckSelfCollision() bool {
	head := s.Body[0]
	for i := 1; i < len(s.Body); i++ {
		if head.X == s.Body[i].X && head.Y == s.Body[i].Y {
			return true
		}
	}
	return false
}

// CheckCollisionWith はこの蛇の頭が他の蛇と衝突した場合trueを返す
func (s *Snake) CheckCollisionWith(other *Snake) bool {
	if !other.Alive {
		return false
	}
	head := s.Body[0]
	for _, segment := range other.Body {
		if head.X == segment.X && head.Y == segment.Y {
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
