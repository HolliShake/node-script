package main

import (
	"dev/types"
	"os"
	"strings"
)

type TFileJob struct {
	Path   string
	Data   []rune
	Ast    *TAst
	Env    *TEnv
	IsDone bool
}

type TDelayedImport struct {
	SrcFile TFileJob
	Node    *TAst
}

type TMissingTypeJob struct {
	file    TFileJob
	NameAst *TAst
	TypeAst *TAst
}

type TForward struct {
	State        *TState
	Files        []TFileJob
	Imports      []TFileJob
	Delayed      []TDelayedImport
	MissingTypes []TMissingTypeJob
}

// FILE

func (f *TForward) HasFile(path string) bool {
	for _, file := range f.Files {
		if file.Path == path {
			return true
		}
	}
	return false
}

func (f *TForward) GetFile(path string) TFileJob {
	for i := range f.Files {
		if f.Files[i].Path == path {
			return f.Files[i]
		}
	}
	panic("file not found")
}

func (f *TForward) PopFile() TFileJob {
	if len(f.Files) == 0 {
		panic("no file to pop")
	}
	file := f.Files[len(f.Files)-1]
	f.Files = f.Files[:len(f.Files)-1]
	return file
}

func (f *TForward) PushFile(file TFileJob) {
	f.Files = append(f.Files, file)
}

// IMPORT

func (f *TForward) HasImport() bool {
	return len(f.Imports) > 0
}

func (f *TForward) PopImport() TFileJob {
	if !f.HasImport() {
		panic("no import to pop")
	}
	file := f.Imports[len(f.Imports)-1]
	f.Imports = f.Imports[:len(f.Imports)-1]
	return file
}

func (f *TForward) PushImport(file TFileJob) {
	f.Imports = append(f.Imports, file)
}

// DELAYED

func (f *TForward) HasDelayed() bool {
	return len(f.Delayed) > 0
}

func (f *TForward) PopDelayed() TDelayedImport {
	if !f.HasDelayed() {
		panic("no delayed import to pop")
	}
	delayed := f.Delayed[len(f.Delayed)-1]
	f.Delayed = f.Delayed[:len(f.Delayed)-1]
	return delayed
}

func (f *TForward) PushDelayed(delayed TDelayedImport) {
	f.Delayed = append(f.Delayed, delayed)
}

// MISSING TYPES
func (f *TForward) HasMissingTypes() bool {
	return len(f.MissingTypes) > 0
}

func (f *TForward) PopMissingTypes() TMissingTypeJob {
	if !f.HasMissingTypes() {
		panic("no missing type to pop")
	}
	missingType := f.MissingTypes[len(f.MissingTypes)-1]
	f.MissingTypes = f.MissingTypes[:len(f.MissingTypes)-1]
	return missingType
}

func (f *TForward) PushMissingTypes(missingType TMissingTypeJob) {
	f.MissingTypes = append(f.MissingTypes, missingType)
}

func (f *TForward) getType(fileJob TFileJob, node *TAst) *types.TTyping {
	switch node.Ttype {
	case AstIDN:
		if fileJob.Env.HasLocalSymbol(node.Str0) {
			return fileJob.Env.GetSymbol(node.Str0).DataType
		}
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
		keyType := f.getType(fileJob, keyAst)
		valType := f.getType(fileJob, valAst)
		if keyType == nil {
			f.PushMissingTypes(TMissingTypeJob{
				file:    fileJob,
				NameAst: keyAst, // For this time, Pass the keyAst here.
				TypeAst: keyAst,
			})
		}
		if valType == nil {
			f.PushMissingTypes(TMissingTypeJob{
				file:    fileJob,
				NameAst: valAst, // For this time, Pass the valAst here.
				TypeAst: valAst,
			})
		}
		if keyType == nil || valType == nil {
			return nil
		}
		if !types.IsValidKey(keyType) {
			RaiseLanguageCompileError(
				fileJob.Path,
				fileJob.Data,
				"invalid hashmap key type",
				keyAst.Position,
			)
		}
		return types.THashMap(keyType, valType)
	case AstTypeArray:
		elementAst := node.Ast0
		elementType := f.getType(fileJob, elementAst)
		if elementType == nil {
			f.PushMissingTypes(TMissingTypeJob{
				file:    fileJob,
				NameAst: elementAst, // For this time, Pass the valAst here.
				TypeAst: elementAst,
			})
			return nil
		}
		return types.TArray(elementType)
	}
	return nil
}

func (f *TForward) forwardStruct(fileJob TFileJob, node *TAst) {
	newEnv := CreatEnv(fileJob.Env)
	nameNode := node.Ast0
	namesNode := node.AstArr0
	typesNode := node.AstArr1
	attributes := make([]*types.TPair, 0)
	for i := 0; i < len(namesNode); i++ {
		attrN := namesNode[i]
		typeN := typesNode[i]
		dataType := f.getType(fileJob, typeN)
		if dataType == nil {
			f.PushMissingTypes(TMissingTypeJob{
				file:    fileJob,
				NameAst: attrN,
				TypeAst: typeN,
			})
			continue
		}
		if newEnv.HasLocalSymbol(attrN.Str0) {
			RaiseLanguageCompileError(
				fileJob.Path,
				fileJob.Data,
				"duplicate attribute name",
				attrN.Position,
			)
		}
		newEnv.AddSymbol(TSymbol{
			Name:         attrN.Str0,
			DataType:     dataType,
			Position:     attrN.Position,
			IsGlobal:     false,
			IsConst:      false,
			IsInitialize: false,
		})
		attributes = append(attributes, types.CreatePair(attrN.Str0, dataType))
	}
	fileJob.Env.AddSymbol(TSymbol{
		Name:         nameNode.Str0,
		DataType:     types.TStruct(nameNode.Str0, attributes),
		Position:     node.Position,
		IsGlobal:     true,
		IsConst:      true,
		IsInitialize: true,
	})
}

func (f *TForward) forwardFunc(fileJob TFileJob, node *TAst) {

}

func (f *TForward) forwardImport(fileJob TFileJob, node *TAst) {
	pathNode := node.Ast0
	namesNode := node.AstArr0
	if pathNode.Ttype != AstStr {
		RaiseLanguageCompileError(
			fileJob.Path,
			fileJob.Data,
			"invalid import path, import path must be in a form of string",
			pathNode.Position,
		)
	}
	if !(strings.HasPrefix(pathNode.Str0, "./") || strings.HasPrefix(pathNode.Str0, "../")) {
		RaiseLanguageCompileError(
			fileJob.Path,
			fileJob.Data,
			"invalid import path, import path must be relative",
			pathNode.Position,
		)
	}

	actualPath := ResolvePath(GetDir(fileJob.Path), pathNode.Str0)

	// If wala pa nakita sa f.Files
	// E push sa pending imports (f.Imports)
	if !f.HasFile(actualPath) {
		data, err := os.ReadFile(actualPath)
		if err != nil {
			RaiseLanguageCompileError(
				fileJob.Path,
				fileJob.Data,
				"import file not found",
				pathNode.Position,
			)
		}

		parser := CreateParser(actualPath, string(data))
		ast := parser.Parse()

		childFile := TFileJob{
			Path:   actualPath,
			Data:   parser.Tokenizer.Data,
			Ast:    ast,
			Env:    CreatEnv(nil),
			IsDone: true,
		}
		f.PushDelayed(TDelayedImport{
			SrcFile: fileJob,
			Node:    node,
		})
		f.PushImport(childFile)
		return
	}

	// Otherwise, kung nakita na or nag exists. kuhaa tanan property niya sa sulod

	childFile := f.GetFile(actualPath)

	for i := range namesNode {
		nameNode := namesNode[i]
		if nameNode.Ttype != AstIDN {
			RaiseLanguageCompileError(
				fileJob.Path,
				fileJob.Data,
				"invalid import name, import name must be in a form of identifier",
				nameNode.Position,
			)
		}
		if !childFile.Env.HasLocalSymbol(nameNode.Str0) {
			RaiseLanguageCompileError(
				fileJob.Path,
				fileJob.Data,
				"imported symbol not found",
				nameNode.Position,
			)
		}
		if fileJob.Env.HasGlobalSymbol(nameNode.Str0) {
			RaiseLanguageCompileError(
				fileJob.Path,
				fileJob.Data,
				"duplicate symbol name",
				nameNode.Position,
			)
		}
		fileJob.Env.AddSymbol(TSymbol{
			Name:         nameNode.Str0,
			DataType:     childFile.Env.GetSymbol(nameNode.Str0).DataType,
			Position:     nameNode.Position,
			IsGlobal:     true,
			IsConst:      false,
			IsInitialize: false,
		})
	}
}

func (f *TForward) forward(fileJob TFileJob) {
	body := fileJob.Ast.AstArr0
	for i := range body {
		child := body[i]
		switch child.Ttype {
		case AstStruct:
			f.forwardStruct(fileJob, child)
		case AstFunc:
			f.forwardFunc(fileJob, child)
		case AstImport:
			f.forwardImport(fileJob, child)
		}
	}
}

func (f *TForward) build() {
	// Build the file
	for len(f.Imports) > 0 {
		importFile := f.PopImport()
		f.forward(importFile)
		f.PushFile(importFile)
	}

	// Build the delayed imports
	for f.HasDelayed() {
		delayedImport := f.PopDelayed()
		f.forwardImport(delayedImport.SrcFile, delayedImport.Node)
	}

	// Missing types resolution
	for f.HasImport() {
		missingType := f.PopMissingTypes()
		finalType := f.getType(missingType.file, missingType.TypeAst)
		if finalType == nil {
			RaiseLanguageCompileError(
				missingType.file.Path,
				missingType.file.Data,
				"missing type",
				missingType.TypeAst.Position,
			)
		}
	}
}

func forwardDeclairation(state *TState, path string, data []rune, ast *TAst) {
	// Create a new forward declaration
	forward := new(TForward)
	forward.State = state
	forward.Files = make([]TFileJob, 0)
	forward.MissingTypes = make([]TMissingTypeJob, 0)

	job := TFileJob{
		Path:   path,
		Data:   data,
		Ast:    ast,
		Env:    CreatEnv(nil),
		IsDone: false,
	}

	forward.forward(job)
	forward.build()
}
