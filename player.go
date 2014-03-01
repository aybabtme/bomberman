package main

import (
	"math/rand"
)

type PlayerState struct {
	Name                     string
	X, Y                     int
	Bombs, MaxBomb, MaxRange int
	Alive                    bool
	CurBoard                 Board
}

type PlayerMove string

const (
	Up      = PlayerMove("up")
	Down    = PlayerMove("down")
	Left    = PlayerMove("left")
	Right   = PlayerMove("right")
	PutBomb = PlayerMove("bomb")
)

type Player interface {
	Name() string
	Move() <-chan PlayerMove
	Update() chan<- PlayerState
}

type InputPlayer struct {
	state PlayerState

	// Comms
	update  chan PlayerState
	inMove  <-chan PlayerMove
	outMove chan PlayerMove
}

func NewInputPlayer(state PlayerState, input <-chan PlayerMove) Player {
	i := &InputPlayer{
		state:   state,
		update:  make(chan PlayerState),
		inMove:  input,
		outMove: make(chan PlayerMove, 1), // Rate-limiting to 1 move per turn
	}

	go func() {
		for i.state.Alive {
			select {
			case move := <-i.inMove:
				i.forwardMove(move)
			case i.state = <-i.update:
				log.Debugf("[%s] New state.", i.Name())
			}
		}
	}()

	return i
}

func (i *InputPlayer) forwardMove(move PlayerMove) {
	select {
	case i.outMove <- move:
	default:
		// Drop it
	}
}

func (i *InputPlayer) Name() string {
	return i.state.Name
}

func (i *InputPlayer) Move() <-chan PlayerMove {
	return i.outMove
}

func (i *InputPlayer) Update() chan<- PlayerState {
	return i.update
}

///////////////////////
// AI

type RandomPlayer struct {
	state   PlayerState
	update  chan PlayerState
	outMove chan PlayerMove
}

func NewRandomPlayer(state PlayerState, seed int64) *RandomPlayer {
	r := &RandomPlayer{
		state:   state,
		update:  make(chan PlayerState),
		outMove: make(chan PlayerMove, 1),
	}

	go func() {
		rnd := rand.New(rand.NewSource(seed))
		for {
			var m PlayerMove
			switch rnd.Intn(5) {
			case 0:
				m = Up
			case 1:
				m = Down
			case 2:
				m = Left
			case 3:
				m = Right
			case 4:
				m = PutBomb
			}
			log.Debugf("[%s] Move: %s", r.state.Name, m)
			r.outMove <- m
		}
	}()

	return r
}

func (i *RandomPlayer) Name() string {
	return i.state.Name
}

func (i *RandomPlayer) Move() <-chan PlayerMove {
	return i.outMove
}

func (i *RandomPlayer) Update() chan<- PlayerState {
	return i.update
}
