package main

import (
	"github.com/aybabtme/bomberman/logger"
	"github.com/aybabtme/bomberman/scheduler"
	"github.com/nsf/termbox-go"
	"time"
)

var (
	Wall = &termbox.Cell{
		Ch: '▓',
		Fg: termbox.ColorGreen,
		Bg: termbox.ColorBlack,
	}

	Rock = &termbox.Cell{
		Ch: '▓',
		Fg: termbox.ColorYellow,
		Bg: termbox.ColorBlack,
	}

	Ground = &termbox.Cell{
		Ch: ' ',
		Fg: termbox.ColorDefault,
		Bg: termbox.ColorDefault,
	}

	Player = &termbox.Cell{
		Ch: '♨',
		Fg: termbox.ColorWhite,
		Bg: termbox.ColorMagenta,
	}

	Bomb = &termbox.Cell{
		Ch: 'ß',
		Fg: termbox.ColorRed,
		Bg: termbox.ColorDefault,
	}

	Flame = &termbox.Cell{
		Ch: '+',
		Fg: termbox.ColorRed,
		Bg: termbox.ColorDefault,
	}
)

type BomberAction struct {
	name     string
	duration int
	doTurn   func(turn int) error
}

func (a *BomberAction) Duration() int {
	return a.duration
}

var (
	log = logger.New("", "bomb.log", logger.Debug)

	schedule     = scheduler.NewScheduler()
	turnDuration = time.Millisecond * 10
	turnTick     = time.NewTicker(turnDuration)
	done         = false

	board [81][25]*termbox.Cell

	x, y         int = 1, 1
	lastX, lastY int
	h, w         int

	blastRadius = 3

	canPlaceBomb = true
)

func setupBoard() {
	log.Debugf("Setup board")
	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[0]); j++ {
			switch {
			case i == 0 || i == len(board)-1:
				board[i][j] = Wall
			case j == 0 || j == len(board[0])-1:
				board[i][j] = Wall
			case j%2 == 0 && i%2 == 0:
				board[i][j] = Wall
			default:
				board[i][j] = Rock
			}
		}
	}

	around(x, y, 3, func(cellX, cellY int) {
		if board[cellX][cellY] == Rock {
			board[cellX][cellY] = Ground
		}
	})
	log.Debugf("Setup board done.")
}

func main() {
	log.Debugf("Starting.")
	setupBoard()
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	evChan := make(chan termbox.Event)

	go func() {
		log.Debugf("Polling events.")
		for {
			ev := termbox.PollEvent()
			evChan <- ev
		}
	}()

	w, h = termbox.Size()
	log.Debugf("Drawing for first time.")
	draw()

	for _ = range turnTick.C {
		if done {
			return
		}
		select {
		case ev := <-evChan:
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

		if schedule.HasNext() {
			schedule.NextTurn()
			schedule.DoTurn(func(a scheduler.Action, turn int) error {
				act := a.(*BomberAction)
				log.Debugf("Doing action '%s', turn %d/%d", act.name, turn, act.Duration())
				return act.doTurn(turn)
			})
		}

		draw()
	}

}

func draw() {
	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[0]); j++ {
			cell := board[i][j]
			termbox.SetCell(i*2, j, cell.Ch, cell.Fg, cell.Bg)
			termbox.SetCell(i*2+1, j, cell.Ch, cell.Fg, cell.Bg)
		}
	}
	termbox.SetCell(x*2, y, Player.Ch, Player.Fg, Player.Bg)
	termbox.SetCell(x*2+1, y, Player.Ch, Player.Fg, Player.Bg)
	termbox.Flush()
}

func doKey(key termbox.Key) {
	switch key {
	case termbox.KeyCtrlC,
		termbox.KeyArrowUp,
		termbox.KeyArrowDown,
		termbox.KeyArrowLeft,
		termbox.KeyArrowRight:
		move(key)
	case termbox.KeySpace:
		placeBomb()
	}
}

func move(key termbox.Key) {
	nextX, nextY := x, y
	switch key {
	case termbox.KeyCtrlC:
		done = true
	case termbox.KeyArrowUp:
		nextY--
	case termbox.KeyArrowDown:
		nextY++
	case termbox.KeyArrowLeft:
		nextX--
	case termbox.KeyArrowRight:
		nextX++
	}

	if canMove(nextX, nextY) {
		lastX, lastY = x, y
		x, y = nextX, nextY
	}
}

func canMove(x, y int) bool {
	switch board[x][y] {
	case Ground:
		return true
	case Flame:
		done = true
		return true
	}
	return false
}

func placeBomb() {
	log.Debugf("Attempting to place bomb.")
	if canPlaceBomb {
		canPlaceBomb = false
	} else {
		log.Debugf("Failed.")
		return
	}

	board[x][y] = Bomb
	tmpX, tmpY := x, y

	allowBombs := func(turn int) error {
		canPlaceBomb = true
		return nil
	}

	doFlameout := func(turn int) error {
		log.Debugf("Flameout.")
		removeFlame(tmpX, tmpY)

		return nil
	}

	doExplosion := func(turn int) error {
		log.Debugf("Exploding.")
		if turn == 0 {

			explode(tmpX, tmpY)
			log.Debugf("Registering flameout.")
			schedule.Register(&BomberAction{
				name:     "doFlameout",
				duration: 1,
				doTurn:   doFlameout,
			}, 70)

			log.Debugf("Registering bomb allowance.")
			schedule.Register(&BomberAction{
				name:     "allowBombs",
				duration: 1,
				doTurn:   allowBombs,
			}, 250)
		}

		switch turn % 4 {
		case 0:
			Flame.Ch = 'x'
			Flame.Fg = termbox.ColorYellow
			Flame.Bg = termbox.ColorWhite
		case 1:
			Flame.Ch = '*'
			Flame.Fg = termbox.ColorYellow
			Flame.Bg = termbox.ColorBlack
		case 2:
			Flame.Ch = '+'
			Flame.Fg = termbox.ColorRed
			Flame.Bg = termbox.ColorYellow
		case 3:
			Flame.Ch = '*'
			Flame.Fg = termbox.ColorRed
			Flame.Bg = termbox.ColorBlack
		}

		return nil
	}

	log.Debugf("Registering explosion.")
	schedule.Register(&BomberAction{
		name:     "doExplosion",
		duration: 70,
		doTurn:   doExplosion,
	}, 200)
}

func explode(eplodeX, eplodeY int) {
	board[eplodeX][eplodeY] = Ground
	cross(eplodeX, eplodeY, blastRadius, func(cellX, cellY int) {
		if cellX == x && cellY == y {
			done = true
		}

		if board[cellX][cellY] != Wall {
			board[cellX][cellY] = Flame
		}
	})
}

func removeFlame(x, y int) {
	cross(x, y, blastRadius, func(cellX, cellY int) {
		if board[cellX][cellY] == Flame {
			board[cellX][cellY] = Ground
		}
	})
}

func min(n, m int) int {
	if n < m {
		return n
	}
	return m
}

func max(n, m int) int {
	if n > m {
		return n
	}
	return m
}

func around(x, y, rad int, apply func(int, int)) {
	for i := max(x-rad, 0); i < min(x+rad, len(board)); i++ {
		for j := max(y-rad, 0); j < min(y+rad, len(board[0])); j++ {
			apply(i, j)
		}
	}
}

func cross(x, y, dist int, apply func(int, int)) {
	// (x,y) and to the right
	for i := x; i < min(x+dist, len(board)); i++ {
		if board[i][y] == Wall {
			break
		}
		apply(i, y)
	}

	// left of (x,y)
	for i := x - 1; i > max(x-dist, 0); i-- {
		if board[i][y] == Wall {
			break
		}
		apply(i, y)
	}

	// below (x,y)
	for j := y + 1; j < min(y+dist, len(board)); j++ {
		if board[x][j] == Wall {
			break
		}
		apply(x, j)
	}

	// above (x,y)
	for j := y - 1; j > max(y-dist, 0); j-- {
		if board[x][j] == Wall {
			break
		}
		apply(x, j)
	}
}
