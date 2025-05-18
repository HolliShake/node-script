package main

import "dev/types"

type TEnv struct {
	Parent  *TEnv
	Symbols []TSymbol
}

func CreateEnv(parent *TEnv) *TEnv {
	env := new(TEnv)
	env.Parent = parent
	env.Symbols = make([]TSymbol, 0)
	return env
}

// API:Export
func (env *TEnv) HasLocalSymbol(name string) bool {
	for i := len(env.Symbols) - 1; i >= 0; i-- {
		if env.Symbols[i].Name == name {
			return true
		}
	}
	return false
}

// API:Export
func (env *TEnv) HasGlobalSymbol(name string) bool {
	current := env
	for current != nil {
		if current.HasLocalSymbol(name) {
			return true
		}
		current = current.Parent
	}
	return false
}

// API:Export
func (env *TEnv) GetSymbol(name string) TSymbol {
	current := env
	for current != nil {
		for i := len(current.Symbols) - 1; i >= 0; i-- {
			if current.Symbols[i].Name == name {
				return current.Symbols[i]
			}
		}
		current = current.Parent
	}
	RaiseSystemError("symbol not found (" + name + ")!!!")
	return TSymbol{}
}

// API:Export
func (env *TEnv) AddSymbol(symbol TSymbol) {
	if env.HasLocalSymbol(symbol.Name) {
		RaiseSystemError("symbol already exists (" + symbol.Name + ")!!!")
	}
	env.Symbols = append(env.Symbols, symbol)
}

func (env *TEnv) UpdateSymbolIsUsed(name string, isUsed bool) {
	current := env
	for current != nil {
		for i := range current.Symbols {
			if current.Symbols[i].Name == name {
				current.Symbols[i].IsUsed = isUsed
				return
			}
		}
		current = current.Parent
	}
}

// Used for defining global constants|variables|functions|classes|interfaces|enums|structs|etc.
func DefineSymbol(env *TEnv, name string, namespace string, module string, dataType *types.TTyping) {
	if env.HasLocalSymbol(name) {
		RaiseSystemError("symbol already exists (" + name + ")!!!")
	}
	env.Symbols = append(env.Symbols, TSymbol{
		Name:         name,
		NameSpace:    namespace,
		Module:       module,
		DataType:     dataType,
		Position:     TPosition{},
		IsGlobal:     true,
		IsConst:      true,
		IsUsed:       false,
		IsInitialize: true,
	})
}
