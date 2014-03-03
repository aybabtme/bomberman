package main

import (
	"github.com/nsf/termbox-go"
)

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
