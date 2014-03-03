package main

import (
	"fmt"
	"github.com/aybabtme/bomberman/cell"
)

// Bombs!
func placeBomb(board *Board, game *Game, placerState *PlayerState) {
	placer := game.players[placerState]
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

	board[x][y].Push(BombObj)

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
		game.schedule.Register(&BomberAction{
			name:     fmt.Sprintf("%s.doFlameout", placer.Name()),
			duration: 1,
			doTurn:   doFlameout,
		}, TurnsToFlamout)

		log.Debugf("[%s] Registering bomb replenishment.", placer.Name())
		game.schedule.Register(&BomberAction{
			name:     fmt.Sprintf("%s.replenishBomb", placer.Name()),
			duration: 1,
			doTurn:   replenishBomb,
		}, TurnsToReplenish)

		return nil
	}

	log.Debugf("[%s] Registering bomb explosion.", placer.Name())
	game.schedule.Register(&BomberAction{
		name:     fmt.Sprintf("%s.doExplosion", placer.Name()),
		duration: 1,
		doTurn:   doExplosion,
	}, TurnsToExplode)
}

func explode(game *Game, board *Board, explodeX, explodeY, radius int) {
	board[explodeX][explodeY].Remove(BombObj)
	board.asCross(explodeX, explodeY, radius, func(c *cell.Cell) bool {

		for playerState, player := range game.players {
			x, y := playerState.X, playerState.Y
			if c.X == x && c.Y == y {
				log.Infof("[%s] Dying in explosion.", player.Name())
				playerState.Alive = false
			}
		}

		switch c.Top() {
		case WallObj:
		case RockObj:
			c.Push(FlameObj)
			return false
		case BombPUObj, RadiusPUObj: // Explosions kill PowerUps
			c.Pop()
			c.Push(FlameObj)
			return false
		default:
			c.Push(FlameObj)
			return true
		}

		if c.Top() != WallObj {
			c.Push(FlameObj)
		}

		return true
	})
}

func removeFlame(board *Board, x, y, radius int) {
	board.asCross(x, y, radius, func(c *cell.Cell) bool {
		if c.Top() == FlameObj {
			c.Pop()
		}
		if c.Top() == RockObj {
			c.Pop()
		}
		return true
	})
}
