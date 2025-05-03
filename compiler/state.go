package main

type TState struct {
	// The current state of the parser
	i int
}

func CreateState() *TState {
	// Create a new state
	state := new(TState)
	return state
}
