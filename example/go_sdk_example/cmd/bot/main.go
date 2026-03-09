package main

import (
	"log"

	bombahead "github.com/N3moAhead/bombahead-go"
)

type SimpleBot struct{}

func (b *SimpleBot) GetNextMove(state *bombahead.GameState, h *bombahead.GameHelpers) bombahead.Action {
	if state == nil || state.Me == nil {
		return bombahead.DoNothing
	}
	bombahead.PrintField(state)

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
