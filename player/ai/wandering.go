package ai

import (
	"github.com/aybabtme/bomberman/player"
	"math/rand"
	"time"
)

type WanderingPlayer struct {
	state   player.State
	update  chan player.State
	outMove chan player.Move
}

func NewWanderingPlayer(state player.State, seed int64) *WanderingPlayer {
	w := &WanderingPlayer{
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
			default:
				time.Sleep(time.Duration(n) * state.TurnDuration)
			}
			w.outMove <- m
		}
	}()

	return w
}

func (w *WanderingPlayer) Name() string {
	return w.state.Name
}

func (w *WanderingPlayer) Move() <-chan player.Move {
	return w.outMove
}

func (w *WanderingPlayer) Update() chan<- player.State {
	return w.update
}
