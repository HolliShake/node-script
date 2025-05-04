package main

import (
	"dev/types"
)

type TState struct {
	// The current state of the parser
	TI08 *types.TTyping
	TI16 *types.TTyping
	TI32 *types.TTyping
	TI64 *types.TTyping
	TNum *types.TTyping
	TStr *types.TTyping
	TBit *types.TTyping
	TNil *types.TTyping
}

func CreateState() *TState {
	// Create a new state
	state := new(TState)
	state.TI08 = types.TInt08()
	state.TI16 = types.TInt16()
	state.TI32 = types.TInt32()
	state.TI64 = types.TInt64()
	state.TNum = types.TNum()
	state.TStr = types.TStr()
	state.TBit = types.TBool()
	state.TNil = types.TVoid()
	return state
}
