package game

import (
	"github.com/aybabtme/bomberman/cell"
	"github.com/aybabtme/bomberman/objects"
	"github.com/aybabtme/bomberman/player"
	"github.com/aybabtme/bomberman/scheduler"
	"math/rand"
	"time"
)

type Game struct {
	rnd *rand.Rand

	Schedule *scheduler.Scheduler
	TurnTick *time.Ticker
	turn     int
	done     bool

	Players map[*player.State]player.Player

	bombPULeft, radiusPULeft int
}

func NewGame(turnDuration time.Duration, totalBombs, totalRadius int) *Game {
	return &Game{
		rnd:          rand.New(rand.NewSource(time.Now().UnixNano())),
		Schedule:     scheduler.NewScheduler(),
		TurnTick:     time.NewTicker(turnDuration),
		done:         false,
		bombPULeft:   totalBombs,
		radiusPULeft: totalRadius,
	}
}

func (g *Game) TryPutRadiusPU(c *cell.Cell, rocksToUse int) {
	if g.rnd.Float32() < float32(g.radiusPULeft)/float32(rocksToUse) {
		g.radiusPULeft--
		c.Push(objects.RadiusPU)
	}
}

func (g *Game) TryPutBombPU(c *cell.Cell, rocksToUse int) {
	if g.rnd.Float32() < float32(g.bombPULeft)/float32(rocksToUse) {
		g.bombPULeft--
		c.Push(objects.BombPU)
	}
}

func (g *Game) SetDone() {
	g.done = true
}

func (g *Game) IsDone() bool {
	return g.done
}

func (g *Game) RunSchedule(onTurn func(scheduler.Action, int) error) {
	if g.Schedule.HasNext() {
		g.Schedule.NextTurn()
		g.Schedule.DoTurn(onTurn)
		g.turn++
	}
}

func (g *Game) Turn() int {
	return g.turn
}
