package main

import (
	"dev/types"
	"fmt"
	"strconv"
	"strings"
)

type TAnalyzer struct {
	state   *TState
	file    TFileJob
	scope   *TScope
	tab     int
	src     string
	stack   *TEvaluationStack
	modules []string
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

func (analyzer *TAnalyzer) addModule(module string) {
	for _, m := range analyzer.modules {
		if m == module {
			return
		}
	}
	analyzer.modules = append(analyzer.modules, module)
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
		if analyzer.file.Env.HasGlobalSymbol(node.Str0) {
			symbol := analyzer.file.Env.GetSymbol(node.Str0)
			analyzer.file.Env.UpdateSymbolIsUsed(node.Str0, true)
			if types.IsStruct(symbol.DataType) {
				return types.ToInstance(symbol.DataType)
			}
			return symbol.DataType
		}
		panic("not implemented")
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
	case AstTypeError:
		return analyzer.state.TErr
	case AstTypeVoid:
		return analyzer.state.TVoid
	case AstTypeTuple:
		elementTypes := make([]*types.TTyping, 0)
		for _, elementAst := range node.AstArr0 {
			elementType := analyzer.getType(elementAst)
			elementTypes = append(elementTypes, elementType)
		}
		return types.TTuple(elementTypes)
	case AstTypeArray:
		elementAst := node.Ast0
		elementType := analyzer.getType(elementAst)
		if elementType == nil {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				INVALID_ARRAY_ELEMENT_TYPE,
				elementAst.Position,
			)
		}
		if !analyzer.state.ArrayTypeExists(elementType) {
			analyzer.state.AddArrayType(elementType)
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
				INVALID_HASHMAP_KEY_TYPE,
				keyAst.Position,
			)
		}
		if keyType != nil && !types.IsValidKey(keyType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				INVALID_HASHMAP_KEY_TYPE,
				keyAst.Position,
			)
		}
		if valType == nil {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				INVALID_HASHMAP_VALUE_TYPE,
				valAst.Position,
			)
		}
		if valType != nil && types.IsVoid(valType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				INVALID_HASHMAP_VALUE_TYPE,
				valAst.Position,
			)
		}
		if !analyzer.state.MapTypeExists(keyType, valType) {
			analyzer.state.AddMapType(keyType, valType)
		}
		return types.THashMap(keyType, valType)
	case AstTypePointer:
		elementAst := node.Ast0
		elementType := analyzer.getType(elementAst)
		if elementType == nil {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid pointer element type",
				elementAst.Position,
			)
		}
		return types.ToPointer(elementType)
	case AstTypeFunc:
		argumentTypes := make([]*types.TPair, 0)
		for index, argumentAst := range node.AstArr0 {
			argumentType := analyzer.getType(argumentAst)
			if argumentType == nil {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					"invalid function argument type",
					argumentAst.Position,
				)
			}
			argumentTypes = append(argumentTypes, types.CreatePair(fmt.Sprintf("$%d", index), argumentType))
		}
		returnType := analyzer.getType(node.Ast0)
		if returnType == nil {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"missing return type",
				node.Position,
			)
		}
		return types.TFunc(false, argumentTypes, returnType, false)
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

func (analyzer *TAnalyzer) expressionAssignLeft(node *TAst) {
	switch node.Ttype {
	case AstIDN:
		if !analyzer.scope.Env.HasGlobalSymbol(node.Str0) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("undefined symbol: %s", node.Str0),
				node.Position,
			)
		}
		symbol := analyzer.scope.Env.GetSymbol(node.Str0)
		if symbol.IsConst {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot assign to constant symbol: %s", node.Str0),
				node.Position,
			)
		}
		if types.IsStruct(symbol.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot assign to struct symbol: %s", node.Str0),
				node.Position,
			)
		}
		analyzer.scope.Env.UpdateSymbolIsUsed(node.Str0, true)
		analyzer.write(symbol.NameSpace, false)
		analyzer.stack.Push(CreateValue(
			symbol.DataType,
			nil,
		))
	case AstMember:
		objectNode := node.Ast0
		memberNode := node.Ast1
		if memberNode.Ttype != AstIDN {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"member name must be an identifier",
				memberNode.Position,
			)
		}
		analyzer.expression(objectNode)
		objectType := analyzer.stack.Pop().DataType
		if !objectType.HasMember(memberNode.Str0) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("object %s has no member %s", objectType.ToString(), memberNode.Str0),
				memberNode.Position,
			)
		}
		analyzer.write(".", false)
		analyzer.write(memberNode.Str0, false)
		member := objectType.GetMember(memberNode.Str0)
		analyzer.stack.Push(CreateValue(
			member.DataType,
			nil,
		))
	case AstIndex:
		objectNode := node.Ast0
		indexNode := node.Ast1
		analyzer.expression(objectNode)
		objectType := analyzer.stack.Pop().DataType
		var elementType *types.TTyping = nil
		if types.IsArr(objectType) {
			elementType = objectType.GetInternal0()
			analyzer.write(".", false)
			analyzer.write("elements", false)
		} else if types.IsMap(objectType) {
			elementType = objectType.GetInternal1()
			analyzer.write(".", false)
			analyzer.write("elements", false)
		} else {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot assign to %s", objectType.ToString()),
				node.Position,
			)
		}
		analyzer.write("[", false)
		analyzer.expression(indexNode)
		indexType := analyzer.stack.Pop().DataType
		if types.IsArr(indexType) && !types.IsAnyInt(indexType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("array index must be an integer, got %s", indexType.ToString()),
				node.Position,
			)
		} else if types.IsMap(objectType) && !types.IsTheSameInstance(indexType, objectType.GetInternal0()) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("map index must be %s, got %s", objectType.GetInternal0().ToString(), indexType.ToString()),
				node.Position,
			)
		}
		analyzer.write("]", false)
		analyzer.stack.Push(CreateValue(
			elementType,
			nil,
		))
	default:
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"invalid left-hand side of assignment",
			node.Position,
		)
	}
}

func (analyzer *TAnalyzer) expression(node *TAst) {
	switch node.Ttype {
	case AstIDN:
		if !analyzer.scope.Env.HasGlobalSymbol(node.Str0) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("undefined symbol: %s", node.Str0),
				node.Position,
			)
		}
		symbol := analyzer.scope.Env.GetSymbol(node.Str0)
		analyzer.scope.Env.UpdateSymbolIsUsed(node.Str0, true)
		analyzer.write(symbol.NameSpace, false)
		// Struct cannot be used as a value.
		if types.IsStruct(symbol.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("struct %s cannot be used as a value", symbol.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			symbol.DataType,
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
	case AstArray:
		saveSrc := analyzer.src
		saveStack := analyzer.stack
		analyzer.src = ""
		analyzer.stack = CreateEvaluationStack()
		elementsNode := node.AstArr0
		var elementType *types.TTyping = nil // Default type.
		for index, childNode := range elementsNode {
			analyzer.expression(childNode)
			topType := analyzer.stack.Pop().DataType
			if elementType != nil {
				if !types.CanStore(elementType, topType) {
					RaiseLanguageCompileError(
						analyzer.file.Path,
						analyzer.file.Data,
						fmt.Sprintf("cannot store %s in array of [%s]", topType.ToString(), elementType.ToString()),
						childNode.Position,
					)
				}
				elementType = types.WhichBigger(elementType, topType)
			} else {
				elementType = topType
			}
			if index < len(elementsNode)-1 {
				analyzer.write(", ", false)
			}
		}
		// Restore
		analyzer.src = saveSrc
		analyzer.stack = saveStack
		analyzer.write(GetArrayConstructor(elementType), false)
		analyzer.write("(", false)
		analyzer.write("[]", false)
		analyzer.write(elementType.ToGoType(), false)
		analyzer.write("{", false)
		for index, childNode := range elementsNode {
			analyzer.expression(childNode)
			actualType := analyzer.stack.Pop().DataType
			if !types.CanStore(elementType, actualType) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					fmt.Sprintf("cannot store %s in array of [%s]", actualType.ToString(), elementType.ToString()),
					childNode.Position,
				)
			}
			if index < len(elementsNode)-1 {
				analyzer.write(", ", false)
			}
		}
		analyzer.write("}", false)
		analyzer.write(")", false)
		if !analyzer.state.ArrayTypeExists(elementType) {
			analyzer.state.AddArrayType(elementType)
		}
		analyzer.stack.Push(CreateValue(
			types.TArray(elementType),
			nil,
		))
	case AstHashMap:
		keysNode := node.AstArr0
		valuesNode := node.AstArr1
		var keyType *types.TTyping = nil
		var valueType *types.TTyping = nil
		// Save
		saveSrc := analyzer.src
		analyzer.src = ""
		for index, keyNode := range keysNode {
			valueNode := valuesNode[index]
			analyzer.expression(keyNode)
			newKeyType := analyzer.stack.Pop().DataType
			if keyType == nil {
				keyType = newKeyType
			} else if !types.CanStore(keyType, newKeyType) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					fmt.Sprintf("cannot store %s in map of [%s]", newKeyType.ToString(), keyType.ToString()),
					keyNode.Position,
				)
			}
			analyzer.write(":", false)
			analyzer.expression(valueNode)
			newValueType := analyzer.stack.Pop().DataType
			if valueType == nil {
				valueType = newValueType
			} else if !types.CanStore(valueType, newValueType) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					fmt.Sprintf("cannot store %s in map of [%s]", newValueType.ToString(), valueType.ToString()),
					valueNode.Position,
				)
			}
			if index < len(keysNode)-1 {
				analyzer.write(", ", false)
			}
		}
		analyzer.src = saveSrc
		// Finalize
		analyzer.write(GetMapConstructor(keyType, valueType), false)
		analyzer.write("(", false)
		analyzer.write("map", false)
		analyzer.write("[", false)
		analyzer.write(keyType.ToGoType(), false)
		analyzer.write("]", false)
		analyzer.write(valueType.ToGoType(), false)
		analyzer.write("{", false)
		for index, keyNode := range keysNode {
			valueNode := valuesNode[index]
			analyzer.expression(keyNode)
			newKeyType := analyzer.stack.Pop().DataType
			if keyType == nil {
				keyType = newKeyType
			} else if !types.CanStore(keyType, newKeyType) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					fmt.Sprintf("cannot store %s in map of [%s]", newKeyType.ToString(), keyType.ToString()),
					keyNode.Position,
				)
			}
			analyzer.write(":", false)
			analyzer.expression(valueNode)
			newValueType := analyzer.stack.Pop().DataType
			if valueType == nil {
				valueType = newValueType
			} else if !types.CanStore(valueType, newValueType) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					fmt.Sprintf("cannot store %s in map of [%s]", newValueType.ToString(), valueType.ToString()),
					valueNode.Position,
				)
			}
			if index < len(keysNode)-1 {
				analyzer.write(", ", false)
			}
		}
		analyzer.write("}", false)
		analyzer.write(")", false)
		if !analyzer.state.MapTypeExists(keyType, valueType) {
			analyzer.state.AddMapType(keyType, valueType)
		}
		analyzer.stack.Push(CreateValue(
			types.THashMap(keyType, valueType),
			nil,
		))
	case AstFunction:
		functionScope := CreateFunctionScope(analyzer.scope, node.Flg0)
		localScope := CreateScope(functionScope, ScopeLocal)
		analyzer.scope = functionScope
		analyzer.scope = localScope
		panics := node.Flg0
		returnTypeNode := node.Ast1
		paramNamesNode := node.AstArr0
		paramTypesNode := node.AstArr1
		childrenNode := node.AstArr2
		analyzer.write("func", false)
		analyzer.srcSp()
		analyzer.write("(", false)
		parametersTypesPair := make([]*types.TPair, 0, len(paramNamesNode))
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
				analyzer.write(", ", false)
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
				Module:       "",
				DataType:     analyzer.getType(paramTypeNode),
				Position:     paramNameNode.Position,
				IsGlobal:     analyzer.scope.InGlobal(),
				IsConst:      false,
				IsUsed:       true,
				IsInitialize: true, // For parameters, we always initialize them.
			})
		}
		analyzer.write(")", false)
		returnType := analyzer.getType(returnTypeNode)
		analyzer.srcSp()
		analyzer.write(returnType.ToGoType(), false)
		analyzer.srcSp()
		analyzer.write("{", false)
		analyzer.incTb()
		for index, childNode := range childrenNode {
			analyzer.statement(childNode)
			if index < len(childrenNode)-1 {
				analyzer.srcNl()
			}
		}
		if functionScope.Return == nil {
			if len(childrenNode) > 0 {
				analyzer.srcNl()
			}
			analyzer.srcTb()
			analyzer.write("return", false)
			analyzer.srcSp()
			analyzer.write(returnType.DefaultValue(), false)
		} else if !types.CanStore(returnType, functionScope.Return) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot return %s, return type must be %s", functionScope.Return.ToString(), returnType.ToString()),
				node.Position,
			)
		}
		if functionScope.Panics && !functionScope.HasPanic {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"function declared to panic, but it does not actually panic",
				node.Position,
			)
		}
		analyzer.decTb()
		analyzer.srcNl()
		analyzer.write("}", false)
		analyzer.scope = analyzer.scope.Parent
		analyzer.scope = analyzer.scope.Parent
		// Check if there are any unused variables.
		env := localScope.Env
		for _, symbol := range env.Symbols {
			if !symbol.IsUsed {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					fmt.Sprintf("unused variable: %s", symbol.Name),
					symbol.Position,
				)
			}
		}
		dataType := types.TFunc(false, parametersTypesPair, returnType, panics)
		analyzer.stack.Push(CreateValue(
			dataType,
			nil,
		))
	case AstPlus2:
		analyzer.expressionAssignLeft(node.Ast0)
		leftType := analyzer.stack.Pop().DataType
		analyzer.write("++", false)
		if !types.IsAnyNumber(leftType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot increment %s", leftType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			nil,
			nil,
		))
	case AstMinus2:
		analyzer.expressionAssignLeft(node.Ast0)
		leftType := analyzer.stack.Pop().DataType
		analyzer.write("--", false)
		if !types.IsAnyNumber(leftType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot decrement %s", leftType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			nil,
			nil,
		))
	case AstMember:
		objectNode := node.Ast0
		memberNode := node.Ast1
		if memberNode.Ttype != AstIDN {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"member name must be an identifier",
				memberNode.Position,
			)
		}
		analyzer.expression(objectNode)
		objectValue := analyzer.stack.Pop()
		if !objectValue.DataType.HasMember(memberNode.Str0) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("object %s has no member %s", objectValue.DataType.ToString(), memberNode.Str0),
				objectNode.Position,
			)
		}
		analyzer.write(".", false)
		analyzer.write(memberNode.Str0, false)
		member := objectValue.DataType.GetMember(memberNode.Str0)
		analyzer.stack.Push(CreateValue(
			member.DataType,
			nil,
		))
	case AstIndex:
		objectNode := node.Ast0
		indexNode := node.Ast1
		analyzer.expression(objectNode)
		objectType := analyzer.stack.Pop().DataType
		var elementType *types.TTyping = nil
		isArrayOrMap := types.IsArr(objectType) || types.IsMap(objectType)
		if types.IsArr(objectType) {
			elementType = objectType.GetInternal0()
			analyzer.write(".", false)
			analyzer.write("Get", false)
		} else if types.IsMap(objectType) {
			elementType = objectType.GetInternal1()
			analyzer.write(".", false)
			analyzer.write("Get", false)
		} else if types.IsStr(objectType) {
			elementType = analyzer.state.TI08
		} else {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot assign to %s", objectType.ToString()),
				node.Position,
			)
		}
		if isArrayOrMap {
			analyzer.write("(", false)
		} else {
			analyzer.write("[", false)
		}
		analyzer.expression(indexNode)
		indexType := analyzer.stack.Pop().DataType
		if types.IsArr(objectType) && !types.IsAnyInt(indexType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("array index must be an integer, got %s", indexType.ToString()),
				node.Position,
			)
		} else if types.IsMap(objectType) && !types.IsTheSameInstance(indexType, objectType.GetInternal0()) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("map index must be %s, got %s", objectType.GetInternal0().ToString(), indexType.ToString()),
				node.Position,
			)
		} else if types.IsStr(objectType) && !types.IsAnyInt(indexType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("string index must be an integer, got %s", indexType.ToString()),
				node.Position,
			)
		}
		if isArrayOrMap {
			analyzer.write(")", false)
		} else {
			analyzer.write("]", false)
		}
		analyzer.stack.Push(CreateValue(
			elementType,
			nil,
		))
	case AstCall:
		objectNode := node.Ast0
		parametersNode := node.AstArr0
		if objectNode.Ttype == AstMember {
			member_obj := objectNode.Ast0
			member_name := objectNode.Ast1
			if member_name.Ttype != AstIDN {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					"member name must be an identifier",
					member_name.Position,
				)
			}
			analyzer.expression(member_obj)
			member_obj_value := analyzer.stack.Pop()
			if !member_obj_value.DataType.HasMethod(member_name.Str0) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					fmt.Sprintf("object %s has no method %s", member_obj_value.DataType.ToString(), member_name.Str0),
					member_name.Position,
				)
			}
			analyzer.write(".", false)
			method := member_obj_value.DataType.GetMethod(member_name.Str0)
			analyzer.write(method.Namespace, false)
			analyzer.stack.Push(CreateValue(
				method.DataType,
				nil,
			))
		} else {
			analyzer.expression(objectNode)
		}
		objectValue := analyzer.stack.Pop()
		if !types.IsFunc(objectValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot call %s", objectValue.DataType.ToString()),
				objectNode.Position,
			)
		}
		if objectValue.DataType.Panics() && analyzer.scope.InFunction() {
			current := analyzer.scope
			for current.Type != ScopeFunction {
				current = current.Parent
			}
			if current.Type == ScopeFunction && !current.Panics {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					fmt.Sprintf("cannot call function '%s' that may panic from a function that does not declare 'panics'", objectValue.DataType.ToString()),
					objectNode.Position,
				)
			}
			current.HasPanic = true
		} else if objectValue.DataType.Panics() && analyzer.scope.InGlobal() {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"cannot call function that may panic from the global scope",
				objectNode.Position,
			)
		}
		analyzer.write("(", false)
		members := objectValue.DataType.GetMembers()
		requiredParameters := members
		if objectValue.DataType.Variadic() {
			requiredParameters = requiredParameters[:len(requiredParameters)-1]
		}
		if !objectValue.DataType.Variadic() && len(requiredParameters) == len(parametersNode) {
			for index, childNode := range parametersNode {
				requiredType := requiredParameters[index].DataType
				analyzer.expression(childNode)
				actualType := analyzer.stack.Pop().DataType
				if !types.CanStore(requiredType, actualType) {
					RaiseLanguageCompileError(
						analyzer.file.Path,
						analyzer.file.Data,
						fmt.Sprintf("expected %s, got %s", requiredType.ToString(), actualType.ToString()),
						childNode.Position,
					)
				}
				if index < len(parametersNode)-1 {
					analyzer.write(", ", false)
				}
			}
		} else if objectValue.DataType.Variadic() && len(requiredParameters) < len(parametersNode) {
			theVariadictParmeter := members[len(members)-1]
			for index, childNode := range parametersNode {
				analyzer.expression(childNode)
				if index < len(requiredParameters) {
					requiredType := requiredParameters[index].DataType
					actualType := analyzer.stack.Pop().DataType
					if !types.CanStore(requiredType, actualType) {
						RaiseLanguageCompileError(
							analyzer.file.Path,
							analyzer.file.Data,
							fmt.Sprintf("expected %s, got %s", requiredType.ToString(), actualType.ToString()),
							childNode.Position,
						)
					}
				} else {
					top := analyzer.stack.Pop()
					if !types.CanStore(theVariadictParmeter.DataType, top.DataType) {
						RaiseLanguageCompileError(
							analyzer.file.Path,
							analyzer.file.Data,
							fmt.Sprintf("expected %s, got %s", theVariadictParmeter.DataType.ToString(), top.DataType.ToString()),
							childNode.Position,
						)
					}
				}
				if index < len(parametersNode)-1 {
					analyzer.write(", ", false)
				}
			}
		} else {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("expected %d parameters, got %d", len(requiredParameters), len(parametersNode)),
				objectNode.Position,
			)
		}

		analyzer.write(")", false)
		analyzer.stack.Push(CreateValue(
			objectValue.DataType.GetReturnType(),
			nil,
		))
	case AstStruct:
		objectNode := node.Ast0
		namesNode := node.AstArr0
		valuesNode := node.AstArr1
		if objectNode.Ttype != AstIDN {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"struct name must be an identifier",
				objectNode.Position,
			)
		}
		if len(namesNode) != len(valuesNode) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"struct name and values must have the same length",
				objectNode.Position,
			)
		}
		if !analyzer.scope.Env.HasGlobalSymbol(objectNode.Str0) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("struct %s not found", objectNode.Str0),
				objectNode.Position,
			)
		}
		objectInfo := analyzer.scope.Env.GetSymbol(objectNode.Str0)
		objDataType := objectInfo.DataType
		if !types.IsStruct(objDataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("struct %s is not a struct", objectNode.Str0),
				objectNode.Position,
			)
		}
		analyzer.write(objectInfo.NameSpace, false)
		analyzer.srcSp()
		analyzer.write("{", false)
		for index, childNode := range namesNode {
			if childNode.Ttype != AstIDN {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					"struct name must be an identifier",
					childNode.Position,
				)
			}
			if !objDataType.HasMember(childNode.Str0) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					fmt.Sprintf("struct %s has no member %s", objectNode.Str0, childNode.Str0),
					childNode.Position,
				)
			}
			analyzer.write(childNode.Str0, false)
			analyzer.write(":", false)
			analyzer.srcSp()
			memberType := objDataType.GetMember(childNode.Str0).DataType
			analyzer.expression(valuesNode[index])
			actualType := analyzer.stack.Pop().DataType
			if !types.CanStore(memberType, actualType) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					fmt.Sprintf("cannot store %s in %s", actualType.ToString(), memberType.ToString()),
					childNode.Position,
				)
			}
			if index < len(namesNode)-1 {
				analyzer.write(", ", false)
			}
		}
		analyzer.write("}", false)
		analyzer.stack.Push(CreateValue(
			types.ToInstance(objDataType),
			nil,
		))
	case AstIf:
		conditionNode := node.Ast0
		bodyNode := node.Ast1
		elseBodyNode := node.Ast2
		// Save src
		saveSrc := analyzer.src
		analyzer.src = ""
		analyzer.expression(bodyNode)
		expectedType := analyzer.stack.Pop().DataType
		// Restore
		analyzer.src = saveSrc
		analyzer.write("(", false)
		analyzer.write("func()", false)
		analyzer.srcSp()
		analyzer.write(expectedType.ToGoType(), false)
		analyzer.srcSp()
		analyzer.write("{", true)
		analyzer.incTb()
		analyzer.write("if", false)
		analyzer.srcSp()
		analyzer.expression(conditionNode)
		analyzer.stack.Pop()
		analyzer.srcSp()
		analyzer.write("{", true)
		analyzer.incTb()
		analyzer.srcTb()
		analyzer.write("return", false)
		analyzer.srcSp()
		analyzer.expression(bodyNode)
		analyzer.stack.Pop()
		analyzer.srcNl()
		analyzer.decTb()
		analyzer.write("}", false)
		analyzer.write("else", false)
		analyzer.srcSp()
		analyzer.write("{", true)
		analyzer.incTb()
		analyzer.srcTb()
		analyzer.write("return", false)
		analyzer.srcSp()
		analyzer.expression(elseBodyNode)
		elseType := analyzer.stack.Pop().DataType
		analyzer.srcNl()
		analyzer.decTb()
		analyzer.write("}", false)
		analyzer.write("}", false)
		analyzer.write(")", false)
		analyzer.write("()", false)
		if !types.CanStore(expectedType, elseType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("else branch must return %s, got %s", expectedType.ToString(), elseType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			expectedType,
			nil,
		))
	case AstPlus:
		analyzer.write("+", false)
		analyzer.expression(node.Ast0)
		value := analyzer.stack.Pop()
		analyzer.stack.Push(CreateValue(
			value.DataType,
			nil,
		))
	case AstMinus:
		analyzer.write("-", false)
		analyzer.expression(node.Ast0)
		value := analyzer.stack.Pop()
		analyzer.stack.Push(CreateValue(
			value.DataType,
			nil,
		))
	case AstNot:
		analyzer.write("!", false)
		analyzer.expression(node.Ast0)
		value := analyzer.stack.Pop()
		if !types.IsBool(value.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"expected bool, got "+value.DataType.ToString(),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			analyzer.state.TBit,
			nil,
		))
	case AstBitNot:
		analyzer.write("^", false)
		analyzer.expression(node.Ast0)
		value := analyzer.stack.Pop()
		analyzer.stack.Push(CreateValue(
			value.DataType,
			nil,
		))
	case AstAllocation:
		objectNode := node.Ast0
		src := analyzer.src
		analyzer.src = ""
		analyzer.expression(objectNode)
		value := analyzer.stack.Pop()
		if !types.IsStructInstance(value.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"allocation expression must be a struct",
				objectNode.Position,
			)
		}
		analyzer.src = src
		if !value.DataType.HasConstructor() {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"struct has no constructor",
				objectNode.Position,
			)
		}
		analyzer.write(fmt.Sprintf("new_%s", value.DataType.GoTypePure(true)), false)
		analyzer.write("(", false)
		analyzer.expression(objectNode)
		analyzer.stack.Pop()
		analyzer.write(")", false)
		analyzer.stack.Push(CreateValue(
			types.ToPointer(value.DataType),
			nil,
		))
	case AstMul:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" * ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic("*", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot multiply %s and %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			types.WhichBigger(lhsValue.DataType, rhsValue.DataType),
			nil,
		))
	case AstDiv:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" / ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic("/", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot divide %s by %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			types.WhichBigger(lhsValue.DataType, rhsValue.DataType),
			nil,
		))
	case AstMod:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" % ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic("%", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot modulo %s by %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			types.WhichBigger(lhsValue.DataType, rhsValue.DataType),
			nil,
		))
	case AstAdd:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" + ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic("+", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot add %s and %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			types.WhichBigger(lhsValue.DataType, rhsValue.DataType),
			nil,
		))
	case AstSub:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" - ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic("-", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot subtract %s and %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			types.WhichBigger(lhsValue.DataType, rhsValue.DataType),
			nil,
		))
	case AstShl:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" << ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic("<<", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot shift left %s by %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			types.WhichBigger(lhsValue.DataType, rhsValue.DataType),
			nil,
		))
	case AstShr:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" >> ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic(">>", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot shift right %s by %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			types.WhichBigger(lhsValue.DataType, rhsValue.DataType),
			nil,
		))
	case AstLt:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" < ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic("<", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot compare %s and %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			analyzer.state.TBit,
			nil,
		))
	case AstLe:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" <= ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic("<=", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot compare %s and %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			analyzer.state.TBit,
			nil,
		))
	case AstGt:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" > ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic(">", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot compare %s and %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			analyzer.state.TBit,
			nil,
		))
	case AstGe:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" >= ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic(">=", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot compare %s and %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			analyzer.state.TBit,
			nil,
		))
	case AstEq:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" == ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic("==", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot compare %s and %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			analyzer.state.TBit,
			nil,
		))
	case AstNe:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" != ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic("!=", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot compare %s and %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			analyzer.state.TBit,
			nil,
		))
	case AstAnd:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" && ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic("&&", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot compare %s and %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			types.WhichBigger(lhsValue.DataType, rhsValue.DataType),
			nil,
		))
	case AstOr:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" | ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic("|", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot compare %s and %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			analyzer.state.TI64,
			nil,
		))
	case AstXor:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" ^ ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic("^", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot compare %s and %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			types.WhichBigger(lhsValue.DataType, rhsValue.DataType),
			nil,
		))
	case AstLogAnd:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" && ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic("&&", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot compare %s and %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			analyzer.state.TBit,
			nil,
		))
	case AstLogOr:
		lhsNode := node.Ast0
		rhsNode := node.Ast1
		analyzer.expression(lhsNode)
		lhsValue := analyzer.stack.Pop()
		analyzer.write(" || ", false)
		analyzer.expression(rhsNode)
		rhsValue := analyzer.stack.Pop()
		if !types.CanDoArithmetic("||", lhsValue.DataType, rhsValue.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot compare %s and %s", lhsValue.DataType.ToString(), rhsValue.DataType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			analyzer.state.TBit,
			nil,
		))
	case AstAssign:
		analyzer.expressionAssignLeft(node.Ast0)
		leftType := analyzer.stack.Pop().DataType
		analyzer.write(" = ", false)
		analyzer.expression(node.Ast1)
		rightType := analyzer.stack.Pop().DataType
		if !types.CanStore(leftType, rightType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot assign %s to %s", rightType.ToString(), leftType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			nil,
			nil,
		))
	case AstMulAssign:
		analyzer.expressionAssignLeft(node.Ast0)
		leftType := analyzer.stack.Pop().DataType
		analyzer.write(" *= ", false)
		analyzer.expression(node.Ast1)
		rightType := analyzer.stack.Pop().DataType
		if !types.CanDoArithmetic("*", leftType, rightType) || !types.CanStore(leftType, rightType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot assign %s to %s", rightType.ToString(), leftType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			nil,
			nil,
		))
	case AstDivAssign:
		analyzer.expressionAssignLeft(node.Ast0)
		leftType := analyzer.stack.Pop().DataType
		analyzer.write(" /= ", false)
		analyzer.expression(node.Ast1)
		rightType := analyzer.stack.Pop().DataType
		if !types.CanDoArithmetic("/", leftType, rightType) || !types.CanStore(leftType, rightType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot assign %s to %s", rightType.ToString(), leftType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			nil,
			nil,
		))
	case AstModAssign:
		analyzer.expressionAssignLeft(node.Ast0)
		leftType := analyzer.stack.Pop().DataType
		analyzer.write(" %= ", false)
		analyzer.expression(node.Ast1)
		rightType := analyzer.stack.Pop().DataType
		if !types.CanDoArithmetic("%", leftType, rightType) || !types.CanStore(leftType, rightType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot assign %s to %s", rightType.ToString(), leftType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			nil,
			nil,
		))
	case AstAddAssign:
		analyzer.expressionAssignLeft(node.Ast0)
		leftType := analyzer.stack.Pop().DataType
		analyzer.write(" += ", false)
		analyzer.expression(node.Ast1)
		rightType := analyzer.stack.Pop().DataType
		if !types.CanDoArithmetic("+", leftType, rightType) || !types.CanStore(leftType, rightType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot assign %s to %s", rightType.ToString(), leftType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			nil,
			nil,
		))
	case AstSubAssign:
		analyzer.expressionAssignLeft(node.Ast0)
		leftType := analyzer.stack.Pop().DataType
		analyzer.write(" -= ", false)
		analyzer.expression(node.Ast1)
		rightType := analyzer.stack.Pop().DataType
		if !types.CanDoArithmetic("-", leftType, rightType) || !types.CanStore(leftType, rightType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot assign %s to %s", rightType.ToString(), leftType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			nil,
			nil,
		))
	case AstShlAssign:
		analyzer.expressionAssignLeft(node.Ast0)
		leftType := analyzer.stack.Pop().DataType
		analyzer.write(" <<= ", false)
		analyzer.expression(node.Ast1)
		rightType := analyzer.stack.Pop().DataType
		if !types.CanDoArithmetic("<<", leftType, rightType) || !types.CanStore(leftType, rightType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot assign %s to %s", rightType.ToString(), leftType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			nil,
			nil,
		))
	case AstShrAssign:
		analyzer.expressionAssignLeft(node.Ast0)
		leftType := analyzer.stack.Pop().DataType
		analyzer.write(" >>= ", false)
		analyzer.expression(node.Ast1)
		rightType := analyzer.stack.Pop().DataType
		if !types.CanDoArithmetic(">>", leftType, rightType) || !types.CanStore(leftType, rightType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot assign %s to %s", rightType.ToString(), leftType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			nil,
			nil,
		))
	case AstAndAssign:
		analyzer.expressionAssignLeft(node.Ast0)
		leftType := analyzer.stack.Pop().DataType
		analyzer.write(" &= ", false)
		analyzer.expression(node.Ast1)
		rightType := analyzer.stack.Pop().DataType
		if !types.CanDoArithmetic("&", leftType, rightType) || !types.CanStore(leftType, rightType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot assign %s to %s", rightType.ToString(), leftType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			nil,
			nil,
		))
	case AstOrAssign:
		analyzer.expressionAssignLeft(node.Ast0)
		leftType := analyzer.stack.Pop().DataType
		analyzer.write(" |= ", false)
		analyzer.expression(node.Ast1)
		rightType := analyzer.stack.Pop().DataType
		if !types.CanDoArithmetic("|", leftType, rightType) || !types.CanStore(leftType, rightType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot assign %s to %s", rightType.ToString(), leftType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			nil,
			nil,
		))
	case AstXorAssign:
		analyzer.expressionAssignLeft(node.Ast0)
		leftType := analyzer.stack.Pop().DataType
		analyzer.write(" ^= ", false)
		analyzer.expression(node.Ast1)
		rightType := analyzer.stack.Pop().DataType
		if !types.CanDoArithmetic("^", leftType, rightType) || !types.CanStore(leftType, rightType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("cannot assign %s to %s", rightType.ToString(), leftType.ToString()),
				node.Position,
			)
		}
		analyzer.stack.Push(CreateValue(
			nil,
			nil,
		))
	case AstBindAssign:
		// Short variable declarations are not allowed in global or single scopes
		if analyzer.scope.InGlobal() || analyzer.scope.InSingle() {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"short variable declaration is not allowed in global scope",
				node.Position,
			)
		}

		// Validate that the left side is either an identifier or tuple expression
		if node.Ast0.Ttype != AstIDN && node.Ast0.Ttype != AstTupleExpression {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid variable name, variable name must be in a form of identifier",
				node.Ast0.Position,
			)
		}

		// Handle single variable declaration
		if node.Ast0.Ttype == AstIDN {
			// Check for duplicate variable names early
			if analyzer.scope.Env.HasLocalSymbol(node.Ast0.Str0) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					"duplicate variable namesss",
					node.Ast0.Position,
				)
			}
			analyzer.write(node.Ast0.Str0, false)
		} else {
			// Handle tuple unpacking
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

		// Write the assignment operator and evaluate the right-hand expression
		analyzer.write(" := ", false)
		analyzer.expression(node.Ast1)

		// Register the variable in the symbol table
		if node.Ast0.Ttype == AstIDN {
			// Double-check for duplicate variable names
			if analyzer.scope.Env.HasLocalSymbol(node.Ast0.Str0) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					fmt.Sprintf("duplicate variable name: %s", node.Ast0.Str0),
					node.Ast0.Position,
				)
			}

			// Add the symbol to the environment
			analyzer.scope.Env.AddSymbol(TSymbol{
				Name:         node.Ast0.Str0,
				NameSpace:    node.Ast0.Str0,
				DataType:     analyzer.stack.Pop().DataType,
				Position:     node.Ast0.Position,
				IsGlobal:     analyzer.scope.InGlobal(),
				IsConst:      false,
				IsUsed:       false,
				IsInitialize: true,
			})
		} else {
			// Handle tuple unpacking
			mustTupleType := analyzer.stack.Pop().DataType

			// Verify that the right-hand side is a tuple
			if !types.IsTuple(mustTupleType) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					"cannot unpack non-tuple type",
					node.Ast1.Position,
				)
			}

			// Get the tuple elements
			tupleTypes := mustTupleType.GetElements()

			// Verify that the number of variables matches the tuple size
			if len(tupleTypes) != len(node.Ast0.AstArr0) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					"unpacked tuple must have the same number of elements as the number of variables",
					node.Ast1.Position,
				)
			}

			// Register each variable from the tuple
			for index, variableNode := range node.Ast0.AstArr0 {
				variableType := tupleTypes[index]

				// Check for duplicate variable names
				if analyzer.scope.Env.HasLocalSymbol(variableNode.Str0) {
					RaiseLanguageCompileError(
						analyzer.file.Path,
						analyzer.file.Data,
						fmt.Sprintf("duplicate variable name: %s", variableNode.Str0),
						variableNode.Position,
					)
				}

				// Add the symbol to the environment
				analyzer.scope.Env.AddSymbol(TSymbol{
					Name:         variableNode.Str0,
					NameSpace:    variableNode.Str0,
					DataType:     variableType,
					Position:     variableNode.Position,
					IsGlobal:     analyzer.scope.InGlobal(),
					IsConst:      false,
					IsUsed:       false,
					IsInitialize: true,
				})
			}
		}
		analyzer.stack.Push(CreateValue(
			nil,
			nil,
		))
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
	case AstFunction,
		AstMethod:
		analyzer.visitFunction(node)
	case AstImport:
		analyzer.visitImport(node)
	case AstVar:
		analyzer.visitVar(node)
	case AstConst:
		analyzer.visitConst(node)
	case AstLocal:
		analyzer.visitLocal(node)
	case AstIf:
		analyzer.visitIf(node)
	case AstRunStmnt:
		analyzer.visitRunStmnt(node)
	case AstReturnStmnt:
		analyzer.visitReturn(node)
	case AstEmptyStmnt:
		analyzer.write("", false)
	case AstExpressionStmnt:
		// Expression statements are not allowed in global scope
		if analyzer.scope.InGlobal() {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"expression is not allowed in here",
				node.Position,
			)
		}
		analyzer.writePosition(node.Position)
		analyzer.srcTb()
		analyzer.expression(node.Ast0)
		value := analyzer.stack.Pop()
		// Prevent standalone identifiers and constants that have no effect
		if IsNoEffectValueNode(node.Ast0) &&
			!types.IsVoid(value.DataType) &&
			!types.IsTuple(value.DataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"unused expression: this expression has no effect and its result is discarded",
				node.Position,
			)
		}
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
	structName := JoinVariableName(GetFileNameWithoutExtension(analyzer.file.Path), nameNode.Str0)
	analyzer.write(fmt.Sprintf("type %s struct", structName), false)
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
		items := make([]*types.TTyping, 0, 8) // Pre-allocate capacity for better performance
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
				members := current.GetMembers()
				if len(members) > 0 {
					// Ensure capacity before appending
					if cap(items)-len(items) < len(members) {
						newItems := make([]*types.TTyping, len(items), len(items)+len(members))
						copy(newItems, items)
						items = newItems
					}
					for _, member := range members {
						items = append(items, member.DataType)
					}
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

	// Create a constructor for the struct
	analyzer.srcNl()
	analyzer.write(fmt.Sprintf("func new_%s(instance %s) *%s", structName, structName, structName), false)
	analyzer.srcSp()
	analyzer.write("{", true)
	analyzer.incTb()
	analyzer.srcTb()
	analyzer.write(fmt.Sprintf("newInstance := new(%s)", structName), true)
	for _, attrNode := range namesNode {
		analyzer.srcTb()
		analyzer.write(fmt.Sprintf("newInstance.%s = instance.%s", attrNode.Str0, attrNode.Str0), true)
	}
	analyzer.srcTb()
	analyzer.write("return newInstance", true)
	analyzer.decTb()
	analyzer.write("}", false)

	// Create a String method for the struct
	analyzer.srcNl()
	analyzer.write(fmt.Sprintf("func (instance %s) String() string", structName), false)
	analyzer.srcSp()
	analyzer.write("{", true)
	analyzer.incTb()
	analyzer.srcTb()
	analyzer.write("str := \"\"", true)
	analyzer.srcTb()
	analyzer.write(fmt.Sprintf("str += \"%s \"", nameNode.Str0), true)
	analyzer.srcTb()
	analyzer.write("str += \"{ \"", true)
	analyzer.incTb()
	for index, attrNode := range namesNode {
		analyzer.srcTb()
		analyzer.write(fmt.Sprintf("str += fmt.Sprintf(\"%s: %%v\", instance.%s)", attrNode.Str0, attrNode.Str0), true)
		if index < len(namesNode)-1 {
			analyzer.write("str += \", \"", true)
		}
	}
	analyzer.decTb()
	analyzer.srcTb()
	analyzer.write("str += \" }\"", true)
	analyzer.srcTb()
	analyzer.write("return str", true)
	analyzer.decTb()
	analyzer.write("}", false)
}

func (analyzer *TAnalyzer) visitFunction(node *TAst) {
	if !analyzer.scope.InGlobal() {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"function is not allowed here",
			node.Position,
		)
	}
	analyzer.writePosition(node.Position)
	functionScope := CreateFunctionScope(analyzer.scope, node.Flg0)
	localScope := CreateScope(functionScope, ScopeLocal)
	analyzer.scope = functionScope
	analyzer.scope = localScope
	isMethod := node.Ttype == AstMethod
	panics := node.Flg0
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
			Name:         thisArgNode.Str0,
			NameSpace:    thisArgNode.Str0,
			Module:       "",
			DataType:     analyzer.getType(thisArgTypeNode),
			Position:     thisArgNode.Position,
			IsGlobal:     analyzer.scope.InGlobal(),
			IsConst:      true,
			IsUsed:       true,
			IsInitialize: true, // For parameters, we always initialize them.
		})
	}
	analyzer.write(JoinVariableName(GetFileNameWithoutExtension(analyzer.file.Path), nameNode.Str0), false)
	analyzer.write("(", false)
	parametersTypesPair := make([]*types.TPair, 0, len(paramNamesNode))
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
			analyzer.write(", ", false)
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
			Module:       "",
			DataType:     analyzer.getType(paramTypeNode),
			Position:     paramNameNode.Position,
			IsGlobal:     analyzer.scope.InGlobal(),
			IsConst:      false,
			IsUsed:       true,
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
	if functionScope.Return == nil {
		if len(childrenNode) > 0 {
			analyzer.srcNl()
		}
		analyzer.srcTb()
		analyzer.write("return", false)
		analyzer.srcSp()
		analyzer.write(returnType.DefaultValue(), false)
	} else if !types.CanStore(returnType, functionScope.Return) {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			fmt.Sprintf("cannot return %s, return type must be %s", functionScope.Return.ToString(), returnType.ToString()),
			node.Position,
		)
	}
	if functionScope.Panics && !functionScope.HasPanic {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"function declared to panic, but it does not actually panic",
			node.Position,
		)
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
				fmt.Sprintf("method '%s' already exists for type %s", nameNode.Str0, thisArgType.ToString()),
				nameNode.Position,
			)
		}
		thisArgType.AddMethod(
			nameNode.Str0,
			JoinVariableName(GetFileNameWithoutExtension(analyzer.file.Path), nameNode.Str0),
			types.TFunc(
				false,
				parametersTypesPair,
				returnType,
				panics,
			),
		)
	}
	// Check if there are any unused variables.
	env := localScope.Env
	for _, symbol := range env.Symbols {
		if !symbol.IsUsed {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("unused variable: %s", symbol.Name),
				symbol.Position,
			)
		}
	}
}

// Imports are already handled by Forwarder,
// so we don't need to handle them here, except for
// functions and other forms of variables.
// However, we need to write "// import" in the source code
// to make it easier to read.
func (analyzer *TAnalyzer) visitImport(node *TAst) {
	// Imports are already handled by Forwarder
	// so we don't need to handle them here, except for
	// functions and other forms of variables.
	// However, we need to write "// import" in the source code
	// to make it easier to read.
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

	// Validate import path format
	if pathNode.Ttype != AstStr {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"invalid import path, import path must be in a form of string",
			pathNode.Position,
		)
	}

	// Ensure import has at least one attribute
	if len(namesNode) == 0 {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"invalid import, import must have at least one attribute",
			node.Position,
		)
	}

	if strings.HasPrefix(pathNode.Str0, "go:") {
		// Make sure this was handled by Forwarder
		pkg := pathNode.Str0[3:]
		analyzer.addModule(fmt.Sprintf("\"%s\"", pkg))
		asTypes := make([]struct {
			dataType *types.TTyping
			name     string
		}, 0)
		asVars := make([]struct {
			dataType *types.TTyping
			name     string
		}, 0)

		for _, nameNode := range namesNode {
			if nameNode.Ttype != AstIDN {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					INVALID_IMPORT_NAME,
					nameNode.Position,
				)
			}
			if !analyzer.scope.Env.HasLocalSymbol(nameNode.Str0) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					fmt.Sprintf("symbol %s not found in import", nameNode.Str0),
					nameNode.Position,
				)
			}
			symbol := analyzer.scope.Env.GetSymbol(nameNode.Str0)
			if types.IsStruct(symbol.DataType) {
				asTypes = append(asTypes, struct {
					dataType *types.TTyping
					name     string
				}{
					dataType: symbol.DataType,
					name:     nameNode.Str0,
				})
			} else {
				asVars = append(asVars, struct {
					dataType *types.TTyping
					name     string
				}{
					dataType: symbol.DataType,
					name:     nameNode.Str0,
				})
			}
		}

		if len(asTypes) > 0 {
			for _, asType := range asTypes {
				info := analyzer.scope.Env.GetSymbol(asType.name)
				analyzer.write("type", false)
				analyzer.srcSp()
				analyzer.write(info.NameSpace, false)
				analyzer.srcSp()
				analyzer.write(fmt.Sprintf("%s.%s", pkg, asType.name), true)
			}
		}

		if len(asVars) > 0 {
			analyzer.write("var", false)
			analyzer.write("(", true)
			analyzer.incTb()
			for _, asVar := range asVars {
				info := analyzer.scope.Env.GetSymbol(asVar.name)
				analyzer.srcTb()
				if types.IsArr(info.DataType) {
					elementType := asVar.dataType.GetInternal0()
					analyzer.write(fmt.Sprintf(
						"%s %s = %s(%s.%s)",
						info.NameSpace,
						asVar.dataType.ToGoType(),
						GetArrayConstructor(elementType),
						pkg, asVar.name,
					), true)
				} else if types.IsMap(info.DataType) {
					keyType := info.DataType.GetInternal0()
					valueType := info.DataType.GetInternal1()
					analyzer.write(fmt.Sprintf(
						"%s %s = %s(%s.%s)",
						info.NameSpace,
						asVar.dataType.ToGoType(),
						GetMapConstructor(keyType, valueType),
						pkg,
						asVar.name,
					), true)
				} else {
					analyzer.write(fmt.Sprintf(
						"%s %s = %s.%s",
						info.NameSpace,
						asVar.dataType.ToGoType(),
						pkg,
						asVar.name,
					), true)
				}
			}
			analyzer.decTb()
			analyzer.write(")", false)
		}
		return
	}

	// Ensure import path is relative
	if !(strings.HasPrefix(pathNode.Str0, "./") || strings.HasPrefix(pathNode.Str0, "../")) {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"invalid import path, import path must be relative",
			pathNode.Position,
		)
	}

	// Verify imported file exists
	actualPath := ResolvePath(GetDir(analyzer.file.Path), pathNode.Str0)
	if !analyzer.state.HasFile(actualPath) {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"imported file not found",
			pathNode.Position,
		)
	}

	// Process each import name
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
		analyzer.writePosition(nameNode.Position)
		analyzer.srcTb()
		// For "var" (global), var fileName_a int = 100;
		// For "const" (global), const fileName_a int = 100;
		// For "const" (local), var a int = 100;
		// For "local" (local), var a int = 100;
		variableName := JoinVariableName(GetFileNameWithoutExtension(analyzer.file.Path), nameNode.Str0)
		goType := dataType.ToGoType()
		analyzer.write(fmt.Sprintf("%s %s", variableName, goType), false)
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
				NameSpace:    variableName,
				DataType:     dataType,
				Position:     nameNode.Position,
				IsGlobal:     analyzer.scope.InGlobal(),
				IsConst:      false,
				IsUsed:       false,
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
		analyzer.writePosition(nameNode.Position)
		analyzer.srcTb()
		// For "var" (global), var fileName_a int = 100;
		// For "const" (global), const fileName_a int = 100;
		// For "const" (local), var a int = 100;
		// For "local" (local), var a int = 100;
		variableName := nameNode.Str0
		if analyzer.scope.InGlobal() {
			variableName = JoinVariableName(GetFileNameWithoutExtension(analyzer.file.Path), nameNode.Str0)
		}
		analyzer.write(fmt.Sprintf("%s %s", variableName, dataType.ToGoType()), false)

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
				NameSpace:    variableName,
				Module:       "",
				DataType:     dataType,
				Position:     nameNode.Position,
				IsGlobal:     analyzer.scope.InGlobal(),
				IsConst:      true,
				IsUsed:       false,
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
	// Check if we're in a valid scope for local variables
	if !analyzer.scope.InLocal() || analyzer.scope.InSingle() {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"local variable is not allowed here",
			node.Position,
		)
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

		// Validate variable name
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

		// Validate variable type
		if types.IsVoid(dataType) {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				"invalid variable type, variable type cannot be void",
				nameNode.Position,
			)
		}

		analyzer.writePosition(nameNode.Position)
		analyzer.srcTb()

		// For "var" (global), var fileName_a int = 100;
		// For "const" (global), const fileName_a int = 100;
		// For "const" (local), var a int = 100;
		// For "local" (local), var a int = 100;
		analyzer.write(fmt.Sprintf("%s %s", nameNode.Str0, dataType.ToGoType()), false)

		// Handle variable initialization
		if valuNode != nil {
			analyzer.write(" = ", false)
			analyzer.expression(valuNode)
			valueType := analyzer.stack.Pop().DataType

			// Type compatibility check
			if !types.CanStore(dataType, valueType) {
				RaiseLanguageCompileError(
					analyzer.file.Path,
					analyzer.file.Data,
					fmt.Sprintf("cannot assign %s to %s", valueType.ToString(), dataType.ToString()),
					nameNode.Position,
				)
			}
		} else {
			// Use default value when no initializer is provided
			analyzer.write(" = ", false)
			analyzer.write(dataType.DefaultValue(), false)
		}

		// Add newline between variable declarations
		if index < len(namesNode)-1 {
			analyzer.srcNl()
		}

		// Register the symbol in the environment
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
				Module:       "",
				DataType:     dataType,
				Position:     nameNode.Position,
				IsGlobal:     false,
				IsConst:      false,
				IsUsed:       false,
				IsInitialize: valuNode != nil,
			})
		}
	}

	analyzer.decTb()
	analyzer.srcNl()
	analyzer.srcTb()
	analyzer.write(")", false)
}

func (analyzer *TAnalyzer) visitIf(node *TAst) {
	conditionNode := node.Ast0
	thenNode := node.Ast1
	elseNode := node.Ast2
	analyzer.write("if", false)
	analyzer.srcSp()
	analyzer.expression(conditionNode)
	analyzer.stack.Pop()
	analyzer.srcSp()
	if thenNode.Ttype == AstCodeBlock {
		analyzer.statement(thenNode)
	} else {
		analyzer.write("{", true)
		analyzer.incTb()
		analyzer.statement(thenNode)
		analyzer.decTb()
		analyzer.srcNl()
		analyzer.srcTb()
		analyzer.write("}", false)
	}
	if elseNode != nil {
		analyzer.write("else", false)
		analyzer.srcSp()
		analyzer.write("{", true)
		analyzer.incTb()
		analyzer.statement(elseNode)
		analyzer.decTb()
		analyzer.srcNl()
		analyzer.srcTb()
		analyzer.write("}", false)
	}
}

func (analyzer *TAnalyzer) visitRunStmnt(node *TAst) {
	exprNode := node.Ast0
	if exprNode.Ttype != AstCall {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"invalid run statement, run statement must be a function call",
			node.Position,
		)
	}
	analyzer.write("go ", false)
	analyzer.expression(exprNode)
	analyzer.stack.Pop()
}

func (analyzer *TAnalyzer) visitReturn(node *TAst) {
	if !analyzer.scope.InFunction() {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"return statement is not allowed here",
			node.Position,
		)
	}
	// Capture the function
	currentScope := analyzer.scope
	for currentScope.Type != ScopeFunction {
		currentScope = currentScope.Parent
	}
	// Return statement
	exprNode := node.Ast0
	analyzer.srcTb()
	analyzer.write("return", false)
	if exprNode != nil {
		analyzer.srcSp()
		analyzer.expression(exprNode)
		currentScope.Return = analyzer.stack.Pop().DataType
	} else {
		currentScope.Return = types.TVoid()
	}
}

func (analyzer *TAnalyzer) program(node *TAst) {
	// Create a new environment for the built-in functions.
	builtInEnv := CreateEnv(nil)
	Load(builtInEnv)
	// Collect required modules
	defer func() {
		src := analyzer.src
		analyzer.src = ""
		analyzer.write("package main", true)
		// Collect required modules
		analyzer.addModule("\"fmt\"")
		if analyzer.file.IsMain {
			analyzer.addModule("\"os\"")
		}
		// Use a map to track unique modules for better performance
		moduleMap := make(map[string]struct{})
		for _, symbol := range builtInEnv.Symbols {
			if symbol.IsUsed && len(symbol.Module) > 0 {
				if _, exists := moduleMap[symbol.Module]; !exists {
					moduleMap[symbol.Module] = struct{}{}
					analyzer.addModule(fmt.Sprintf("\"%s\"", symbol.Module))
				}
			}
		}
		if len(analyzer.modules) > 0 {
			analyzer.write(fmt.Sprintf("import (%s)", strings.Join(analyzer.modules, "\n")), true)
		}
		analyzer.src = analyzer.src + src
	}()

	// Set the parent of the file's environment to the built-in environment.
	analyzer.file.Env.Parent = builtInEnv

	// Create a new scope for the global variables.
	globalScope := &TScope{
		Parent: nil,
		Env:    analyzer.file.Env,
		Type:   ScopeGlobal,
		Return: nil,
	}

	// Set the scope to the global scope.
	analyzer.scope = globalScope

	// Analyze the program.
	lastIdx := len(node.AstArr0) - 1
	for index, child := range node.AstArr0 {
		analyzer.statement(child)
		if index < lastIdx {
			analyzer.srcNl()
		}
	}

	// Check if there are any unused variables.
	env := globalScope.Env
	for _, symbol := range env.Symbols {
		if !symbol.IsUsed && !symbol.IsGlobal {
			RaiseLanguageCompileError(
				analyzer.file.Path,
				analyzer.file.Data,
				fmt.Sprintf("unused variable: %s", symbol.Name),
				symbol.Position,
			)
		}
	}

	if !analyzer.file.IsMain {
		return
	}

	// Main file validation
	if !analyzer.file.Env.HasLocalSymbol("main") {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"main function is not defined",
			node.Position,
		)
	}

	mainFunc := analyzer.file.Env.GetSymbol("main")
	if !types.IsFunc(mainFunc.DataType) {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"main function is not a function",
			node.Position,
		)
	}

	members := mainFunc.DataType.GetMembers()
	if len(members) != 1 {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"main function must have exactly one parameter",
			node.Position,
		)
	}

	paramType := members[0].DataType
	if !types.IsArr(paramType) {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"main function parameter must be an array",
			node.Position,
		)
	}

	if !types.IsStr(paramType.GetInternal0()) {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			"main function parameter must be an array of strings",
			node.Position,
		)
	}

	analyzer.srcNl()
	analyzer.srcNl()
	analyzer.write("func main() {", true)
	analyzer.incTb()
	analyzer.srcTb()
	analyzer.write(
		fmt.Sprintf("%s(%s(os.Args))", mainFunc.NameSpace, GetArrayConstructor(paramType.GetInternal0())),
		false,
	)
	analyzer.decTb()
	analyzer.srcNl()
	analyzer.write("}", false)
}

// API:Export
func (analyzer *TAnalyzer) Analyze() string {
	analyzer.program(analyzer.file.Ast)
	if analyzer.stack.Size() != 0 {
		repr := ""
		for _, value := range analyzer.stack.stack {
			repr = repr + value.DataType.ToString() + ", "
		}
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			fmt.Sprintf("evaluation stack is not empty (%s)", repr),
			analyzer.file.Ast.Position,
		)
	}
	return analyzer.src
}
