package main

import (
	"github.com/aybabtme/bomberman/cell"
	"github.com/aybabtme/bomberman/logger"
	"github.com/aybabtme/bomberman/scheduler"
	"github.com/nsf/termbox-go"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

const (
	LogLevel = logger.Debug

	MinX = 1
	MaxX = 49
	MinY = 1
	MaxY = 21

	RockFreeArea   = 1
	RockPercentage = 50

	TotalRadiusPU = 20
	TotalBombPU   = 20

	DefaultMaxBomb    = 1
	DefaultBombRadius = 2

	TurnDuration     = time.Millisecond * 10
	TurnsToFlamout   = 70
	TurnsToReplenish = 250
	TurnsToExplode   = 200
)

var (
	log = logger.New("", "bomb.log", LogLevel)

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

	h, w                     int
	bombPULeft, radiusPULeft int
}

func (g *Game) probablyPutRadiusUP(c *cell.Cell, rocksToUse int) {
	if rand.Float32() < float32(g.radiusPULeft)/float32(rocksToUse) {
		g.radiusPULeft--
		c.Push(RadiusPUObj)
	}
}

func (g *Game) probablyPutBombUP(c *cell.Cell, rocksToUse int) {
	if rand.Float32() < float32(g.bombPULeft)/float32(rocksToUse) {
		g.bombPULeft--
		c.Push(BombPUObj)
	}
}

func main() {
	log.Infof("Starting Bomberman")

	game := &Game{
		schedule:     scheduler.NewScheduler(),
		turnTick:     time.NewTicker(TurnDuration),
		done:         false,
		bombPULeft:   TotalBombPU,
		radiusPULeft: TotalRadiusPU,
	}

	log.Debugf("Initializing local player.")
	localState := &leftTopCorner
	localPlayer, inputChan := initLocalPlayer(*localState)

	game.players = map[*PlayerState]Player{
		localState:         localPlayer,
		&rightBottomCorner: NewRandomPlayer(rightBottomCorner, time.Now().UnixNano()),
		&leftBottomCorner:  NewWanderingPlayer(leftBottomCorner, time.Now().UnixNano()),
		&rightTopCorner:    NewImmobilePlayer(rightTopCorner),
	}

	log.Debugf("Setup board.")
	board := SetupBoard(game)
	for pState := range game.players {
		pState.CurBoard = board.Clone()
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
	board.draw(game.players)

	log.Debugf("Starting.")

	game.mainLoop(board, evChan)
}

func (g *Game) mainLoop(board *Board, evChan <-chan termbox.Event) {
	for _ = range g.turnTick.C {
		if g.done {
			log.Infof("Game requested to stop.")
			return
		}

		g.receiveEvents(evChan)
		g.runSchedule()
		g.applyPlayerMoves(board)
		board.draw(g.players)
		updatePlayers(g.players, board)

		alives := []Player{}
		for pState, player := range g.players {
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

func initLocalPlayer(pState PlayerState) (Player, chan<- PlayerMove) {
	keyPlayerChan := make(chan PlayerMove, 1)
	keyPlayer := NewInputPlayer(pState, keyPlayerChan)
	return keyPlayer, keyPlayerChan
}

//////////////
// Events

func (g *Game) receiveEvents(evChan <-chan termbox.Event) {
	select {
	case ev := <-evChan:
		switch ev.Type {
		case termbox.EventResize:
			g.w, g.h = ev.Width, ev.Height
		case termbox.EventError:
			g.done = true
		case termbox.EventKey:
			g.doKey(ev.Key)
		}
	default:
	}
}

func (g *Game) doKey(key termbox.Key) {
	switch key {
	case termbox.KeyCtrlC:
		g.done = true
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

func (g *Game) runSchedule() {
	if !g.schedule.HasNext() {
		return
	}
	g.schedule.NextTurn()

	g.schedule.DoTurn(func(a scheduler.Action, turn int) error {
		act := a.(*BomberAction)
		log.Debugf("[%s] !!! turn %d/%d", act.name, turn, act.Duration())
		return act.doTurn(turn)
	})
}

//////////////
// Players

func (g *Game) applyPlayerMoves(board *Board) {
	for pState, player := range g.players {
		if pState.Alive {
			select {
			case m := <-player.Move():
				g.movePlayer(board, pState, m)
			default:
			}
		}
	}
}

func updatePlayers(players map[*PlayerState]Player, board *Board) {
	for pState, player := range players {
		pState.CurBoard = board.Clone()
		select {
		case player.Update() <- *pState:
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

func (g *Game) movePlayer(board *Board, pState *PlayerState, action PlayerMove) {
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
		placeBomb(board, g, pState)
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
	c := board[x][y]
	switch c.Top() {
	case BombPUObj:
		pState.MaxBomb++
		c.Pop()
		log.Infof("[%s] Powerup! Max bombs: %d", pState.Name, pState.MaxBomb)
	case RadiusPUObj:
		pState.MaxRadius++
		c.Pop()
		log.Infof("[%s] Powerup! Max radius: %d", pState.Name, pState.MaxRadius)
	}
}
