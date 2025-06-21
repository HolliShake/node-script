package main

import (
	"dev/types"
)

type TState struct {
	// The current state of the parser
	Files      []TFileJob
	TI08       *types.TTyping
	TI16       *types.TTyping
	TI32       *types.TTyping
	TI64       *types.TTyping
	TNum       *types.TTyping
	TStr       *types.TTyping
	TBit       *types.TTyping
	TNil       *types.TTyping
	TErr       *types.TTyping
	ArrayTypes []*types.TTyping // Array of types
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
	state.TErr = types.TError()
	state.ArrayTypes = make([]*types.TTyping, 0)
	return state
}

func (state *TState) SetFile(files []TFileJob) {
	state.Files = files
}

func (state *TState) HasFile(path string) bool {
	for _, file := range state.Files {
		if file.Path == path {
			return true
		}
	}
	return false
}

func (state *TState) GetFile(path string) TFileJob {
	for _, file := range state.Files {
		if file.Path == path {
			return file
		}
	}
	RaiseSystemError("file not found")
	return TFileJob{}
}

func (state *TState) ArrayTypeExists(t *types.TTyping) bool {
	for _, arrayType := range state.ArrayTypes {
		if arrayType.ToNormalName() == t.ToNormalName() {
			return true
		}
	}
	return false
}

func (state *TState) AddArrayType(t *types.TTyping) {
	state.ArrayTypes = append(state.ArrayTypes, t)
}

func (state *TState) GenerateArrays() string {
	code := "package main"
	code += "\n\n"
	code += "import ("
	code += "\n\t\"strings\""
	code += "\n\t\"fmt\""
	code += "\n)"
	code += "\n\n"
	for _, arrayType := range state.ArrayTypes {
		code += GenerateArrayCode(arrayType)
		code += "\n\n"
	}

	return code
}
