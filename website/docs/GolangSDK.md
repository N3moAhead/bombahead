# bombahead-go

Go SDK for building Bomberman bots that connect to a Bombahead game server over WebSocket.

## Installation

```
go get github.com/N3moAhead/bombahead-go
```

## Quick Start

1. Implement the `bombahead.Bot` interface.
2. Call `bombahead.Run(yourBot)` from `main`.
4. Run your program.

## Core API

### Run

```go
func Run(userBot Bot)
```

Starts the bot client loop:

- Connects to server.
- Marks player ready.
- Receives game state messages.
- Builds `GameHelpers`.
- Calls `userBot.GetNextMove(state, helpers)`.
- Sends returned action to server.
- Re-readies automatically after `back_to_lobby`.

This function blocks until connection closes or a fatal runtime error occurs.

### Bot

```go
type Bot interface {
    GetNextMove(state *GameState, helpers *GameHelpers) Action
}
```

Implement this interface to provide your bot logic.

### NewGameHelpers

```go
func NewGameHelpers(state *GameState) *GameHelpers
```

Creates helper utilities bound to the current game state.

## Types and Models

### Action

```go
type Action string
```

Possible actions:

- `MoveUp`
- `MoveDown`
- `MoveLeft`
- `MoveRight`
- `PlaceBomb`
- `DoNothing`

### CellType

```go
type CellType string
```

Possible values:

- `Air`
- `Wall`
- `Box`

### Position

```go
type Position struct {
    X int
    Y int
}
```

Methods:

- `DistanceTo(other Position) int`: Manhattan distance between two positions.

### Player

```go
type Player struct {
    ID     string
    Pos    Position
    Health int
    Score  int
}
```

### Bomb

```go
type Bomb struct {
    Pos  Position
    Fuse int
}
```

### Field

```go
type Field struct {
    Width  int
    Height int
    Cells  []CellType
}
```

Methods:

- `CellAt(pos Position) CellType`: Returns cell at `pos`, or `Wall` when out of bounds.

### GameState

```go
type GameState struct {
    CurrentTick int
    Me          *Player
    Opponents   []Player
    Players     []Player
    Field       Field
    Bombs       []Bomb
    Explosions  []Position
}
```

Represents all data your bot receives for one tick.

## GameHelpers API

`GameHelpers` provides utility functions for pathing and safety checks.

```go
type GameHelpers struct {
    State *GameState
}
```

### IsWalkable

```go
func (h *GameHelpers) IsWalkable(pos Position) bool
```

Returns `true` if position:

- Is inside board bounds.
- Is not a `Wall`.
- Is not a `Box`.
- Is not currently occupied by a bomb.

### GetAdjacentWalkablePositions

```go
func (h *GameHelpers) GetAdjacentWalkablePositions(pos Position) []Position
```

Returns walkable adjacent cells in this fixed order:

- Up
- Right
- Down
- Left

### GetNextActionTowards

```go
func (h *GameHelpers) GetNextActionTowards(start, target Position) Action
```

Uses BFS pathfinding and returns the next movement action from `start` toward `target`.

- Returns `DoNothing` if `start == target`.
- Returns `DoNothing` if no valid path exists.

### IsSafe

```go
func (h *GameHelpers) IsSafe(pos Position) bool
```

Returns `false` when:

- Position is out of bounds.
- Position currently has a bomb.
- Position is in active explosion cells.
- Position is in predicted blast range of bombs that will trigger now (`Fuse <= 1`) or by chain reaction.

### GetNearestSafePosition

```go
func (h *GameHelpers) GetNearestSafePosition(start Position) Position
```

Finds closest safe and walkable cell with BFS.

- Returns `start` if it is already safe.
- Returns `start` if no safe reachable cell exists.

### FindNearestBox

```go
func (h *GameHelpers) FindNearestBox(start Position) (Position, bool)
```

Finds nearest reachable `Box` tile.

- BFS can traverse `Air` and `Box`.
- BFS does not traverse through `Wall`.
- Returns `found=false` when no box is reachable.

## Complete Minimal Bot Example

This example:

- Escapes danger first.
- Moves toward nearest box.
- Places a bomb when standing next to a box.
- Otherwise idles.

```go
package main

import (
	"log"

	"github.com/N3moAhead/bombahead-go"
)

type SimpleBot struct{}

func (b *SimpleBot) GetNextMove(state *bombahead.GameState, h *bombahead.GameHelpers) bombahead.Action {
	if state == nil || state.Me == nil {
		return bombahead.DoNothing
	}

	me := state.Me.Pos

	if !h.IsSafe(me) {
		safe := h.GetNearestSafePosition(me)
		return h.GetNextActionTowards(me, safe)
	}

	boxPos, found := h.FindNearestBox(me)
	if found {
		dist := me.DistanceTo(boxPos)
		if dist == 1 {
			return bombahead.PlaceBomb
		}
		if dist > 1 {
			return h.GetNextActionTowards(me, boxPos)
		}
	}

	return bombahead.DoNothing
}

func main() {
	log.Println("Starting SimpleBot...")
	bombahead.Run(&SimpleBot{})
}
```

## Suggested Project Layout

```text
my-bot/
  go.mod
  cmd/
    mybot/
      main.go
```

## Error Handling Notes

- `Run` logs and exits on connection failures or send failures.
- Some malformed server payloads are logged and skipped, allowing loop continuation.
- Unknown message types are ignored by design.

## Compatibility

- Module: `github.com/N3moAhead/bombahead-go`
- Go version from this SDK: `go 1.23`
