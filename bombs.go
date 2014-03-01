package main

import (
	"fmt"
)

// Bombs!
func placeBomb(board *Board, states map[*PlayerState]Player, placerState *PlayerState) {
	placer := states[placerState]
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
	radius := placerState.MaxRange

	board[x][y] = BombCell
	tmpX, tmpY := x, y

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
		removeFlame(board, tmpX, tmpY, radius)
		return nil
	}

	doExplosion := func(turn int) error {
		log.Debugf("[%s] Bomb exploding.", placer.Name())

		explode(states, board, tmpX, tmpY, radius)

		log.Debugf("[%s] Registering flameout.", placer.Name())
		schedule.Register(&BomberAction{
			name:     fmt.Sprintf("%s.doFlameout", placer.Name()),
			duration: 1,
			doTurn:   doFlameout,
		}, TurnsToFlamout)

		log.Debugf("[%s] Registering bomb replenishment.", placer.Name())
		schedule.Register(&BomberAction{
			name:     fmt.Sprintf("%s.replenishBomb", placer.Name()),
			duration: 1,
			doTurn:   replenishBomb,
		}, TurnsToReplenish)

		return nil
	}

	log.Debugf("[%s] Registering bomb explosion.", placer.Name())
	schedule.Register(&BomberAction{
		name:     fmt.Sprintf("%s.doExplosion", placer.Name()),
		duration: 1,
		doTurn:   doExplosion,
	}, TurnsToExplode)
}

func explode(states map[*PlayerState]Player, board *Board, explodeX, explodeY, radius int) {
	board[explodeX][explodeY] = GroundCell
	board.asCross(explodeX, explodeY, radius, func(cellX, cellY int) {

		for playerState, player := range states {
			x, y := playerState.X, playerState.Y
			if cellX == x && cellY == y {
				log.Infof("[%s] Dying in explosion.", player.Name())
				playerState.Alive = false
			}
		}

		if board[cellX][cellY] != WallCell {
			board[cellX][cellY] = FlameCell
		}
	})
}

func removeFlame(board *Board, x, y, radius int) {
	board.asCross(x, y, radius, func(cellX, cellY int) {
		if board[cellX][cellY] == FlameCell {
			board[cellX][cellY] = GroundCell
		}
	})
}
