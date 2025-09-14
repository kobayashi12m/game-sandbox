package utils

import "time"

const (
	GRID_SIZE     = 40
	INITIAL_SPEED = 100 * time.Millisecond
)

// Direction represents movement direction
type Direction struct {
	X int `json:"x"`
	Y int `json:"y"`
}

var (
	DIRECTIONS = map[string]Direction{
		"UP":    {X: 0, Y: -1},
		"DOWN":  {X: 0, Y: 1},
		"LEFT":  {X: -1, Y: 0},
		"RIGHT": {X: 1, Y: 0},
	}
)