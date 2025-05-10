package main

import (
	"dev/types"
	"fmt"
	"strings"
)

type TAnalyzer struct {
	state *TState
	file  TFileJob
	tab   int
	src   string
}

func CreateAnalyzer(state *TState, file TFileJob) *TAnalyzer {
	analyzer := new(TAnalyzer)
	analyzer.state = state
	analyzer.file = file
	analyzer.tab = 0
	analyzer.src = ""
	return analyzer
}

// Source UTIL
func (analyzer *TAnalyzer) incTab() {
	analyzer.tab++
}

func (analyzer *TAnalyzer) decTab() {
	analyzer.tab--
}

func (analyzer *TAnalyzer) sourceTab() {
	for i := 0; i < analyzer.tab; i++ {
		analyzer.src += "\t"
	}
}

func (analyzer *TAnalyzer) sourceNewline() {
	analyzer.src += "\n"
}

func (analyzer *TAnalyzer) sourceSpace() {
	analyzer.src += " "
}

func (analyzer *TAnalyzer) write(part string) {
	analyzer.src += part
}

func (analyzer *TAnalyzer) writeLine(part string) {
	analyzer.src += part + "\n"
}

func (analyzer *TAnalyzer) writePosition(position TPosition) {
	analyzer.writeLine(fmt.Sprintf("//line %s:%d", analyzer.file.Path, position.SLine))
}

// Begin

// Should not return nil.
// If it returns nil, it means the type is not found.
// So we should raise an error.
func (analyzer *TAnalyzer) getType(node *TAst) *types.TTyping {
	switch node.Ttype {
	case AstIDN:
		if analyzer.file.Env.HasLocalSymbol(node.Str0) {
			return analyzer.file.Env.GetSymbol(node.Str0).DataType
		}
	case AstTypeInt8:
		return analyzer.state.TI08
	case AstTypeInt16:
		return analyzer.state.TI16
	case AstTypeInt32:
		return analyzer.state.TI32
	case AstTypeInt64:
		return analyzer.state.TI64
	case AstTypeNum:
		return analyzer.state.TNum
	case AstTypeStr:
		return analyzer.state.TStr
	case AstTypeBool:
		return analyzer.state.TBit
	case AstTypeVoid:
		return analyzer.state.TNil
	case AstTypeArray:
		elementAst := node.Ast0
		elementType := analyzer.getType(elementAst)
		if elementType == nil {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid array element type",
				elementAst.Position,
			)
		}
		return types.TArray(elementType)
	case AstTypeHashMap:
		keyAst := node.Ast0
		valAst := node.Ast1
		keyType := analyzer.getType(keyAst)
		valType := analyzer.getType(valAst)
		if keyType == nil {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid hashmap key type",
				keyAst.Position,
			)
		}
		if keyType != nil && !types.IsValidKey(keyType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid hashmap key type",
				keyAst.Position,
			)
		}
		if valType == nil {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid hashmap value type",
				valAst.Position,
			)
		}
		if valType != nil && !types.IsStruct(valType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid hashmap value type",
				valAst.Position,
			)
		}
		return types.THashMap(keyType, valType)
	default:
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"invalid data type",
			node.Position,
		)
	}
	return nil
}

func (analyzer *TAnalyzer) expression(node *TAst) {
	// switch node.Ttype {
	// case AstTypeBinary:
	// 	analyzer.expression(node.Left)
	// 	analyzer.expression(node.Right)
	// case AstTypeUnary:
	// 	analyzer.expression(node.Child)
	// case AstTypeLiteral:
	// 	analyzer.literal(node)
	// }
}

func (analyzer *TAnalyzer) statement(node *TAst) {
	switch node.Ttype {
	case AstStruct:
		analyzer.visitStruct(node)
	case AstFunc,
		AstMethod:
		analyzer.visitFunc(node)
	case AstImport:
		analyzer.visitImport(node)
		// default:
		// 	RaiseLanguageCompileError(
		// 		analyzer.file.Path,
		// 		analyzer.file.Data,
		// 		"not implemented statement",
		// 		node.Position,
		// 	)
	}
}

func (analyzer *TAnalyzer) visitStruct(node *TAst) {
	analyzer.writePosition(node.Position)
	nameNode := node.Ast0
	namesNode := node.AstArr0
	typesNode := node.AstArr1
	if nameNode.Ttype != AstIDN {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"invalid struct name, struct name must be in a form of identifier",
			nameNode.Position,
		)
	}
	if len(namesNode) <= 0 {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"invalid struct, struct must have at least one attribute",
			node.Position,
		)
	}
	if !analyzer.file.Env.HasGlobalSymbol(nameNode.Str0) {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"undefined struct",
			nameNode.Position,
		)
	}
	thisStruct := analyzer.file.Env.GetSymbol(nameNode.Str0)
	analyzer.write(fmt.Sprintf("type %s struct", JoinVariableName(GetFileNameWithoutExtension(analyzer.file.Path), nameNode.Str0)))
	analyzer.sourceSpace()
	analyzer.writeLine("{")
	analyzer.incTab()
	for index, attrNode := range namesNode {
		typeNode := typesNode[index]
		if attrNode.Ttype != AstIDN {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid attribute name, attribute name must be in a form of identifier",
				attrNode.Position,
			)
		}
		dataType := analyzer.getType(typeNode)
		// Check for cycle member
		items := make([]*types.TTyping, 0)
		items = append(items, dataType)
		for len(items) > 0 {
			current := items[0]
			items = items[1:]
			if types.IsTheSameInstance(current, thisStruct.DataType) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					fmt.Sprintf("struct '%s' contains a cycle: member '%s' has the same type as its containing struct", thisStruct.Name, attrNode.Str0),
					attrNode.Position,
				)
			}
			if types.IsStruct(current) {
				for _, member := range current.GetMembers() {
					items = append(items, member.DataType)
				}
			}
		}
		analyzer.sourceTab()
		analyzer.write(fmt.Sprintf("%s %s", attrNode.Str0, dataType.ToGoType()))
		if index < len(namesNode)-1 {
			analyzer.sourceNewline()
		}
	}
	analyzer.decTab()
	analyzer.sourceNewline()
	analyzer.write("}")
}

func (analyzer *TAnalyzer) visitFunc(node *TAst) {
	analyzer.writePosition(node.Position)
	isMethod := node.Ttype == AstMethod
	nameNode := node.Ast0
	returnTypeNode := node.Ast1
	paramNamesNode := node.AstArr0
	paramTypesNode := node.AstArr1
	childrenNode := node.AstArr2
	analyzer.write("func")
	analyzer.sourceSpace()
	if isMethod {
		analyzer.write("(")
		thisArgNode := paramNamesNode[0]
		thisArgTypeNode := paramTypesNode[0]
		analyzer.write(fmt.Sprintf("%s %s", thisArgNode.Str0, analyzer.getType(thisArgTypeNode).ToGoType()))
		analyzer.write(")")
		analyzer.sourceSpace()
		paramNamesNode = paramNamesNode[1:]
		paramTypesNode = paramTypesNode[1:]
	}
	analyzer.write(JoinVariableName(GetFileNameWithoutExtension(analyzer.file.Path), nameNode.Str0))
	analyzer.write("(")
	for index, paramNameNode := range paramNamesNode {
		paramTypeNode := paramTypesNode[index]
		analyzer.write(fmt.Sprintf("%s %s", paramNameNode.Str0, analyzer.getType(paramTypeNode).ToGoType()))
		if index < len(paramNamesNode)-1 {
			analyzer.sourceNewline()
		}
	}
	analyzer.write(")")
	returnType := analyzer.getType(returnTypeNode)
	analyzer.sourceSpace()
	analyzer.write(returnType.ToGoType())
	analyzer.sourceSpace()
	analyzer.writeLine("{")
	analyzer.incTab()
	for index, childNode := range childrenNode {
		analyzer.statement(childNode)
		if index < len(childrenNode)-1 {
			analyzer.sourceNewline()
		}
	}
	analyzer.decTab()
	analyzer.sourceNewline()
	analyzer.write("}")
}

// Imports are already handled by Forwarder
// so we don't need to handle them here, except for
// functions and other forms of variables.
// However, we need to write "// import" in the source code
// to make it easier to read.
func (analyzer *TAnalyzer) visitImport(node *TAst) {
	analyzer.writePosition(node.Position)
	pathNode := node.Ast0
	namesNode := node.AstArr0
	if pathNode.Ttype != AstStr {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"invalid import path, import path must be in a form of string",
			pathNode.Position,
		)
	}
	if len(namesNode) <= 0 {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"invalid import, import must have at least one attribute",
			node.Position,
		)
	}
	if !(strings.HasPrefix(pathNode.Str0, "./") || strings.HasPrefix(pathNode.Str0, "../")) {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"invalid import path, import path must be relative",
			pathNode.Position,
		)
	}
	actualPath := ResolvePath(GetDir(analyzer.file.Path), pathNode.Str0)
	if !analyzer.state.HasFile(actualPath) {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"imported file not found",
			pathNode.Position,
		)
	}
	importedFile := analyzer.state.GetFile(actualPath)
	for index, nameNode := range namesNode {
		if nameNode.Ttype != AstIDN {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid import name, import name must be in a form of identifier",
				nameNode.Position,
			)
		}
		if !importedFile.Env.HasLocalSymbol(nameNode.Str0) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"imported symbol not found",
				nameNode.Position,
			)
		}
		importedSymbol := importedFile.Env.GetSymbol(nameNode.Str0)
		if !types.IsStruct(importedSymbol.DataType) {
			if analyzer.file.Env.HasGlobalSymbol(nameNode.Str0) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					"duplicate symbol name",
					nameNode.Position,
				)
			}
			analyzer.file.Env.AddSymbol(importedSymbol)
		}
		analyzer.write(fmt.Sprintf("/* import %s -> %s */", analyzer.file.Path, nameNode.Str0))
		if index < len(namesNode)-1 {
			analyzer.sourceNewline()
		}
	}
}

func (analyzer *TAnalyzer) program(node *TAst) {
	analyzer.writeLine("package main")
	for index, child := range node.AstArr0 {
		analyzer.statement(child)
		if index < len(node.AstArr0)-1 {
			analyzer.sourceNewline()
		}
	}
}

// API:Export
func (analyzer *TAnalyzer) Analyze() string {
	analyzer.program(analyzer.file.Ast)
	return analyzer.src
}
