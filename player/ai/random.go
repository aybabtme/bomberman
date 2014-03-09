package ai

import (
	"github.com/aybabtme/bomberman/player"
	"math/rand"
	"time"
)

type RandomPlayer struct {
	state   player.State
	update  chan player.State
	outMove chan player.Move
}

func NewRandomPlayer(state player.State, seed int64) player.Player {
	r := &RandomPlayer{
		state:   state,
		update:  make(chan player.State),
		outMove: make(chan player.Move, 1),
	}

	go func() {
		rnd := rand.New(rand.NewSource(seed))
		for {
			var m player.Move
			switch n := rnd.Intn(10); n {
			case 0:
				m = player.Up
			case 1:
				m = player.Down
			case 2:
				m = player.Left
			case 3:
				m = player.Right
			case 4:
				m = player.PutBomb
			default:
				time.Sleep(time.Duration(n) * state.TurnDuration)
			}
			r.outMove <- m
		}
	}()

	return r
}

func (r *RandomPlayer) Name() string {
	return r.state.Name
}

func (r *RandomPlayer) Move() <-chan player.Move {
	return r.outMove
}

func (r *RandomPlayer) Update() chan<- player.State {
	return r.update
}
