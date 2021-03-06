package player

import (
	"github.com/aybabtme/bomberman/cell"
	"time"
)

type State struct {
	Turn                      int
	TurnDuration              time.Duration
	Name                      string
	X, Y, LastX, LastY        int
	Bombs, MaxBomb, MaxRadius int
	Alive                     bool
	Board                     [][]*cell.Exported
	GameObject                cell.GameObject
	Message                   string
}

type Move string

const (
	Up      = Move("up")
	Down    = Move("down")
	Left    = Move("left")
	Right   = Move("right")
	PutBomb = Move("bomb")
)

type Player interface {
	Name() string
	Move() <-chan Move
	Update() chan<- State
}
