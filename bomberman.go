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
	RockPercentage = 50

	TotalRadiusPU = 10
	TotalBombPU   = 10

	DefaultMaxBomb    = 1
	DefaultBombRadius = 2

	TurnDuration     = time.Millisecond * 10
	TurnsToFlamout   = 70 / 3
	TurnsToReplenish = 250 / 3
	TurnsToExplode   = 200 / 3
)

type TboxGameObj struct {
	*termbox.Cell
	name        string
	traversable bool
}

func (to *TboxGameObj) Draw(x, y int) {
	termbox.SetCell(x*2, y, to.Ch, to.Fg, to.Bg)
	termbox.SetCell(x*2+1, y, to.Ch, to.Fg, to.Bg)
}

func (t *TboxGameObj) Traversable() bool {
	return t.traversable
}

func (t *TboxGameObj) GoString() string {
	return t.name
}

type TboxPlayerObj struct {
	name string
}

func (to TboxPlayerObj) Draw(x, y int) {
	fg, bg := termbox.ColorWhite, termbox.ColorMagenta
	termbox.SetCell(x*2, y, []rune(to.name)[0], fg, bg)
	termbox.SetCell(x*2+1, y, []rune(to.name)[1], fg, bg)
}

func (t TboxPlayerObj) Traversable() bool {
	return true
}

var (
	WallObj = &TboxGameObj{
		&termbox.Cell{
			Ch: '▓',
			Fg: termbox.ColorGreen,
			Bg: termbox.ColorBlack,
		},
		"Wall",
		false,
	}

	RockObj = &TboxGameObj{
		&termbox.Cell{
			Ch: '▓',
			Fg: termbox.ColorYellow,
			Bg: termbox.ColorBlack,
		},
		"Rock",
		false,
	}

	GroundObj = &TboxGameObj{
		&termbox.Cell{
			Ch: ' ',
			Fg: termbox.ColorDefault,
			Bg: termbox.ColorDefault,
		},
		"Ground",
		true,
	}

	BombObj = &TboxGameObj{
		&termbox.Cell{
			Ch: 'ß',
			Fg: termbox.ColorRed,
			Bg: termbox.ColorDefault,
		},
		"Bomb",
		false,
	}

	FlameObj = &TboxGameObj{
		&termbox.Cell{
			Ch: '+',
			Fg: termbox.ColorRed,
			Bg: termbox.ColorDefault,
		},
		"Flame",
		true,
	}

	BombPUObj = &TboxGameObj{
		&termbox.Cell{
			Ch: 'Ⓑ',
			Fg: termbox.ColorYellow,
			Bg: termbox.ColorMagenta,
		},
		"PowerUp(Bomb)",
		true,
	}

	RadiusPUObj = &TboxGameObj{
		&termbox.Cell{
			Ch: 'Ⓡ',
			Fg: termbox.ColorYellow,
			Bg: termbox.ColorMagenta,
		},
		"PowerUp(Radius)",
		true,
	}

	leftTopCorner = PlayerState{
		Name:       "p1",
		X:          MinX,
		Y:          MinY,
		Bombs:      0,
		MaxBomb:    DefaultMaxBomb,
		MaxRadius:  DefaultBombRadius,
		Alive:      true,
		GameObject: &TboxPlayerObj{"p1"},
	}

	rightBottomCorner = PlayerState{
		Name:       "p2",
		X:          MaxX,
		Y:          MaxY,
		Bombs:      0,
		MaxBomb:    DefaultMaxBomb,
		MaxRadius:  DefaultBombRadius,
		Alive:      true,
		GameObject: &TboxPlayerObj{"p2"},
	}

	leftBottomCorner = PlayerState{
		Name:       "p3",
		X:          MinX,
		Y:          MaxY,
		Bombs:      0,
		MaxBomb:    DefaultMaxBomb,
		MaxRadius:  DefaultBombRadius,
		Alive:      true,
		GameObject: &TboxPlayerObj{"p3"},
	}

	rightTopCorner = PlayerState{
		Name:       "p4",
		X:          MaxX,
		Y:          MinY,
		Bombs:      0,
		MaxBomb:    DefaultMaxBomb,
		MaxRadius:  DefaultBombRadius,
		Alive:      true,
		GameObject: &TboxPlayerObj{"p4"},
	}
)

type Game struct {
	schedule *scheduler.Scheduler
	turnTick *time.Ticker
	done     bool

	players map[*PlayerState]Player

	h, w                    int
	bombPULeft, rangePULeft int
}

var (
	log = logger.New("", "bomb.log", LogLevel)
)

func main() {
	log.Infof("Starting Bomberman")

	game := &Game{
		schedule:    scheduler.NewScheduler(),
		turnTick:    time.NewTicker(TurnDuration),
		done:        false,
		bombPULeft:  TotalBombPU,
		rangePULeft: TotalRadiusPU,
	}

	log.Debugf("Initializing local player.")
	localState := &leftTopCorner
	localPlayer, inputChan := initLocalPlayer(*localState)

	game.players = map[*PlayerState]Player{
		localState: localPlayer,
		// &rightBottomCorner: NewRandomPlayer(rightBottomCorner, time.Now().UnixNano()),
		// &leftBottomCorner:  NewWanderingPlayer(leftBottomCorner, time.Now().UnixNano()),
		&rightTopCorner: NewImmobilePlayer(rightTopCorner),
	}

	log.Debugf("Setup board.")
	board := SetupBoard(game)
	for state := range game.players {
		state.CurBoard = board.Clone()
	}

	log.Debugf("Initializing termbox.")
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()
	game.w, game.h = termbox.Size()

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
	board.draw(game.players)

	log.Debugf("Starting.")
	for _ = range game.turnTick.C {
		if game.done {
			log.Infof("Game requested to stop.")
			return
		}

		receiveEvents(game, evChan)
		runSchedule(game)
		applyPlayerMoves(game, board)
		board.draw(game.players)
		updatePlayers(game.players, board)

		alives := []Player{}
		for state, player := range game.players {
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

func initLocalPlayer(state PlayerState) (Player, chan<- PlayerMove) {
	keyPlayerChan := make(chan PlayerMove)
	keyPlayer := NewInputPlayer(state, keyPlayerChan)
	return keyPlayer, keyPlayerChan
}

//////////////
// Events

func receiveEvents(game *Game, evChan <-chan termbox.Event) {
	select {
	case ev := <-evChan:
		switch ev.Type {
		case termbox.EventResize:
			game.w, game.h = ev.Width, ev.Height
		case termbox.EventError:
			game.done = true
		case termbox.EventKey:
			doKey(game, ev.Key)
		}
	default:
	}
}

func doKey(game *Game, key termbox.Key) {
	switch key {
	case termbox.KeyCtrlC:
		game.done = true
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

func runSchedule(game *Game) {
	if !game.schedule.HasNext() {
		return
	}
	game.schedule.NextTurn()

	game.schedule.DoTurn(func(a scheduler.Action, turn int) error {
		act := a.(*BomberAction)
		log.Debugf("[%s] !!! turn %d/%d", act.name, turn, act.Duration())
		return act.doTurn(turn)
	})
}

//////////////
// Players

func applyPlayerMoves(game *Game, board *Board) {
	for state, player := range game.players {
		if state.Alive {
			select {
			case m := <-player.Move():
				move(game, board, state, m)
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

func move(game *Game, board *Board, pState *PlayerState, action PlayerMove) {
	nextX, nextY := pState.X, pState.Y
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
		placeBomb(board, game, pState)
	}

	if !board.Traversable(nextX, nextY) {
		return
	}

	if board[nextX][nextY].Top() == FlameObj {
		pState.Alive = false
		log.Infof("[%s] Died moving into flame.", pState.Name)
		cell := board[pState.X][pState.Y]
		if !cell.Remove(pState.GameObject) {
			log.Panicf("[%s] player not found at (%d, %d), cell=%#v",
				pState.Name, pState.X, pState.Y, cell)
		}
		return
	}

	pState.LastX, pState.LastY = pState.X, pState.Y
	pState.X, pState.Y = nextX, nextY

	pickPowerUps(board, pState, nextX, nextY)

	cell := board[pState.LastX][pState.LastY]
	if !cell.Remove(pState.GameObject) {
		log.Panicf("[%s] player not found at (%d, %d), cell=%#v",
			pState.Name, pState.X, pState.Y, cell)
	}
	board[nextX][nextY].Push(pState.GameObject)
}

func pickPowerUps(board *Board, pState *PlayerState, x, y int) {
	switch board[x][y].Top() {
	case BombPUObj:
		pState.MaxBomb++
		board[x][y].Pop()
		log.Infof("[%s] Powerup! Max bombs: %d", pState.Name, pState.MaxBomb)
	case RadiusPUObj:
		pState.MaxRadius++
		board[x][y].Pop()
		log.Infof("[%s] Powerup! Max radius: %d", pState.Name, pState.MaxRadius)
	}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
