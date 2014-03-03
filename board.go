package main

import (
	"github.com/aybabtme/bomberman/cell"
	"github.com/nsf/termbox-go"
	"math/rand"
)

type Board [MaxX + 2][MaxY + 2]*cell.Cell

func (board *Board) setupMap() (free, used int) {
	board.forEachIndex(func(_ *cell.Cell, x, y int) {
		board[x][y] = cell.NewCell(GroundObj, x, y)
		switch {
		case
			x == 0 || x == len(board)-1,    // Left and right
			y == 0 || y == len(board[0])-1, // Top and bottom
			y%2 == 0 && x%2 == 0:           // Every second cell
			board[x][y].Push(WallObj)
			used++
		default:
			free++
		}
	})
	return
}

func SetupBoard(game *Game) *Board {
	board := &Board{}

	free, _ := board.setupMap()

	needRock := free * RockPercentage
	rockPlaced := 0
	rockProb := func(rockLeft, freeCell int) float32 {
		return float32(rockLeft) / float32(freeCell) / 100.0
	}

	putRock := func(cell *cell.Cell) {
		needRock--
		rockPlaced++
		cell.Push(RockObj)
	}

	groundTest := func(c *cell.Cell) bool { return c.Top() == GroundObj }
	board.filter(groundTest, func(c *cell.Cell) {
		dice := rand.Float32()
		prob := rockProb(needRock, free)
		if dice < prob {
			putRock(c)
		}
	})

	for state := range game.players {
		if !state.Alive {
			continue
		}

		x, y := state.X, state.Y
		board.asSquare(x, y, 5, func(cell *cell.Cell) {
			if cell.Top() == RockObj {
				cell.Pop()
				rockPlaced--
			}
		})
		board[x][y].Push(state.GameObject)
	}

	bombProb := func(bombLeft, freeRocks int) float32 {
		return float32(bombLeft) / float32(freeRocks)
	}

	radiusProb := func(radiusLeft, freeRocks int) float32 {
		return float32(radiusLeft) / float32(freeRocks)
	}

	rocksForBomb := rockPlaced / 2
	rocksForRadius := rockPlaced / 2

	putBombPU := func(c *cell.Cell) {
		if rand.Float32() < bombProb(game.bombPULeft, rocksForBomb) {
			game.bombPULeft--
			c.Push(BombPUObj)
		}
		rocksForBomb--
	}

	putRadiusPU := func(c *cell.Cell) {
		if rand.Float32() < radiusProb(game.rangePULeft, rocksForRadius) {
			game.rangePULeft--
			c.Push(RadiusPUObj)
		}
		rocksForRadius--
	}

	board.filter(func(c *cell.Cell) bool { return c.Top() == RockObj },
		func(c *cell.Cell) {
			// rock, _ := c.Pop()
			switch rand.Intn(2) {
			case 0:
				putBombPU(c)
			case 1:
				putRadiusPU(c)
			}
			// c.Push(rock)
		})

	return board
}

func (b *Board) Traversable(x, y int) bool {
	return b[x][y].Top().Traversable()
}

func (board *Board) draw(players map[*PlayerState]Player) {
	board.forEach(func(c *cell.Cell) {
		c.Top().Draw(c.X, c.Y)
	})

	for state := range players {
		if !state.Alive {
			board[state.X][state.Y].Remove(state.GameObject)
			continue
		}
	}

	termbox.Flush()
}

func (b *Board) Clone() Board {
	clone := Board{}
	b.forEach(func(c *cell.Cell) {
		clone[c.X][c.Y] = cell.NewCell(c.Top(), c.X, c.Y)
	})
	return clone
}

///////////
// Helpers

// functional iterations

func (b *Board) forEachIndex(apply func(*cell.Cell, int, int)) {
	for x, inner := range b {
		for y, cell := range inner {
			apply(cell, x, y)
		}
	}
}

func (b *Board) forEach(apply func(*cell.Cell)) {
	b.forEachIndex(func(c *cell.Cell, x, y int) { apply(c) })
}

func (b *Board) filter(test func(*cell.Cell) bool, apply func(*cell.Cell)) {
	b.forEach(func(cell *cell.Cell) {
		if test(cell) {
			apply(cell)
		}
	})
}

func (b *Board) asSquare(x, y, rad int, apply func(*cell.Cell)) {
	for i := max(x-rad, 0); i <= min(x+rad, len(b)-1); i++ {
		for j := max(y-rad, 0); j <= min(y+rad, len(b[0])-1); j++ {
			apply(b[i][j])
		}
	}
}

func (b *Board) asCross(x, y, dist int, apply func(int, int)) {
	// (x,y) and to the right
	for i := x; i < min(x+dist, len(b)); i++ {
		if b[i][y].Top() == WallObj {
			break
		}
		apply(i, y)
	}

	// left of (x,y)
	for i := x - 1; i > max(x-dist, 0); i-- {
		if b[i][y].Top() == WallObj {
			break
		}
		apply(i, y)
	}

	// below (x,y)
	for j := y + 1; j < min(y+dist, len(b)); j++ {
		if b[x][j].Top() == WallObj {
			break
		}
		apply(x, j)
	}

	// above (x,y)
	for j := y - 1; j > max(y-dist, 0); j-- {
		if b[x][j].Top() == WallObj {
			break
		}
		apply(x, j)
	}
}

// Integer math

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
