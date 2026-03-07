package classic

import (
	"github.com/N3moAhead/bombahead/server/pkg/types"
)

func (c *Classic) update() []types.Vec2 {
	// Process player inputs first to ensure their actions are part of this tick's calculations
	c.applyPlayerInput()

	// Gotta clean up the mess from last tick
	c.resetExplosions()

	destroyedBoxes := c.updateBombs()

	c.damagePlayersInExplosions()

	return destroyedBoxes
}

func (c *Classic) damagePlayersInExplosions() {
	for _, player := range c.players {
		if _, ok := c.explosions[player.Pos.String()]; ok {
			player.Health -= 1
		}
	}
}

func (c *Classic) applyPlayerInput() {
	for _, player := range c.players {
		switch player.NextMove {
		case MOVE_UP:
			newPos := player.Pos.Add(types.Vec2{X: 0, Y: -1})
			if !c.isWalkable(newPos) {
				player.Pos = newPos
			}
		case MOVE_RIGHT:
			newPos := player.Pos.Add(types.Vec2{X: 1, Y: 0})
			if !c.isWalkable(newPos) {
				player.Pos = newPos
			}
		case MOVE_DOWN:
			newPos := player.Pos.Add(types.Vec2{X: 0, Y: 1})
			if !c.isWalkable(newPos) {
				player.Pos = newPos
			}
		case MOVE_LEFT:
			newPos := player.Pos.Add(types.Vec2{X: -1, Y: 0})
			if !c.isWalkable(newPos) {
				player.Pos = newPos
			}
		case PLACE_BOMB:
			if !c.containsBomb(player.Pos) {
				newBomb := NewBomb(player.Pos)
				c.bombs[newBomb.Pos.String()] = newBomb
			}
		default:
			// Is the do nothing move
			// or the user has not defined input move
		}
	}
}

func (c *Classic) isWalkable(pos types.Vec2) bool {
	return c.field.isTileBlocked(pos.X, pos.Y) || c.containsBomb(pos)
}

func (c *Classic) updateBombs() []types.Vec2 {
	var allDestroyedBoxes []types.Vec2
	// Iterate over a copy of keys, as `explodeBomb` can modify c.bombs in a chain reaction.
	bombKeys := make([]string, 0, len(c.bombs))
	for k := range c.bombs {
		bombKeys = append(bombKeys, k)
	}

	for _, key := range bombKeys {
		bomb, exists := c.bombs[key]
		if !exists { // It might have been destroyed by another bomb in the same tick
			continue
		}

		bomb.Fuse -= 1
		if bomb.Fuse < 1 {
			delete(c.bombs, bomb.Pos.String())
			destroyedBoxes := c.explodeBomb(bomb.Pos, bomb_explosion_radius)
			allDestroyedBoxes = append(allDestroyedBoxes, destroyedBoxes...)
		}
	}
	return allDestroyedBoxes
}

func (c *Classic) resetExplosions() {
	for k := range c.explosions {
		delete(c.explosions, k)
	}
}

func (c *Classic) explodeBomb(pos types.Vec2, distance int) []types.Vec2 {
	var destroyedBoxes []types.Vec2
	// Up
	destroyedBoxes = append(destroyedBoxes, c.createExplodePath(pos, types.NewVec2(0, -1), distance)...)
	// Right
	destroyedBoxes = append(destroyedBoxes, c.createExplodePath(pos, types.NewVec2(1, 0), distance)...)
	// Down
	destroyedBoxes = append(destroyedBoxes, c.createExplodePath(pos, types.NewVec2(0, 1), distance)...)
	// Left
	destroyedBoxes = append(destroyedBoxes, c.createExplodePath(pos, types.NewVec2(-1, 0), distance)...)
	return destroyedBoxes
}

func (c *Classic) createExplodePath(pos types.Vec2, dir types.Vec2, distance int) []types.Vec2 {
	// I won't check if the explosion is out of bounds because a bomb can't be placed
	// out of bounds so it could not occur...
	if distance == 0 {
		return nil
	}
	tile := c.field.getTile(pos.X, pos.Y)
	if tile == WALL {
		return nil
	}
	if tile == BOX {
		c.field.setTile(pos.X, pos.Y, AIR)
		c.addExplosion(pos)
		return []types.Vec2{pos}
	}
	if tile == AIR {
		var destroyedBoxes []types.Vec2
		// Check if the current tile contains a bomb to trigger a chain reaction
		if c.containsBomb(pos) {
			// It is important to delete the bomb before calling `explodeBomb` to prevent infinite recursion
			delete(c.bombs, pos.String())
			// This explosion triggers another bomb
			destroyedBoxes = c.explodeBomb(pos, bomb_explosion_radius)
		}
		c.addExplosion(pos)
		// Continue the explosion path
		recursiveDestroyedBoxes := c.createExplodePath(pos.Add(dir), dir, distance-1)
		return append(destroyedBoxes, recursiveDestroyedBoxes...)
	}
	return nil
}

func (c *Classic) addExplosion(pos types.Vec2) {
	if _, ok := c.explosions[pos.String()]; !ok {
		c.explosions[pos.String()] = pos
	}
}

func (c *Classic) isGameOver() bool {
	alivePlayers := 0
	for _, player := range c.players {
		if player.Health > 0 {
			alivePlayers += 1
		}
	}

	// If there are one or less players left the
	// game is over
	return alivePlayers <= 1
}

func (c *Classic) getGameState() ClassicStatePayload {
	// Get Players
	pStates := []PlayerState{}
	for _, player := range c.players {
		pState := PlayerState{
			ID:     player.ID,
			Pos:    player.Pos,
			Health: player.Health,
			Score:  player.Score,
		}
		pStates = append(pStates, pState)
	}

	// Get Field
	field := []Tile{}
	for y := range field_height {
		for x := range field_width {
			field = append(field, c.field.getTile(x, y))
		}
	}
	fieldState := FieldState{
		Width:  field_width,
		Height: field_height,
		Field:  field,
	}
	// Get Bombs
	bombs := []BombState{}
	for _, bomb := range c.bombs {
		bombs = append(bombs, BombState{Pos: bomb.Pos, Fuse: bomb.Fuse})
	}
	// Get Explosions
	explosions := []types.Vec2{}
	for _, ePos := range c.explosions {
		explosions = append(explosions, ePos)
	}

	return ClassicStatePayload{
		Players:    pStates,
		Field:      fieldState,
		Bombs:      bombs,
		Explosions: explosions,
	}
}

func (c *Classic) resetPlayerInputs() {
	for _, player := range c.players {
		player.NextMove = NO_INPUT_DEFINED
	}
}
