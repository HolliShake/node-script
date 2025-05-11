package main

import "dev/types"

type TScopeType int

const (
	ScopeGlobal   TScopeType = iota
	ScopeLocal    TScopeType = iota
	ScopeStruct   TScopeType = iota
	ScopeFunction TScopeType = iota
	ScopeLoop     TScopeType = iota
	ScopeSingle   TScopeType = iota
)

type TScope struct {
	Parent *TScope
	Type   TScopeType
	Return *types.TTyping
}

func CreateScope(parent *TScope, scopeType TScopeType) *TScope {
	scope := new(TScope)
	scope.Parent = parent
	scope.Type = scopeType
	scope.Return = nil
	return scope
}

func (scope *TScope) InGlobal() bool {
	return scope.Type == ScopeGlobal
}

func (scope *TScope) InLocal() bool {
	current := scope
	for current != nil {
		if current.Type == ScopeLocal {
			return true
		}
		current = current.Parent
	}
	return false
}

func (scope *TScope) InStruct() bool {
	current := scope
	for current != nil {
		if current.Type == ScopeStruct {
			return true
		}
		current = current.Parent
	}
	return false
}

func (scope *TScope) InLoop() bool {
	current := scope
	for current != nil {
		if current.Type == ScopeLoop {
			return true
		}
		current = current.Parent
	}
	return false
}

func (scope *TScope) InSingle() bool {
	return scope.Type == ScopeSingle
}
