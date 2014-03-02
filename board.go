package main

import (
	"github.com/nsf/termbox-go"
	"math/rand"
)

type Board [MaxX + 2][MaxY + 2]*termbox.Cell

func SetupBoard(players map[*PlayerState]Player) *Board {
	board := &Board{}

	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[0]); j++ {
			switch {
			case i == 0 || i == len(board)-1:
				board[i][j] = WallCell
			case j == 0 || j == len(board[0])-1:
				board[i][j] = WallCell
			case j%2 == 0 && i%2 == 0:
				board[i][j] = WallCell
			default:
				if rand.Intn(100) < 100-WallPercentage {
					board[i][j] = GroundCell
				} else {
					board[i][j] = RockCell
				}
			}
		}
	}

	for state := range players {
		if !state.Alive {
			continue
		}

		x, y := state.X, state.Y
		board.asSquare(x, y, 1, func(cellX, cellY int) {
			if board[cellX][cellY] == RockCell {
				board[cellX][cellY] = GroundCell
			}
		})
	}

	return board
}

func (b *Board) Traversable(x, y int) bool {
	switch b[x][y] {
	case GroundCell, FlameCell, BombPUCell, RadiusPUCell:
		return true
	}
	return false
}

func (board *Board) draw(players map[*PlayerState]Player) {
	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[0]); j++ {
			cell := board[i][j]
			termbox.SetCell(i*2, j, cell.Ch, cell.Fg, cell.Bg)
			termbox.SetCell(i*2+1, j, cell.Ch, cell.Fg, cell.Bg)
		}
	}

	for state, player := range players {
		if !state.Alive {
			continue
		}
		termbox.SetCell(state.X*2, state.Y, []rune(player.Name())[0], PlayerCell.Fg, PlayerCell.Bg)
		termbox.SetCell(state.X*2+1, state.Y, []rune(player.Name())[1], PlayerCell.Fg, PlayerCell.Bg)
	}

	termbox.Flush()
}

func (b *Board) Clone() Board {
	clone := Board{}
	for i := range b {
		for j := range b[i] {
			clone[i][j] = &(*b[i][j])
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
		if b[i][y] == WallCell {
			break
		}
		apply(i, y)
	}

	// left of (x,y)
	for i := x - 1; i > max(x-dist, 0); i-- {
		if b[i][y] == WallCell {
			break
		}
		apply(i, y)
	}

	// below (x,y)
	for j := y + 1; j < min(y+dist, len(b)); j++ {
		if b[x][j] == WallCell {
			break
		}
		apply(x, j)
	}

	// above (x,y)
	for j := y - 1; j > max(y-dist, 0); j-- {
		if b[x][j] == WallCell {
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
