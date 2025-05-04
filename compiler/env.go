package main

type TEnv struct {
	Parent  *TEnv
	symbols []TSymbol
}

func CreatEnv(parent *TEnv) *TEnv {
	env := new(TEnv)
	env.Parent = parent
	env.symbols = make([]TSymbol, 0)
	return env
}

// API:Export
func (env *TEnv) HasLocalSymbol(name string) bool {
	for i := len(env.symbols) - 1; i >= 0; i-- {
		if env.symbols[i].Name == name {
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
		if !current.HasLocalSymbol(name) {
			panic("symbol not found (" + name + ")!!!")
		}
		for i := len(current.symbols) - 1; i >= 0; i-- {
			if current.symbols[i].Name == name {
				return current.symbols[i]
			}
		}
		current = current.Parent
	}
	panic("symbol not found (" + name + ")!!!")
}

// API:Export
func (env *TEnv) AddSymbol(symbol TSymbol) {
	if env.HasLocalSymbol(symbol.Name) {
		panic("symbol already exists (" + symbol.Name + ")!!!")
	}
	env.symbols = append(env.symbols, symbol)
}
