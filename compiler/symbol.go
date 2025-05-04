package main

import "dev/types"

type TSymbol struct {
	Name         string
	DataType     *types.TTyping
	Position     TPosition
	IsGlobal     bool
	IsConst      bool
	IsInitialize bool
}
