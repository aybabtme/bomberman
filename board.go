package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"math/rand"
)

// Cell is a cell on the board. A cell can have many z layers.
type Cell struct {
	base    GameObject
	zLayers []GameObject
	X, Y    int
}

// NewCell creates a cell with base as z layer 0.
func NewCell(base GameObject, x, y int) *Cell {
	return &Cell{
		base:    base,
		zLayers: make([]GameObject, 0, 0),
		X:       x,
		Y:       y,
	}
}

// Top returns the object on top of the z layers.
func (c *Cell) Top() GameObject {
	if len(c.zLayers) == 0 {
		return c.base
	}
	return c.zLayers[len(c.zLayers)-1]
}

// Push adds an object to the top of the z layers.
func (c *Cell) Push(o GameObject) {
	c.zLayers = append(c.zLayers, o)
}

// Pop returns the top object.  It will remove the object from the layers unless
// it's the base layer.
func (c *Cell) Pop() (GameObject, bool) {
	if len(c.zLayers) == 0 {
		return c.base, false
	}
	var pop GameObject
	pop = c.zLayers[len(c.zLayers)-1]
	c.zLayers = c.zLayers[:len(c.zLayers)-1]
	return pop, true
}

// RemoveLayer removes the object at layer z, unless if z is the base layer. If
// z is out of bound, this will panic.
func (c *Cell) RemoveLayer(z int) GameObject {
	// Case z == 0
	if z == 0 {
		return c.base
	}

	// Case z >= len, z <= 0: panic
	removed := c.zLayers[z-1]

	// Case z == 1: will skip body of loop
	for i := z - 1; i < len(c.zLayers)-1; i++ {
		c.zLayers[i] = c.zLayers[i+1]
	}

	// z has at least 1 element
	c.zLayers = c.zLayers[:len(c.zLayers)-1]
	return removed
}

// Remove finds an GameObject to remove in the Cells z layers. The base
// layer can't be removed.
func (c *Cell) Remove(toRemove GameObject) bool {
	if c.base == toRemove {
		return false
	}
	for z, obj := range c.zLayers {
		if obj == toRemove {
			c.RemoveLayer(z + 1)
			return true
		}
	}
	return false
}

// Layer looks up the object at layer z. If z is out of bound, this will panic.
func (c *Cell) Layer(z int) GameObject {
	if z == 0 {
		return c.base
	}
	// Invalid Z will panic
	return c.zLayers[z-1]
}

// Depth gives the depth of the z layer, including the base layer.
func (c *Cell) Depth() int {
	return 1 + len(c.zLayers)
}

func (c *Cell) GoString() string {
	return fmt.Sprintf("cell{x=%d,y=%d,z=%d,base=%#v,layers=%#v}",
		c.X, c.Y, c.Depth(), c.base, c.zLayers)
}

type Board [MaxX + 2][MaxY + 2]*Cell

func SetupBoard(game *Game) *Board {
	board := &Board{}

	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[0]); j++ {
			board[i][j] = NewCell(GroundObj, i, j)

			switch {
			case i == 0 || i == len(board)-1:
				board[i][j].Push(WallObj)
			case j == 0 || j == len(board[0])-1:
				board[i][j].Push(WallObj)
			case j%2 == 0 && i%2 == 0:
				board[i][j].Push(WallObj)
			default:
				if rand.Intn(100) < 100-WallPercentage {
					// Push nothing, top is GroundObj
				} else {
					board[i][j].Push(RockObj)
				}
			}
		}
	}

	for state := range game.players {
		if !state.Alive {
			continue
		}

		x, y := state.X, state.Y
		board.asSquare(x, y, 5, func(cellX, cellY int) {
			cell := board[cellX][cellY]
			if cell.Top() == RockObj {
				cell.Pop()
			}
		})
		board[x][y].Push(state.GameObject)
	}

	return board
}

func (b *Board) Traversable(x, y int) bool {
	return b[x][y].Top().Traversable()
}

func (board *Board) draw(players map[*PlayerState]Player) {
	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[0]); j++ {
			board[i][j].Top().Draw(i, j)
		}
	}

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
	for i := range b {
		for j := range b[i] {
			clone[i][j] = NewCell(b[i][j].Top(), i, j)
		}
	}
	return clone
}

func (b *Board) asSquare(x, y, rad int, apply func(int, int)) {
	for i := max(x-rad, 0); i <= min(x+rad, len(b)-1); i++ {
		for j := max(y-rad, 0); j <= min(y+rad, len(b[0])-1); j++ {
			apply(i, j)
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
