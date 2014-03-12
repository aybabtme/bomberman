package main

import (
	"fmt"
	"github.com/aybabtme/bomberman/board"
	"github.com/aybabtme/bomberman/cell"
	"github.com/aybabtme/bomberman/game"
	"github.com/aybabtme/bomberman/objects"
	"github.com/aybabtme/bomberman/player"
)

// Bombs!
func placeBomb(board board.Board, game *game.Game, placerState *player.State) {
	placer := game.Players[placerState]
	log.Debugf("[%s] Attempting to place bomb (%d/%d).",
		placer.Name(), placerState.Bombs, placerState.MaxBomb)

	switch {
	case placerState.Bombs > placerState.MaxBomb:
		log.Panicf("'%s' has %d/%d bombs", placer.Name(), placerState.Bombs, placerState.MaxBomb)
	case placerState.Bombs == placerState.MaxBomb:
		log.Debugf("Failed.")
		return
	}

	placerState.Bombs++
	x, y := placerState.X, placerState.Y
	// radius is snapshot'd at this point in time
	radius := placerState.MaxRadius

	replenishBomb := func(turn int) error {
		if placerState.Bombs > 0 {
			placerState.Bombs--
		} else {
			log.Errorf("[%s] Too many bombs, %d (max %d)", placer.Name(), placerState.Bombs, placerState.MaxBomb)
		}
		return nil
	}

	doFlameout := func(turn int) error {
		log.Debugf("[%s] Bomb flameout.", placer.Name())
		removeFlame(board, x, y, radius)
		return nil
	}

	doExplosion := func(turn int) error {
		log.Debugf("[%s] Bomb exploding.", placer.Name())

		explode(game, board, x, y, radius)

		log.Debugf("[%s] Registering flameout.", placer.Name())
		game.Schedule.Register(&BomberAction{
			name:     fmt.Sprintf("%s.doFlameout", placer.Name()),
			duration: 1,
			doTurn:   doFlameout,
		}, TurnsToFlamout)

		log.Debugf("[%s] Registering bomb replenishment.", placer.Name())
		game.Schedule.Register(&BomberAction{
			name:     fmt.Sprintf("%s.replenishBomb", placer.Name()),
			duration: 1,
			doTurn:   replenishBomb,
		}, TurnsToReplenish)

		return nil
	}

	doPlaceBomb := func(turn int) error {

		board[x][y].Push(objects.Bomb)

		log.Debugf("[%s] Registering bomb explosion.", placer.Name())
		game.Schedule.Register(&BomberAction{
			name:     fmt.Sprintf("%s.doExplosion", placer.Name()),
			duration: 1,
			doTurn:   doExplosion,
		}, TurnsToExplode)
		return nil
	}

	game.Schedule.Register(&BomberAction{
		name:     fmt.Sprintf("%s.placeBomb", placer.Name()),
		duration: 1,
		doTurn:   doPlaceBomb,
	}, 1)

}

func explode(game *game.Game, board board.Board, explodeX, explodeY, radius int) {
	board[explodeX][explodeY].Remove(objects.Bomb)
	board.AsCross(explodeX, explodeY, radius, func(c *cell.Cell) bool {

		for playerState, player := range game.Players {
			x, y := playerState.X, playerState.Y
			if c.X == x && c.Y == y {
				log.Infof("[%s] Dying in explosion.", player.Name())
				playerState.Alive = false
			}
		}

		switch c.Top() {
		case objects.Wall:
		case objects.Rock:
			c.Push(objects.Flame)
			return false
		case objects.BombPU, objects.RadiusPU: // Explosions kill PowerUps
			c.Pop()
			c.Push(objects.Flame)
			return false
		default:
			c.Push(objects.Flame)
			return true
		}

		if c.Top() != objects.Wall {
			c.Push(objects.Flame)
		}

		return true
	})
}

func removeFlame(board board.Board, x, y, radius int) {
	board.AsCross(x, y, radius, func(c *cell.Cell) bool {
		if c.Top() == objects.Flame {
			c.Pop()
		}
		if c.Top() == objects.Rock {
			c.Pop()
		}
		return true
	})
}
