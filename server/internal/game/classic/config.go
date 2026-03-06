package classic

import "time"

const (
	// --- Field ---
	field_width    = 11
	field_height   = 11
	box_spawn_rate = 75

	// --- Bombs ---
	fuse_ticks            = 10
	bomb_explosion_radius = 3 // center + 2 fields in each direction

	// --- Game ---
	MIN_PLAYERS      = 2
	MAX_PLAYERS      = 4
	TICK_RATE        = 200 * time.Millisecond
	WIN_SCORE_POINTS = 250
	MAX_GAME_TIME    = 3 * time.Minute

	// --- Player ---
	initial_health = 3
)
