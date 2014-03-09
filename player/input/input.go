package input

import (
	"github.com/aybabtme/bomberman/player"
)

type InputPlayer struct {
	state player.State

	// Comms
	update  chan player.State
	inMove  <-chan player.Move
	outMove chan player.Move
}

func NewInputPlayer(state player.State, input <-chan player.Move) player.Player {
	i := &InputPlayer{
		state:   state,
		update:  make(chan player.State),
		inMove:  input,
		outMove: make(chan player.Move, 1), // Rate-limiting to 1 move per turn
	}

	go func() {
		for i.state.Alive {
			select {
			case move := <-i.inMove:
				i.forwardMove(move)
			case i.state = <-i.update:
			}
		}
	}()

	return i
}

func (i *InputPlayer) forwardMove(move player.Move) {
	select {
	case i.outMove <- move:
	default:
		// Drop it
	}
}

func (i *InputPlayer) Name() string {
	return i.state.Name
}

func (i *InputPlayer) Move() <-chan player.Move {
	return i.outMove
}

func (i *InputPlayer) Update() chan<- player.State {
	return i.update
}
