package ai

import (
	"github.com/aybabtme/bomberman/player"
)

type ImmobilePlayer struct {
	state   player.State
	update  chan player.State
	outMove chan player.Move
}

func NewImmobilePlayer(state player.State) player.Player {
	return &ImmobilePlayer{
		state:   state,
		update:  make(chan player.State),
		outMove: make(chan player.Move, 1),
	}
}

func (w *ImmobilePlayer) Name() string {
	return w.state.Name
}

func (w *ImmobilePlayer) Move() <-chan player.Move {
	return w.outMove
}

func (w *ImmobilePlayer) Update() chan<- player.State {
	return w.update
}
