package main

import (
	"dev/types"
	"fmt"
	"strconv"
	"strings"
)

type TAnalyzer struct {
	state *TState
	file  TFileJob
	scope *TScope
	tab   int
	src   string
	stack *TEvaluationStack
}

func CreateAnalyzer(state *TState, file TFileJob) *TAnalyzer {
	analyzer := new(TAnalyzer)
	analyzer.state = state
	analyzer.file = file
	analyzer.scope = nil
	analyzer.tab = 0
	analyzer.src = ""
	analyzer.stack = CreateEvaluationStack()
	return analyzer
}

// Source UTIL
func (analyzer *TAnalyzer) incTb() {
	analyzer.tab++
}

func (analyzer *TAnalyzer) decTb() {
	analyzer.tab--
}

func (analyzer *TAnalyzer) srcTb() {
	for i := 0; i < analyzer.tab; i++ {
		analyzer.src += "\t"
	}
}

func (analyzer *TAnalyzer) srcNl() {
	analyzer.src += "\n"
}

func (analyzer *TAnalyzer) srcSp() {
	analyzer.src += " "
}

func (analyzer *TAnalyzer) write(part string, newline bool) {
	analyzer.src += part
	if newline {
		analyzer.src += "\n"
	}
}

func (analyzer *TAnalyzer) writePosition(position TPosition) {
	analyzer.write(fmt.Sprintf("//line %s:%d", analyzer.file.Path, position.SLine), true)
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
	switch node.Ttype {
	case AstIDN:
		if !analyzer.scope.Env.HasLocalSymbol(node.Str0) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"undefined symbol",
				node.Position,
			)
		}
		symbolInfo := analyzer.scope.Env.GetSymbol(node.Str0)
		analyzer.write(symbolInfo.NameSpace, false)
		analyzer.stack.Push(CreateValue(
			symbolInfo.DataType,
			nil,
		))
	case AstInt:
		i64, err := strconv.ParseInt(node.Str0, 10, 64)
		if err != nil {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("invalid integer value %s", node.Str0),
				node.Position,
			)
		}
		analyzer.write(node.Str0, false)
		switch SizeOfInt(i64) {
		case 8:
			analyzer.stack.Push(CreateValue(
				analyzer.state.TI08,
				int8(i64),
			))
		case 16:
			analyzer.stack.Push(CreateValue(
				analyzer.state.TI16,
				int16(i64),
			))
		case 32:
			analyzer.stack.Push(CreateValue(
				analyzer.state.TI32,
				int32(i64),
			))
		case 64:
			analyzer.stack.Push(CreateValue(
				analyzer.state.TI64,
				i64,
			))
		default:
			analyzer.stack.Push(CreateValue(
				analyzer.state.TI08,
				int8(i64),
			))
		}
	case AstNum:
		f64, err := strconv.ParseFloat(node.Str0, 64)
		if err != nil {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("invalid float value %s", node.Str0),
				node.Position,
			)
		}
		analyzer.write(node.Str0, false)
		analyzer.stack.Push(CreateValue(
			analyzer.state.TNum,
			f64,
		))
	case AstStr:
		analyzer.write(fmt.Sprintf("\"%s\"", node.Str0), false)
		analyzer.stack.Push(CreateValue(
			analyzer.state.TStr,
			node.Str0,
		))
	case AstBool:
		analyzer.write(node.Str0, false)
		analyzer.stack.Push(CreateValue(
			analyzer.state.TBit,
			node.Str0 == "true",
		))
	case AstNull:
		analyzer.write("nil", false)
		analyzer.stack.Push(CreateValue(
			analyzer.state.TNil,
			nil,
		))
	case AstTupleExpression:
		tupleTypes := make([]*types.TTyping, 0)
		for index, childNode := range node.AstArr0 {
			analyzer.expression(childNode)
			dataType := analyzer.stack.Pop().DataType
			tupleTypes = append(tupleTypes, dataType)
			if index < len(node.AstArr0)-1 {
				analyzer.write(", ", false)
			}
		}
		analyzer.stack.Push(CreateValue(
			types.TTuple(tupleTypes),
			nil,
		))
	case AstBindAssign:
		if analyzer.scope.InGlobal() || analyzer.scope.InSingle() {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"short variable declaration is not allowed in global scope",
				node.Position,
			)
		}
		if node.Ast0.Ttype != AstIDN && node.Ast0.Ttype != AstTupleExpression {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid variable name, variable name must be in a form of identifier",
				node.Ast0.Position,
			)
		}
		if node.Ast0.Ttype == AstIDN {
			if analyzer.file.Env.HasLocalSymbol(node.Ast0.Str0) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					"duplicate variable name",
					node.Ast0.Position,
				)
			}
			analyzer.write(node.Ast0.Str0, false)
		} else {
			for index, variableNode := range node.Ast0.AstArr0 {
				if variableNode.Ttype != AstIDN {
					RaiseLanguageCompileError(
						analyzer.file.Path,
						analyzer.file.Data,
						"invalid variable name, variable name must be in a form of identifier",
						variableNode.Position,
					)
				}
				analyzer.write(variableNode.Str0, false)
				if index < len(node.Ast0.AstArr0)-1 {
					analyzer.write(", ", false)
				}
			}
		}
		analyzer.write(" := ", false)
		analyzer.expression(node.Ast1)
		if node.Ast0.Ttype == AstIDN {
			if analyzer.scope.Env.HasLocalSymbol(node.Ast0.Str0) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					"duplicate variable name",
					node.Ast0.Position,
				)
			}
			analyzer.scope.Env.AddSymbol(TSymbol{
				Name:      node.Ast0.Str0,
				NameSpace: node.Ast0.Str0,
				DataType:  analyzer.stack.Pop().DataType,
				Position:  node.Ast0.Position,
				IsGlobal:  analyzer.scope.InGlobal(),
			})
		} else {
			mustTupleType := analyzer.stack.Pop().DataType
			if !types.IsTuple(mustTupleType) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					"cannot unpack non-tuple type",
					node.Ast1.Position,
				)
			}
			tupleTypes := mustTupleType.GetElements()
			if len(tupleTypes) != len(node.Ast0.AstArr0) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					"unpacked tuple must have the same number of elements as the number of variables",
					node.Ast1.Position,
				)
			}
			for index, variableNode := range node.Ast0.AstArr0 {
				variableType := tupleTypes[index]
				if analyzer.scope.Env.HasLocalSymbol(variableNode.Str0) {
					RaiseLanguageCompileError(
						analyzer.file.Path,
						analyzer.file.Data,
						"duplicate variable name",
						variableNode.Position,
					)
				}
				analyzer.scope.Env.AddSymbol(TSymbol{
					Name:      variableNode.Str0,
					NameSpace: variableNode.Str0,
					DataType:  variableType,
					Position:  variableNode.Position,
					IsGlobal:  analyzer.scope.InGlobal(),
				})
			}
		}
	default:
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			fmt.Sprintf("not implemented expression: %d", node.Ttype),
			node.Position,
		)
	}
}

func (analyzer *TAnalyzer) statement(node *TAst) {
	switch node.Ttype {
	case AstStruct:
		analyzer.visitStruct(node)
	case AstDefine,
		AstMethod:
		analyzer.visitDefine(node)
	case AstImport:
		analyzer.visitImport(node)
	case AstVar:
		analyzer.visitVar(node)
	case AstConst:
		analyzer.visitConst(node)
	case AstLocal:
		analyzer.visitLocal(node)
	case AstEmptyStmnt:
		analyzer.write("", false)
	case AstExpressionStmnt:
		if analyzer.scope.InGlobal() {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"expression is not allowed in here",
				node.Position,
			)
		}
		if node.Ast0.Ttype == AstIDN || IsConstantValueNode(node.Ast0) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"unused expression: this expression has no effect and its result is discarded",
				node.Position,
			)
		}
		analyzer.srcTb()
		analyzer.expression(node.Ast0)
	default:
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"not implemented statement",
			node.Position,
		)
	}
}

func (analyzer *TAnalyzer) visitStruct(node *TAst) {
	if !analyzer.scope.InGlobal() {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"struct is not allowed here",
			node.Position,
		)
	}
	analyzer.writePosition(node.Position)
	analyzer.scope = CreateScope(analyzer.scope, ScopeStruct)
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
	// Struct name must use pascal case
	if !IsPascalCase(nameNode.Str0) {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"invalid struct name, struct name must be in a form of pascal case",
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
	analyzer.write(fmt.Sprintf("type %s struct", JoinVariableName(GetFileNameWithoutExtension(analyzer.file.Path), nameNode.Str0)), false)
	analyzer.srcSp()
	analyzer.write("{", true)
	analyzer.incTb()
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
		// Attribute name must use pascal case
		if !IsPascalCase(attrNode.Str0) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid attribute name, struct attribute must be in a form of pascal case",
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
		analyzer.srcTb()
		analyzer.write(fmt.Sprintf("%s %s", attrNode.Str0, dataType.ToGoType()), false)
		if index < len(namesNode)-1 {
			analyzer.srcNl()
		}
	}
	analyzer.decTb()
	analyzer.srcNl()
	analyzer.write("}", false)
	analyzer.scope = analyzer.scope.Parent
}

func (analyzer *TAnalyzer) visitDefine(node *TAst) {
	if !analyzer.scope.InGlobal() {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"function is not allowed here",
			node.Position,
		)
	}
	analyzer.writePosition(node.Position)
	analyzer.scope = CreateScope(analyzer.scope, ScopeFunction)
	analyzer.scope = CreateScope(analyzer.scope, ScopeLocal)
	isMethod := node.Ttype == AstMethod
	nameNode := node.Ast0
	returnTypeNode := node.Ast1
	paramNamesNode := node.AstArr0
	paramTypesNode := node.AstArr1
	childrenNode := node.AstArr2
	if nameNode.Ttype != AstIDN {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			INVALID_FUNCTION_NAME,
			nameNode.Position,
		)
	}
	// Function name must use camel case
	if !isMethod && !IsCamelCase(nameNode.Str0) {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"invalid function name, function name must be in a form of camel case",
			nameNode.Position,
		)
	} else if isMethod && !IsPascalCase(nameNode.Str0) {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"invalid method name, method name must be in a form of pascal case",
			nameNode.Position,
		)
	}
	analyzer.write("func", false)
	analyzer.srcSp()
	var thisArgType *types.TTyping = nil
	if isMethod {
		analyzer.write("(", false)
		thisArgNode := paramNamesNode[0]
		thisArgTypeNode := paramTypesNode[0]
		thisArgType = analyzer.getType(thisArgTypeNode)
		analyzer.write(fmt.Sprintf("%s %s", thisArgNode.Str0, thisArgType.ToGoType()), false)
		analyzer.write(")", false)
		analyzer.srcSp()
		paramNamesNode = paramNamesNode[1:]
		paramTypesNode = paramTypesNode[1:]
		if analyzer.scope.Env.HasLocalSymbol(thisArgNode.Str0) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				INVALID_FUNCTION_PARAM_NAME_DUPLICATE,
				thisArgNode.Position,
			)
		}
		analyzer.scope.Env.AddSymbol(TSymbol{
			Name:      thisArgNode.Str0,
			NameSpace: thisArgNode.Str0,
			DataType:  analyzer.getType(thisArgTypeNode),
			Position:  thisArgNode.Position,
			IsGlobal:  analyzer.scope.InGlobal(),
		})
	}
	analyzer.write(JoinVariableName(GetFileNameWithoutExtension(analyzer.file.Path), nameNode.Str0), false)
	analyzer.write("(", false)
	parametersTypesPair := make([]*types.TPair, 0)
	for index, paramNameNode := range paramNamesNode {
		paramTypeNode := paramTypesNode[index]
		if paramNameNode.Ttype != AstIDN {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				INVALID_FUNCTION_PARAM_NAME,
				paramNameNode.Position,
			)
		}
		// Parameter name must use camel case
		if !IsCamelCase(paramNameNode.Str0) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid parameter name, parameter name must be in a form of camel case",
				paramNameNode.Position,
			)
		}
		parameterType := analyzer.getType(paramTypeNode)
		parametersTypesPair = append(parametersTypesPair, types.CreatePair(paramNameNode.Str0, parameterType))
		analyzer.write(fmt.Sprintf("%s %s", paramNameNode.Str0, parameterType.ToGoType()), false)
		if index < len(paramNamesNode)-1 {
			analyzer.srcNl()
		}
		if analyzer.scope.Env.HasLocalSymbol(paramNameNode.Str0) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				INVALID_FUNCTION_PARAM_NAME_DUPLICATE,
				paramNameNode.Position,
			)
		}
		analyzer.scope.Env.AddSymbol(TSymbol{
			Name:         paramNameNode.Str0,
			NameSpace:    paramNameNode.Str0,
			DataType:     analyzer.getType(paramTypeNode),
			Position:     paramNameNode.Position,
			IsGlobal:     analyzer.scope.InGlobal(),
			IsConst:      false,
			IsInitialize: true, // For parameters, we always initialize them.
		})
	}
	analyzer.write(")", false)
	returnType := analyzer.getType(returnTypeNode)
	analyzer.srcSp()
	analyzer.write(returnType.ToGoType(), false)
	analyzer.srcSp()
	analyzer.write("{", true)
	analyzer.incTb()
	for index, childNode := range childrenNode {
		analyzer.statement(childNode)
		if index < len(childrenNode)-1 {
			analyzer.srcNl()
		}
	}
	analyzer.decTb()
	analyzer.srcNl()
	analyzer.write("}", false)
	analyzer.scope = analyzer.scope.Parent
	analyzer.scope = analyzer.scope.Parent
	if isMethod && thisArgType != nil {
		// Save to type
		if thisArgType.HasMethod(nameNode.Str0) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"method already exists",
				nameNode.Position,
			)
		}
		thisArgType.AddMethod(
			nameNode.Str0,
			JoinVariableName(GetFileNameWithoutExtension(analyzer.file.Path), nameNode.Str0),
			types.TFunc(
				parametersTypesPair,
				returnType,
			),
		)
	}
}

// Imports are already handled by Forwarder
// so we don't need to handle them here, except for
// functions and other forms of variables.
// However, we need to write "// import" in the source code
// to make it easier to read.
func (analyzer *TAnalyzer) visitImport(node *TAst) {
	if !analyzer.scope.InGlobal() {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"import is not allowed here",
			node.Position,
		)
	}
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
	for index, nameNode := range namesNode {
		if nameNode.Ttype != AstIDN {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				INVALID_IMPORT_NAME,
				nameNode.Position,
			)
		}
		analyzer.write(fmt.Sprintf("/* import %s -> %s */", analyzer.file.Path, nameNode.Str0), false)
		if index < len(namesNode)-1 {
			analyzer.srcNl()
		}
	}
}

func (analyzer *TAnalyzer) visitVar(node *TAst) {
	if !analyzer.scope.InGlobal() {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"variable is not allowed here",
			node.Position,
		)
	}
	if analyzer.scope.InGlobal() {
		analyzer.srcTb()
		analyzer.writePosition(node.Position)
	}
	namesNode := node.AstArr0
	typesNode := node.AstArr1
	valusNode := node.AstArr2
	analyzer.write("var", false)
	analyzer.srcSp()
	analyzer.write("(", true)
	analyzer.incTb()
	for index, nameNode := range namesNode {
		typeNode := typesNode[index]
		valuNode := valusNode[index]
		if nameNode.Ttype != AstIDN {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				INVALID_VARIABLE_NAME,
				nameNode.Position,
			)
		}
		// Global variable must use pascal case
		if !IsPascalCase(nameNode.Str0) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid variable name, global variable must be in a form of pascal case",
				nameNode.Position,
			)
		}
		dataType := analyzer.getType(typeNode)
		if types.IsVoid(dataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid variable type, variable type cannot be void",
				nameNode.Position,
			)
		}
		analyzer.srcTb()
		// For "var" (global), var fileName_a int = 100;
		// For "const" (global), const fileName_a int = 100;
		// For "const" (local), var a int = 100;
		// For "local" (local), var a int = 100;
		analyzer.write(fmt.Sprintf("%s %s", JoinVariableName(GetFileNameWithoutExtension(analyzer.file.Path), nameNode.Str0), dataType.ToGoType()), false)
		if valuNode != nil {
			analyzer.write(" = ", false)
			analyzer.expression(valuNode)
			valueType := analyzer.stack.Pop().DataType
			if !types.CanStore(dataType, valueType) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					fmt.Sprintf("cannot assign %s to %s", valueType.ToString(), dataType.ToString()),
					nameNode.Position,
				)
			}
		} else {
			analyzer.write(" = ", false)
			analyzer.write(dataType.DefaultValue(), false)
		}
		if index < len(namesNode)-1 {
			analyzer.srcNl()
		}
		if analyzer.scope.InLocal() {
			// Check if the symbol already exists in the local scope.
			// If it doesn't exist in the local scope, we save it to the environment;
			// otherwise, we raise an error.
			if analyzer.scope.Env.HasLocalSymbol(nameNode.Str0) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					INVALID_VARIABLE_NAME_DUPLICATE,
					nameNode.Position,
				)
			}
			analyzer.scope.Env.AddSymbol(TSymbol{
				Name:         nameNode.Str0,
				NameSpace:    JoinVariableName(GetFileNameWithoutExtension(analyzer.file.Path), nameNode.Str0),
				DataType:     dataType,
				Position:     nameNode.Position,
				IsGlobal:     analyzer.scope.InGlobal(),
				IsConst:      false,
				IsInitialize: valuNode != nil,
			})
		}

	}
	analyzer.decTb()
	analyzer.srcNl()
	analyzer.write(")", false)
}

func (analyzer *TAnalyzer) visitConst(node *TAst) {
	if analyzer.scope.InSingle() {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"constant is not allowed here",
			node.Position,
		)
	}
	if analyzer.scope.InGlobal() {
		analyzer.srcTb()
		analyzer.writePosition(node.Position)
	}
	namesNode := node.AstArr0
	typesNode := node.AstArr1
	valusNode := node.AstArr2
	analyzer.srcTb()
	if analyzer.scope.InGlobal() {
		analyzer.write("const", false)
	} else {
		analyzer.write("var", false)
	}
	analyzer.srcSp()
	analyzer.write("(", true)
	analyzer.incTb()
	for index, nameNode := range namesNode {
		typeNode := typesNode[index]
		valuNode := valusNode[index]
		if nameNode.Ttype != AstIDN {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				INVALID_VARIABLE_NAME,
				nameNode.Position,
			)
		}
		// Global variable must use pascal case
		if analyzer.scope.InGlobal() && !IsPascalCase(nameNode.Str0) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid constant name, global constant name must be in a form of pascal case",
				nameNode.Position,
			)
		} else if !analyzer.scope.InGlobal() && !IsCamelCase(nameNode.Str0) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid constant name, local constant name must be in a form of camel case",
				nameNode.Position,
			)
		}
		dataType := analyzer.getType(typeNode)
		if types.IsVoid(dataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid constant type, constant type cannot be void",
				nameNode.Position,
			)
		}
		analyzer.srcTb()
		// For "var" (global), var fileName_a int = 100;
		// For "const" (global), const fileName_a int = 100;
		// For "const" (local), var a int = 100;
		// For "local" (local), var a int = 100;
		if analyzer.scope.InGlobal() {
			analyzer.write(fmt.Sprintf("%s %s", JoinVariableName(GetFileNameWithoutExtension(analyzer.file.Path), nameNode.Str0), dataType.ToGoType()), false)
		} else {
			analyzer.write(fmt.Sprintf("%s %s", nameNode.Str0, dataType.ToGoType()), false)
		}
		if analyzer.scope.InGlobal() && valuNode != nil && !IsConstantValueNode(valuNode) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid constant value, constant value must be a literal constant",
				nameNode.Position,
			)
		}
		if valuNode != nil {
			analyzer.write(" = ", false)
			analyzer.expression(valuNode)
			valueType := analyzer.stack.Pop().DataType
			if !types.CanStore(dataType, valueType) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					fmt.Sprintf("cannot assign %s to %s", valueType.ToString(), dataType.ToString()),
					nameNode.Position,
				)
			}
		} else {
			analyzer.write(" = ", false)
			analyzer.write(dataType.DefaultValue(), false)
		}
		if index < len(namesNode)-1 {
			analyzer.srcNl()
		}
		if analyzer.scope.InLocal() {
			// Check if the symbol already exists in the local scope.
			// If it doesn't exist in the local scope, we save it to the environment;
			// otherwise, we raise an error.
			if analyzer.scope.Env.HasLocalSymbol(nameNode.Str0) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					INVALID_VARIABLE_NAME_DUPLICATE,
					nameNode.Position,
				)
			}
			analyzer.scope.Env.AddSymbol(TSymbol{
				Name:         nameNode.Str0,
				NameSpace:    nameNode.Str0,
				DataType:     dataType,
				Position:     nameNode.Position,
				IsGlobal:     analyzer.scope.InGlobal(),
				IsConst:      true,
				IsInitialize: valuNode != nil,
			})
		}
	}
	analyzer.decTb()
	analyzer.srcNl()
	analyzer.srcTb()
	analyzer.write(")", false)
}

func (analyzer *TAnalyzer) visitLocal(node *TAst) {
	if !analyzer.scope.InLocal() || analyzer.scope.InSingle() {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"local variable is not allowed here",
			node.Position,
		)
	}
	if analyzer.scope.InGlobal() {
		analyzer.srcTb()
		analyzer.writePosition(node.Position)
	}
	namesNode := node.AstArr0
	typesNode := node.AstArr1
	valusNode := node.AstArr2
	analyzer.srcTb()
	analyzer.write("var", false)
	analyzer.srcSp()
	analyzer.write("(", true)
	analyzer.incTb()
	for index, nameNode := range namesNode {
		typeNode := typesNode[index]
		valuNode := valusNode[index]
		if nameNode.Ttype != AstIDN {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				INVALID_VARIABLE_NAME,
				nameNode.Position,
			)
		}
		// Local variable must use camel case
		if !IsCamelCase(nameNode.Str0) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid variable name, local variable must be in a form of camel case",
				nameNode.Position,
			)
		}
		dataType := analyzer.getType(typeNode)
		if types.IsVoid(dataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid variable type, variable type cannot be void",
				nameNode.Position,
			)
		}
		analyzer.srcTb()
		// For "var" (global), var fileName_a int = 100;
		// For "const" (global), const fileName_a int = 100;
		// For "const" (local), var a int = 100;
		// For "local" (local), var a int = 100;
		analyzer.write(fmt.Sprintf("%s %s", nameNode.Str0, dataType.ToGoType()), false)
		if valuNode != nil {
			analyzer.write(" = ", false)
			analyzer.expression(valuNode)
			valueType := analyzer.stack.Pop().DataType
			if !types.CanStore(dataType, valueType) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					fmt.Sprintf("cannot assign %s to %s", valueType.ToString(), dataType.ToString()),
					nameNode.Position,
				)
			}
		} else {
			analyzer.write(" = ", false)
			analyzer.write(dataType.DefaultValue(), false)
		}
		if index < len(namesNode)-1 {
			analyzer.srcNl()
		}
		if analyzer.scope.InLocal() {
			// Check if the symbol already exists in the local scope.
			// If it doesn't exist in the local scope, we save it to the environment;
			// otherwise, we raise an error.
			if analyzer.scope.Env.HasLocalSymbol(nameNode.Str0) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					INVALID_VARIABLE_NAME_DUPLICATE,
					nameNode.Position,
				)
			}
			analyzer.scope.Env.AddSymbol(TSymbol{
				Name:         nameNode.Str0,
				NameSpace:    nameNode.Str0,
				DataType:     dataType,
				Position:     nameNode.Position,
				IsGlobal:     false,
				IsConst:      false,
				IsInitialize: valuNode != nil,
			})
		}
	}
	analyzer.decTb()
	analyzer.srcNl()
	analyzer.srcTb()
	analyzer.write(")", false)
}

func (analyzer *TAnalyzer) program(node *TAst) {
	analyzer.write("package main", true)
	analyzer.scope = &TScope{
		Parent: nil,
		Env:    analyzer.file.Env,
		Type:   ScopeGlobal,
		Return: nil,
	}
	for index, child := range node.AstArr0 {
		analyzer.statement(child)
		if index < len(node.AstArr0)-1 {
			analyzer.srcNl()
		}
	}
}

// API:Export
func (analyzer *TAnalyzer) Analyze() string {
	analyzer.program(analyzer.file.Ast)
	return analyzer.src
}
