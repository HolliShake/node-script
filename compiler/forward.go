package main

import (
	"dev/types"
)

type TMissingTypeJob struct {
	NameAst *TAst
	TypeAst *TAst
}

type TForward struct {
	State        *TState
	Path         string
	Data         []rune
	Ast          *TAst
	MissingTypes []TMissingTypeJob
}

func (f *TForward) getType(node *TAst) *types.TTyping {
	switch node.Ttype {
	case AstTypeInt8:
		return f.State.TI08
	case AstTypeInt16:
		return f.State.TI16
	case AstTypeInt32:
		return f.State.TI32
	case AstTypeInt64:
		return f.State.TI64
	case AstTypeNum:
		return f.State.TNum
	case AstTypeStr:
		return f.State.TStr
	case AstTypeBool:
		return f.State.TBit
	case AstTypeVoid:
		return f.State.TNil
	case AstTypeHashMap:
		keyAst := node.Ast0
		valAst := node.Ast1
		keyType := f.getType(keyAst)
		valType := f.getType(valAst)
		if keyType == nil {
			f.MissingTypes = append(f.MissingTypes, TMissingTypeJob{
				NameAst: keyAst, // For this time, Pass the keyAst here.
				TypeAst: keyAst,
			})
		}
		if valType == nil {
			f.MissingTypes = append(f.MissingTypes, TMissingTypeJob{
				NameAst: valAst, // For this time, Pass the valAst here.
				TypeAst: valAst,
			})
		}
		if keyType == nil || valType == nil {
			return nil
		}
		if !types.IsValidKey(keyType) {
			RaiseLanguageCompileError(
				f.Path,
				f.Data,
				"invalid hashmap key type",
				keyAst.Position,
			)
		}
		return types.THashMap(keyType, valType)
	case AstTypeArray:
		elementAst := node.Ast0
		elementType := f.getType(elementAst)
		if elementType == nil {
			f.MissingTypes = append(f.MissingTypes, TMissingTypeJob{
				NameAst: elementAst, // For this time, Pass the valAst here.
				TypeAst: elementAst,
			})
			return nil
		}
		return types.TArray(elementType)
	}
	return nil
}

func (f *TForward) forwardStruct(node *TAst) {
	namesNode := node.AstArr0
	typesNode := node.AstArr1
	for i := 0; i < len(namesNode); i++ {
		attrN := namesNode[i]
		typeN := typesNode[i]
		dataType := f.getType(typeN)
		if dataType == nil {
			f.MissingTypes = append(f.MissingTypes, TMissingTypeJob{
				NameAst: attrN,
				TypeAst: typeN,
			})
		}
	}
}

func (f *TForward) forwardFunc(node *TAst) {

}

func (f *TForward) forwardImport(node *TAst) {

}

func (f *TForward) forward() {
	body := f.Ast.AstArr0
	for i := 0; i < len(body); i++ {
		child := body[i]
		switch child.Ttype {
		case AstStruct:
			f.forwardStruct(child)
		case AstFunc:
			f.forwardFunc(child)
		case AstImport:
			f.forwardImport(child)
		}
	}
	// For missing type resolution
	for len(f.MissingTypes) > 0 {
		top := f.MissingTypes[len(f.MissingTypes)-1]
		f.MissingTypes = f.MissingTypes[:len(f.MissingTypes)-1]
		if f.getType(top.TypeAst) == nil {
			RaiseLanguageCompileError(
				f.Path,
				f.Data,
				"data type not found",
				top.TypeAst.Position,
			)
		}
	}
}

func forwardDeclairation(state *TState, path string, data []rune, ast *TAst) {
	// Create a new forward declaration
	forward := &TForward{
		State:        state,
		Path:         path,
		Data:         data,
		Ast:          ast,
		MissingTypes: make([]TMissingTypeJob, 0),
	}
	forward.forward()
}
