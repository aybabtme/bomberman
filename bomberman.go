package main

import (
	"github.com/aybabtme/bomberman/board"
	"github.com/aybabtme/bomberman/game"
	"github.com/aybabtme/bomberman/logger"
	"github.com/aybabtme/bomberman/objects"
	"github.com/aybabtme/bomberman/player"
	"github.com/aybabtme/bomberman/player/ai"
	"github.com/aybabtme/bomberman/player/input"
	"github.com/aybabtme/bomberman/scheduler"
	"github.com/nsf/termbox-go"
	"math/rand"
	"runtime"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

const (
	MinX = 1
	MaxX = 49
	MinY = 1
	MaxY = 21
)

const (
	LogLevel = logger.Debug

	RockFreeArea = 1
	RockDensity  = 0.50

	TotalRadiusPU = 20
	TotalBombPU   = 20

	DefaultMaxBomb    = 3
	DefaultBombRadius = 3

	TurnDuration     = time.Millisecond * 10
	TurnsToFlamout   = 70
	TurnsToReplenish = 250
	TurnsToExplode   = 200
)

var (
	h, w int

	log = logger.New("", "bomb.log", LogLevel)

	leftTopCorner = player.State{
		Name:       "p1",
		X:          MinX,
		Y:          MinY,
		Bombs:      0,
		MaxBomb:    DefaultMaxBomb,
		MaxRadius:  DefaultBombRadius,
		Alive:      true,
		GameObject: &objects.TboxPlayer{"p1"},
	}

	rightBottomCorner = player.State{
		Name:       "p2",
		X:          MaxX,
		Y:          MaxY,
		Bombs:      0,
		MaxBomb:    DefaultMaxBomb,
		MaxRadius:  DefaultBombRadius,
		Alive:      true,
		GameObject: &objects.TboxPlayer{"p2"},
	}

	leftBottomCorner = player.State{
		Name:       "p3",
		X:          MinX,
		Y:          MaxY,
		Bombs:      0,
		MaxBomb:    DefaultMaxBomb,
		MaxRadius:  DefaultBombRadius,
		Alive:      true,
		GameObject: &objects.TboxPlayer{"p3"},
	}

	rightTopCorner = player.State{
		Name:       "p4",
		X:          MaxX,
		Y:          MinY,
		Bombs:      0,
		MaxBomb:    DefaultMaxBomb,
		MaxRadius:  DefaultBombRadius,
		Alive:      true,
		GameObject: &objects.TboxPlayer{"p4"},
	}
)

func main() {
	log.Infof("Starting Bomberman")

	game := game.NewGame(TurnDuration, TotalBombPU, TotalRadiusPU)

	log.Debugf("Initializing local player.")
	localState := &leftTopCorner
	localPlayer, inputChan := initLocalPlayer(*localState)

	game.Players = map[*player.State]player.Player{
		localState:         localPlayer,
		&rightBottomCorner: ai.NewRandomPlayer(rightBottomCorner, time.Now().UnixNano()),
		&leftBottomCorner:  ai.NewWanderingPlayer(leftBottomCorner, time.Now().UnixNano()),
		&rightTopCorner:    ai.NewImmobilePlayer(rightTopCorner),
	}

	runtime.GOMAXPROCS(1 + len(game.Players))

	log.Debugf("Setup board.")
	board := board.SetupBoard(game, MaxX+2, MaxY+2, RockFreeArea, RockDensity)
	for pState := range game.Players {
		pState.CurBoard = board.Clone()
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
				select {
				case inputChan <- pm:
				default:
					log.Debugf("Dropping event '%#v', player not reading.", ev.Type)
				}

			} else {
				evChan <- ev
			}
		}
	}()

	log.Debugf("Drawing for first time.")
	board.Draw(game.Players)

	log.Debugf("Starting.")

	MainLoop(game, board, evChan)
}

func MainLoop(g *game.Game, board board.Board, evChan <-chan termbox.Event) {
	for _ = range g.TurnTick.C {
		if g.IsDone() {
			log.Infof("Game requested to stop.")
			return
		}

		receiveEvents(g, evChan)

		g.RunSchedule(func(a scheduler.Action, turn int) error {
			act := a.(*BomberAction)
			log.Debugf("[%s] !!! turn %d/%d", act.name, turn, act.Duration())
			return act.doTurn(turn)
		})

		applyPlayerMoves(g, board)
		board.Draw(g.Players)
		updatePlayers(g, board)

		alives := []player.Player{}
		for pState, player := range g.Players {
			if pState.Alive {
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

func initLocalPlayer(pState player.State) (player.Player, chan<- player.Move) {
	keyPlayerChan := make(chan player.Move, 1)
	keyPlayer := input.NewInputPlayer(pState, keyPlayerChan)
	return keyPlayer, keyPlayerChan
}

//////////////
// Events

func receiveEvents(g *game.Game, evChan <-chan termbox.Event) {
	select {
	case ev := <-evChan:
		switch ev.Type {
		case termbox.EventResize:
			w, h = ev.Width, ev.Height
		case termbox.EventError:
			g.SetDone()
		case termbox.EventKey:
			doKey(g, ev.Key)
		}
	default:
	}
}

func doKey(g *game.Game, key termbox.Key) {
	switch key {
	case termbox.KeyCtrlC:
		g.SetDone()
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

//////////////
// Players

func applyPlayerMoves(g *game.Game, board board.Board) {
	for pState, player := range g.Players {
		if pState.Alive {
			select {
			case m := <-player.Move():
				movePlayer(g, board, pState, m)
			default:
			}
		}
	}
}

func updatePlayers(game *game.Game, board board.Board) {
	for pState, player := range game.Players {
		pState.CurBoard = board.Clone()
		pState.Turn = game.Turn()
		select {
		case player.Update() <- *pState:
		default:
		}
	}
}

func toPlayerMove(ev termbox.Event) (player.Move, bool) {
	if ev.Type != termbox.EventKey {
		return player.Move(""), false
	}

	switch ev.Key {
	case termbox.KeyArrowUp:
		return player.Up, true
	case termbox.KeyArrowDown:
		return player.Down, true
	case termbox.KeyArrowLeft:
		return player.Left, true
	case termbox.KeyArrowRight:
		return player.Right, true
	case termbox.KeySpace:
		return player.PutBomb, true
	}

	return player.Move(""), false
}

func movePlayer(g *game.Game, board board.Board, pState *player.State, action player.Move) {
	nextX, nextY := pState.X, pState.Y
	switch action {
	case player.Up:
		nextY--
	case player.Down:
		nextY++
	case player.Left:
		nextX--
	case player.Right:
		nextX++
	case player.PutBomb:
		placeBomb(board, g, pState)
	}

	if !board.Traversable(nextX, nextY) {
		return
	}

	if board[nextX][nextY].Top() == objects.Flame {
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

func pickPowerUps(board board.Board, pState *player.State, x, y int) {
	c := board[x][y]
	switch c.Top() {
	case objects.BombPU:
		pState.MaxBomb++
		c.Pop()
		log.Infof("[%s] Powerup! Max bombs: %d", pState.Name, pState.MaxBomb)
	case objects.RadiusPU:
		pState.MaxRadius++
		c.Pop()
		log.Infof("[%s] Powerup! Max radius: %d", pState.Name, pState.MaxRadius)
	}
}
