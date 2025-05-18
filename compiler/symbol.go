package main

import "dev/types"

type TSymbol struct {
	Name         string
	NameSpace    string
	Module       string
	DataType     *types.TTyping
	Position     TPosition
	IsGlobal     bool
	IsConst      bool
	IsUsed       bool
	IsInitialize bool
}
