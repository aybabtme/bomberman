package objects

import (
	"github.com/aybabtme/bomberman/cell"
	"github.com/nsf/termbox-go"
)

// safety check, forces compiler to complain if they dont
func __mustImplGameObject() []cell.GameObject {
	return []cell.GameObject{
		&TboxObj{},
		&TboxPlayer{},
	}
}

var (
	Wall = &TboxObj{
		&termbox.Cell{
			Ch: '▓',
			Fg: termbox.ColorGreen,
			Bg: termbox.ColorBlack,
		},
		"Wall",
		false,
	}

	Rock = &TboxObj{
		&termbox.Cell{
			Ch: '▓',
			Fg: termbox.ColorYellow,
			Bg: termbox.ColorBlack,
		},
		"Rock",
		false,
	}

	Ground = &TboxObj{
		&termbox.Cell{
			Ch: ' ',
			Fg: termbox.ColorDefault,
			Bg: termbox.ColorDefault,
		},
		"Ground",
		true,
	}

	Bomb = &TboxObj{
		&termbox.Cell{
			Ch: 'ß',
			Fg: termbox.ColorRed,
			Bg: termbox.ColorDefault,
		},
		"Bomb",
		false,
	}

	Flame = &TboxObj{
		&termbox.Cell{
			Ch: '+',
			Fg: termbox.ColorRed,
			Bg: termbox.ColorDefault,
		},
		"Flame",
		true,
	}

	BombPU = &TboxObj{
		&termbox.Cell{
			Ch: 'Ⓑ',
			Fg: termbox.ColorYellow,
			Bg: termbox.ColorMagenta,
		},
		"PowerUp(Bomb)",
		true,
	}

	RadiusPU = &TboxObj{
		&termbox.Cell{
			Ch: 'Ⓡ',
			Fg: termbox.ColorYellow,
			Bg: termbox.ColorMagenta,
		},
		"PowerUp(Radius)",
		true,
	}
)

type TboxObj struct {
	*termbox.Cell
	name        string
	traversable bool
}

func (to *TboxObj) Draw(x, y int) {
	termbox.SetCell(x*2, y, to.Ch, to.Fg, to.Bg)
	termbox.SetCell(x*2+1, y, to.Ch, to.Fg, to.Bg)
}

func (t *TboxObj) Traversable() bool {
	return t.traversable
}

func (t *TboxObj) String() string {
	return t.name
}

type TboxPlayer struct {
	Name string
}

func (t TboxPlayer) Draw(x, y int) {
	fg, bg := termbox.ColorWhite, termbox.ColorMagenta
	termbox.SetCell(x*2, y, []rune(t.Name)[0], fg, bg)
	termbox.SetCell(x*2+1, y, []rune(t.Name)[1], fg, bg)
}

func (t TboxPlayer) Traversable() bool {
	return true
}

func (t *TboxPlayer) String() string {
	return t.Name
}
