package main

import (
	"dev/types"
	"fmt"
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

type TDelayedDefine struct {
	SrcFile TFileJob
	Node    *TAst
}

type TImportLater struct {
	Src TFileJob // The file that we need to import
	Dst TFileJob // The file that we are importing to
	Ast *TAst
}

type TForward struct {
	State          *TState
	Files          []TFileJob
	Imports        []TFileJob
	Delayed        []TDelayedImport
	MissingTypes   []TMissingTypeJob
	DelayedDefines []TDelayedDefine
	ImportLater    []TImportLater
}

// FILE

func (f *TForward) hasFile(path string) bool {
	for _, file := range f.Files {
		if file.Path == path {
			return true
		}
	}
	return false
}

func (f *TForward) getFile(path string) TFileJob {
	for i := range f.Files {
		if f.Files[i].Path == path {
			return f.Files[i]
		}
	}
	panic("file not found")
}

func (f *TForward) pushFile(file TFileJob) {
	f.Files = append(f.Files, file)
}

// IMPORT

func (f *TForward) hasImport() bool {
	return len(f.Imports) > 0
}

func (f *TForward) popImport() TFileJob {
	if !f.hasImport() {
		panic("no import to pop")
	}
	file := f.Imports[len(f.Imports)-1]
	f.Imports = f.Imports[:len(f.Imports)-1]
	return file
}

func (f *TForward) pushImport(file TFileJob) {
	f.Imports = append(f.Imports, file)
}

// DELAYED

func (f *TForward) hasDelayed() bool {
	return len(f.Delayed) > 0
}

func (f *TForward) popDelayed() TDelayedImport {
	if !f.hasDelayed() {
		panic("no delayed import to pop")
	}
	delayed := f.Delayed[len(f.Delayed)-1]
	f.Delayed = f.Delayed[:len(f.Delayed)-1]
	return delayed
}

func (f *TForward) pushDelayed(delayed TDelayedImport) {
	f.Delayed = append(f.Delayed, delayed)
}

// MISSING TYPES

func (f *TForward) hasMissingTypes() bool {
	return len(f.MissingTypes) > 0
}

func (f *TForward) popMissingTypes() TMissingTypeJob {
	if !f.hasMissingTypes() {
		panic("no missing type to pop")
	}
	missingType := f.MissingTypes[len(f.MissingTypes)-1]
	f.MissingTypes = f.MissingTypes[:len(f.MissingTypes)-1]
	return missingType
}

func (f *TForward) pushMissingTypes(missingType TMissingTypeJob) {
	f.MissingTypes = append(f.MissingTypes, missingType)
}

// DELAYED DEFINES

func (f *TForward) hasDelayedDefines() bool {
	return len(f.DelayedDefines) > 0
}

func (f *TForward) popDelayedDefines() TDelayedDefine {
	if !f.hasDelayedDefines() {
		panic("no delayed define to pop")
	}
	delayedDefine := f.DelayedDefines[len(f.DelayedDefines)-1]
	f.DelayedDefines = f.DelayedDefines[:len(f.DelayedDefines)-1]
	return delayedDefine
}

func (f *TForward) pushDelayedDefine(delayedDefine TDelayedDefine) {
	f.DelayedDefines = append(f.DelayedDefines, delayedDefine)
}

// IMPORT LATER

func (f *TForward) hasImportLater() bool {
	return len(f.ImportLater) > 0
}

func (f *TForward) popImportLater() TImportLater {
	if !f.hasImportLater() {
		panic("no import later to pop")
	}
	importLater := f.ImportLater[len(f.ImportLater)-1]
	f.ImportLater = f.ImportLater[:len(f.ImportLater)-1]
	return importLater
}

func (f *TForward) pushImportLater(importLater TImportLater) {
	f.ImportLater = append(f.ImportLater, importLater)
}

// Begin: Forward

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
			f.pushMissingTypes(TMissingTypeJob{
				file:    fileJob,
				NameAst: keyAst, // For this time, Pass the keyAst here.
				TypeAst: keyAst,
			})
		}
		if valType == nil {
			f.pushMissingTypes(TMissingTypeJob{
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
			f.pushMissingTypes(TMissingTypeJob{
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
	newEnv := CreateEnv(fileJob.Env)
	nameNode := node.Ast0
	namesNode := node.AstArr0
	typesNode := node.AstArr1
	if nameNode.Ttype != AstIDN {
		RaiseLanguageCompileError(
			fileJob.Path,
			fileJob.Data,
			"invalid struct name, struct name must be in a form of identifier",
			nameNode.Position,
		)
	}
	if len(namesNode) <= 0 {
		RaiseLanguageCompileError(
			fileJob.Path,
			fileJob.Data,
			"invalid struct, struct must have at least one attribute",
			node.Position,
		)
	}
	if fileJob.Env.HasLocalSymbol(nameNode.Str0) {
		RaiseLanguageCompileError(
			fileJob.Path,
			fileJob.Data,
			"duplicate struct name",
			nameNode.Position,
		)
	}
	attributes := make([]*types.TPair, 0)
	for i := range namesNode {
		attrN := namesNode[i]
		typeN := typesNode[i]
		if attrN.Ttype != AstIDN {
			RaiseLanguageCompileError(
				fileJob.Path,
				fileJob.Data,
				"invalid attribute name, attribute name must be in a form of identifier",
				attrN.Position,
			)
		}
		dataType := f.getType(fileJob, typeN)
		if dataType == nil {
			f.pushMissingTypes(TMissingTypeJob{
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
			NameSpace:    attrN.Str0,
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
		NameSpace:    JoinVariableName(GetFileNameWithoutExtension(fileJob.Path), nameNode.Str0),
		DataType:     types.TStruct(JoinVariableName(GetFileNameWithoutExtension(fileJob.Path), nameNode.Str0), attributes),
		Position:     node.Position,
		IsGlobal:     true,
		IsConst:      true,
		IsInitialize: true,
	})
}

func (f *TForward) forwardDefine(fileJob TFileJob, node *TAst, error bool) {
	nameNode := node.Ast0
	returnTypeNode := node.Ast1
	paramNamesNode := node.AstArr0
	paramTypesNode := node.AstArr1
	if nameNode.Ttype != AstIDN {
		RaiseLanguageCompileError(
			fileJob.Path,
			fileJob.Data,
			"invalid function name, function name must be in a form of identifier",
			nameNode.Position,
		)
	}
	if fileJob.Env.HasLocalSymbol(nameNode.Str0) {
		RaiseLanguageCompileError(
			fileJob.Path,
			fileJob.Data,
			"duplicate function name",
			nameNode.Position,
		)
	}
	parameters := make([]*types.TPair, 0)
	for index, nameNode := range paramNamesNode {
		paramTypeNode := paramTypesNode[index]
		if nameNode.Ttype != AstIDN {
			RaiseLanguageCompileError(
				fileJob.Path,
				fileJob.Data,
				"invalid parameter name, parameter name must be in a form of identifier",
				nameNode.Position,
			)
		}
		paramType := f.getType(fileJob, paramTypeNode)
		if paramType == nil {
			if error {
				RaiseLanguageCompileError(
					fileJob.Path,
					fileJob.Data,
					"missing type, or type is invalid",
					paramTypeNode.Position,
				)
			}
			f.pushDelayedDefine(TDelayedDefine{
				SrcFile: fileJob,
				Node:    node,
			})
			return
		}
		parameters = append(parameters, types.CreatePair(nameNode.Str0, paramType))
	}
	returnType := f.getType(fileJob, returnTypeNode)
	if returnType == nil {
		if error {
			RaiseLanguageCompileError(
				fileJob.Path,
				fileJob.Data,
				"missing type, or type is invalid",
				returnTypeNode.Position,
			)
		}
		f.pushDelayedDefine(TDelayedDefine{
			SrcFile: fileJob,
			Node:    node,
		})
		return
	}
	fileJob.Env.AddSymbol(TSymbol{
		Name:         nameNode.Str0,
		NameSpace:    JoinVariableName(GetFileNameWithoutExtension(fileJob.Path), nameNode.Str0),
		DataType:     types.TFunc(parameters, returnType),
		Position:     node.Position,
		IsGlobal:     true,
		IsConst:      true,
		IsInitialize: true,
	})
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
	if len(namesNode) <= 0 {
		RaiseLanguageCompileError(
			fileJob.Path,
			fileJob.Data,
			"invalid import, import must have at least one attribute",
			node.Position,
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
	if !f.hasFile(actualPath) {
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
			Env:    CreateEnv(nil),
			IsDone: true,
		}
		f.pushDelayed(TDelayedImport{
			SrcFile: fileJob,
			Node:    node,
		})
		f.pushImport(childFile)
		return
	}

	// Otherwise, kung nakita na or nag exists. kuhaa tanan property niya sa sulod

	childFile := f.getFile(actualPath)

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
			f.pushImportLater(TImportLater{
				Src: childFile,
				Dst: fileJob,
				Ast: nameNode,
			})
			continue
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
			NameSpace:    JoinVariableName(GetFileNameWithoutExtension(childFile.Path), nameNode.Str0),
			DataType:     childFile.Env.GetSymbol(nameNode.Str0).DataType,
			Position:     nameNode.Position,
			IsGlobal:     true,
			IsConst:      false,
			IsInitialize: false,
		})
	}
}

func (f *TForward) forwardVar(fileJob TFileJob, node *TAst) {
	namesNode := node.AstArr0
	typesNode := node.AstArr1
	valuNode := node.AstArr2

	for index, nameNode := range namesNode {
		typeNode := typesNode[index]
		valuNode := valuNode[index]
		if nameNode.Ttype != AstIDN {
			RaiseLanguageCompileError(
				fileJob.Path,
				fileJob.Data,
				"invalid variable name, variable name must be in a form of identifier",
				nameNode.Position,
			)
		}
		dataType := f.getType(fileJob, typeNode)
		if dataType == nil {
			f.pushMissingTypes(TMissingTypeJob{
				file:    fileJob,
				NameAst: nameNode,
				TypeAst: typeNode,
			})
			continue
		}
		if fileJob.Env.HasLocalSymbol(nameNode.Str0) {
			RaiseLanguageCompileError(
				fileJob.Path,
				fileJob.Data,
				"duplicate variable name",
				nameNode.Position,
			)
		}
		fileJob.Env.AddSymbol(TSymbol{
			Name:         nameNode.Str0,
			NameSpace:    JoinVariableName(GetFileNameWithoutExtension(fileJob.Path), nameNode.Str0),
			DataType:     dataType,
			Position:     nameNode.Position,
			IsGlobal:     true,
			IsConst:      false,
			IsInitialize: valuNode != nil,
		})
	}
}

func (f *TForward) forwardConst(fileJob TFileJob, node *TAst) {
	namesNode := node.AstArr0
	typesNode := node.AstArr1
	valuNode := node.AstArr2

	for index, nameNode := range namesNode {
		typeNode := typesNode[index]
		valuNode := valuNode[index]
		if nameNode.Ttype != AstIDN {
			RaiseLanguageCompileError(
				fileJob.Path,
				fileJob.Data,
				"invalid variable name, variable name must be in a form of identifier",
				nameNode.Position,
			)
		}
		dataType := f.getType(fileJob, typeNode)
		if dataType == nil {
			f.pushMissingTypes(TMissingTypeJob{
				file:    fileJob,
				NameAst: nameNode,
				TypeAst: typeNode,
			})
			continue
		}
		if fileJob.Env.HasLocalSymbol(nameNode.Str0) {
			RaiseLanguageCompileError(
				fileJob.Path,
				fileJob.Data,
				"duplicate variable name",
				nameNode.Position,
			)
		}
		fileJob.Env.AddSymbol(TSymbol{
			Name:         nameNode.Str0,
			NameSpace:    JoinVariableName(GetFileNameWithoutExtension(fileJob.Path), nameNode.Str0),
			DataType:     dataType,
			Position:     nameNode.Position,
			IsGlobal:     true,
			IsConst:      false,
			IsInitialize: valuNode != nil,
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
		case AstDefine:
			f.forwardDefine(fileJob, child, false)
		case AstImport:
			f.forwardImport(fileJob, child)
		case AstVar:
			f.forwardVar(fileJob, child)
		case AstConst:
			f.forwardConst(fileJob, child)
		}
	}
}

func (f *TForward) build() {
	// Build the file
	for f.hasImport() {
		importFile := f.popImport()
		f.forward(importFile)
		f.pushFile(importFile)
	}

	// Build the delayed imports
	for f.hasDelayed() {
		delayedImport := f.popDelayed()
		f.forwardImport(delayedImport.SrcFile, delayedImport.Node)
	}

	// Missing types resolution
	for f.hasMissingTypes() {
		missingType := f.popMissingTypes()
		finalType := f.getType(missingType.file, missingType.TypeAst)
		if finalType == nil {
			RaiseLanguageCompileError(
				missingType.file.Path,
				missingType.file.Data,
				"missing type, or type is invalid",
				missingType.TypeAst.Position,
			)
		}
	}

	// Delayed defines resolution
	for f.hasDelayedDefines() {
		delayedDefine := f.popDelayedDefines()
		f.forwardDefine(delayedDefine.SrcFile, delayedDefine.Node, true)
	}

	// Import later resolution
	for f.hasImportLater() {
		importLater := f.popImportLater()
		src := importLater.Src
		dst := importLater.Dst
		ast := importLater.Ast
		if !src.Env.HasLocalSymbol(ast.Str0) {
			RaiseLanguageCompileError(
				dst.Path,
				dst.Data,
				fmt.Sprintf("symbol %s not found in import %s", ast.Str0, src.Path),
				ast.Position,
			)
		}
		if dst.Env.HasLocalSymbol(ast.Str0) {
			RaiseLanguageCompileError(
				dst.Path,
				dst.Data,
				fmt.Sprintf("symbol %s already exists in import %s", ast.Str0, dst.Path),
				ast.Position,
			)
		}
		dst.Env.AddSymbol(src.Env.GetSymbol(ast.Str0))
	}
}

// API:Export
func ForwardDeclairation(state *TState, path string, data []rune, ast *TAst) []TFileJob {
	// Create a new forward declaration
	forward := new(TForward)
	forward.State = state
	forward.Files = make([]TFileJob, 0)
	forward.MissingTypes = make([]TMissingTypeJob, 0)

	job := TFileJob{
		Path:   path,
		Data:   data,
		Ast:    ast,
		Env:    CreateEnv(nil),
		IsDone: false,
	}

	forward.forward(job)
	forward.build()

	if !forward.hasFile(path) {
		forward.pushFile(job)
	}

	return forward.Files
}
