package models

import (
	"math/rand/v2"
	"chess-mmo/server/utils"
)

// Reset initializes the snake to its starting state
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
}

// Move advances the snake one step in its current direction
func (s *Snake) Move() {
	if !s.Alive {
		return
	}

	head := s.Body[0]
	newHead := Position{
		X: head.X + s.Direction.X,
		Y: head.Y + s.Direction.Y,
	}

	// Wrap around edges
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

// Grow increases the snake's growth counter
func (s *Snake) Grow(amount int) {
	s.Growing += amount
}

// CheckSelfCollision returns true if the snake's head collides with its body
func (s *Snake) CheckSelfCollision() bool {
	head := s.Body[0]
	for i := 1; i < len(s.Body); i++ {
		if head.X == s.Body[i].X && head.Y == s.Body[i].Y {
			return true
		}
	}
	return false
}

// CheckCollisionWith returns true if this snake's head collides with another snake
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

// ChangeDirection updates the snake's direction, preventing reverse direction
func (s *Snake) ChangeDirection(newDir utils.Direction) {
	// Prevent reverse direction
	if s.Direction.X == -newDir.X && s.Direction.Y == -newDir.Y {
		return
	}
	s.Direction = newDir
}