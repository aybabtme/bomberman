package main

import (
	"github.com/aybabtme/bomberman/logger"
	"github.com/aybabtme/bomberman/scheduler"
	"github.com/nsf/termbox-go"
	"math/rand"
	"time"
)

const (
	LogLevel = logger.Debug

	MinX           = 1
	MaxX           = 49
	MinY           = 1
	MaxY           = 21
	WallPercentage = 50

	DefaultMaxBomb    = 1
	DefaultBombRadius = 5

	TurnDuration     = time.Millisecond * 10
	TurnsToFlamout   = 70
	TurnsToReplenish = 250
	TurnsToExplode   = 200
)

var (
	WallCell = &termbox.Cell{
		Ch: '▓',
		Fg: termbox.ColorGreen,
		Bg: termbox.ColorBlack,
	}

	RockCell = &termbox.Cell{
		Ch: '▓',
		Fg: termbox.ColorYellow,
		Bg: termbox.ColorBlack,
	}

	GroundCell = &termbox.Cell{
		Ch: ' ',
		Fg: termbox.ColorDefault,
		Bg: termbox.ColorDefault,
	}

	PlayerCell = &termbox.Cell{
		Ch: '♨',
		Fg: termbox.ColorWhite,
		Bg: termbox.ColorMagenta,
	}

	BombCell = &termbox.Cell{
		Ch: 'ß',
		Fg: termbox.ColorRed,
		Bg: termbox.ColorDefault,
	}

	FlameCell = &termbox.Cell{
		Ch: '+',
		Fg: termbox.ColorRed,
		Bg: termbox.ColorDefault,
	}

	leftTopCorner = PlayerState{
		Name:     "p1",
		X:        MinX,
		Y:        MinY,
		Bombs:    0,
		MaxBomb:  DefaultMaxBomb,
		MaxRange: DefaultBombRadius,
		Alive:    true,
	}

	rightBottomCorner = PlayerState{
		Name:     "p2",
		X:        MaxX,
		Y:        MaxY,
		Bombs:    0,
		MaxBomb:  DefaultMaxBomb,
		MaxRange: DefaultBombRadius,
		Alive:    true,
	}

	leftBottomCorner = PlayerState{
		Name:     "p3",
		X:        MinX,
		Y:        MaxY,
		Bombs:    0,
		MaxBomb:  DefaultMaxBomb,
		MaxRange: DefaultBombRadius,
		Alive:    true,
	}

	rightTopCorner = PlayerState{
		Name:     "p4",
		X:        MaxX,
		Y:        MinY,
		Bombs:    0,
		MaxBomb:  DefaultMaxBomb,
		MaxRange: DefaultBombRadius,
		Alive:    false,
	}
)

var (
	log = logger.New("", "bomb.log", LogLevel)

	schedule = scheduler.NewScheduler()
	turnTick = time.NewTicker(TurnDuration)
	done     = false

	h, w int
)

func initLocalPlayer(state PlayerState) (Player, chan<- PlayerMove) {
	keyPlayerChan := make(chan PlayerMove)
	keyPlayer := NewInputPlayer(state, keyPlayerChan)
	return keyPlayer, keyPlayerChan
}

func main() {
	log.Infof("Starting Bomberman")

	log.Debugf("Initializing local player.")
	localState := &leftTopCorner
	localPlayer, inputChan := initLocalPlayer(*localState)

	players := map[*PlayerState]Player{
		&leftTopCorner:     localPlayer,
		&rightBottomCorner: NewRandomPlayer(rightBottomCorner, time.Now().UnixNano()),
		&leftBottomCorner:  NewWanderingPlayer(leftBottomCorner, time.Now().UnixNano()),
	}

	log.Debugf("Setup board.")
	board := SetupBoard(players)
	for state := range players {
		state.CurBoard = board.Clone()
	}

	log.Debugf("Initializing termbox.")
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()
	w, h = termbox.Size()

	log.Debugf("Initializing termbox event poller.")
	evChan := make(chan termbox.Event)
	go func() {
		log.Debugf("Polling events.")
		for {
			ev := termbox.PollEvent()
			if pm, ok := toPlayerMove(ev); ok {
				inputChan <- pm
			} else {
				evChan <- ev
			}
		}
	}()

	log.Debugf("Drawing for first time.")
	board.draw(players)

	log.Debugf("Starting.")
	for _ = range turnTick.C {
		if done {
			log.Infof("Game requested to stop.")
			return
		}

		receiveEvents(evChan)
		runSchedule()
		applyPlayerMoves(players, board)
		board.draw(players)
		updatePlayers(players, board)

		alives := []Player{}
		for state, player := range players {
			if state.Alive {
				alives = append(alives, player)
			}
		}
		if len(alives) == 1 {
			log.Infof("%s won. All other players are dead.", alives[0].Name())
			return
		} else if len(alives) == 0 {
			log.Infof("Draw! All players are dead.")
			return
		}
	}
}

//////////////
// Events

func receiveEvents(evChan <-chan termbox.Event) {
	select {
	case ev := <-evChan:
		log.Debugf("Event! %v", ev)
		switch ev.Type {
		case termbox.EventResize:
			w, h = ev.Width, ev.Height
		case termbox.EventError:
			done = true
		case termbox.EventKey:
			doKey(ev.Key)
		}
	default:
	}
}

func doKey(key termbox.Key) {
	switch key {
	case termbox.KeyCtrlC:
		done = true
	}
}

//////////////
// Schedule

type BomberAction struct {
	name     string
	duration int
	doTurn   func(turn int) error
}

func (a *BomberAction) Duration() int {
	return a.duration
}

func runSchedule() {
	if !schedule.HasNext() {
		return
	}
	schedule.NextTurn()

	schedule.DoTurn(func(a scheduler.Action, turn int) error {
		act := a.(*BomberAction)
		log.Debugf("[%s] !!! turn %d/%d", act.name, turn, act.Duration())
		return act.doTurn(turn)
	})
}

//////////////
// Players

func applyPlayerMoves(players map[*PlayerState]Player, board *Board) {
	for state, player := range players {
		if state.Alive {
			select {
			case m := <-player.Move():
				move(board, players, state, m)
			default:
			}
		}
	}
}

func updatePlayers(players map[*PlayerState]Player, board *Board) {
	for state, player := range players {
		state.CurBoard = board.Clone()
		select {
		case player.Update() <- *state:
		default:
		}
	}
}

func toPlayerMove(ev termbox.Event) (PlayerMove, bool) {
	if ev.Type != termbox.EventKey {
		return PlayerMove(""), false
	}

	switch ev.Key {
	case termbox.KeyArrowUp:
		return Up, true
	case termbox.KeyArrowDown:
		return Down, true
	case termbox.KeyArrowLeft:
		return Left, true
	case termbox.KeyArrowRight:
		return Right, true
	case termbox.KeySpace:
		return PutBomb, true
	}

	return PlayerMove(""), false
}

func move(board *Board, states map[*PlayerState]Player, state *PlayerState, action PlayerMove) {
	nextX, nextY := state.X, state.Y
	switch action {
	case Up:
		nextY--
	case Down:
		nextY++
	case Left:
		nextX--
	case Right:
		nextX++
	case PutBomb:
		placeBomb(board, states, state)
	}

	if !traversable(board, states, nextX, nextY) {
		return
	}

	if board[nextX][nextY] == FlameCell {
		state.Alive = false
		log.Infof("[%s] Died moving into flame.", state.Name)
	}
	state.X, state.Y = nextX, nextY
}

func traversable(board *Board, states map[*PlayerState]Player, x, y int) bool {
	switch board[x][y] {
	case GroundCell, FlameCell:
		return true
	}
	return false
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
