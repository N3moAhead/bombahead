package classic

import (
	"math/rand"

	"github.com/N3moAhead/bombahead/server/pkg/types"
)

type Field [field_width * field_height]Tile

func NewField() *Field {
	f := Field{} // Will be initted with all air

	// Let's place some walls :)
	for x := range field_width {
		for y := range field_height {
			f.setTile(x, y, AIR) // Everything is air in the beginning
			// left or right wall
			if x == 0 || x == field_width-1 {
				f.setTile(x, y, WALL)
				continue
			}
			// top or bot wall
			if y == 0 || y == field_height-1 {
				f.setTile(x, y, WALL)
				continue
			}

			// Labyrinth Walls
			if x%2 == 0 && y%2 == 0 {
				f.setTile(x, y, WALL)
				continue
			}

			if rand.Int()%100 < box_spawn_rate {
				f.setTile(x, y, BOX)
			}
		}
	}

	spawnPoints := []types.Vec2{
		types.NewVec2(1, 1),                          // Top-Left
		types.NewVec2(field_width-2, field_height-2), // Bottom-Right
		types.NewVec2(field_width-2, 1),              // Top-Right
		types.NewVec2(1, field_height-2),             // Bottom-Left
	}

	// Clean the spawns
	for _, spawn := range spawnPoints {
		f.clearFields(spawn.X, spawn.Y, 2)
	}

	return &f
}

func (f *Field) clearFields(x, y, depth int) {
	if depth < 1 || f.getTile(x, y) == WALL {
		return
	}
	f.setTile(x, y, AIR)
	depth--
	// top
	f.clearFields(x, y-1, depth)
	// right
	f.clearFields(x+1, y, depth)
	// down
	f.clearFields(x, y+1, depth)
	// left
	f.clearFields(x-1, y, depth)
}

func (f *Field) getTile(x, y int) Tile {
	return f[y*field_height+x]
}

func (f *Field) setTile(x, y int, tile Tile) {
	f[y*field_height+x] = tile
}

func (f *Field) isTileBlocked(x, y int) bool {
	tile := f.getTile(x, y)
	if tile == WALL || tile == BOX {
		return true
	}
	return false
}
