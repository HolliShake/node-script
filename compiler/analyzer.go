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
		if analyzer.file.Env.HasGlobalSymbol(node.Str0) {
			symbol := analyzer.file.Env.GetSymbol(node.Str0)
			analyzer.file.Env.UpdateSymbolIsUsed(node.Str0, true)
			return symbol.DataType
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
				INVALID_ARRAY_ELEMENT_TYPE,
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
		analyzer.write("append", false)
		analyzer.write("(", false)
		analyzer.write("make", false)
		analyzer.write("(", false)
		analyzer.write("[]", false)
		analyzer.write(elementType.ToGoType(), false)
		analyzer.write(", ", false)
		analyzer.write("0", false)
		analyzer.write(")", false)
		analyzer.write(", ", false)
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
		analyzer.write(")", false)
		analyzer.stack.Push(CreateValue(
			types.TArray(elementType),
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
	case AstCall:
		objectNode := node.Ast0
		parametersNode := node.AstArr0
		analyzer.expression(objectNode)
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
		// TODO: Check parameter type is valid.
		requiredParameters := objectValue.DataType.GetMembers()
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
					analyzer.stack.Pop()
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
		if (node.Ast0.Ttype == AstIDN || node.Ast0.Ttype == AstCall || IsConstantValueNode(node.Ast0)) && !types.IsVoid(value.DataType) {
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
		analyzer.write(" ", false)
		analyzer.write(returnType.DefaultValue(), false)
	} else if !types.CanStore(returnType, functionScope.Return) {
		RaiseLanguageCompileError(
			analyzer.file.Path,
			analyzer.file.Data,
			fmt.Sprintf("cannot return %s, return type must be %s", functionScope.Return.ToGoType(), returnType.ToGoType()),
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
				"method already exists",
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
		modules := make([]string, 0, 8) // Pre-allocate capacity
		if analyzer.file.IsMain {
			modules = append(modules, "\"os\"")
		}
		// Use a map to track unique modules for better performance
		moduleMap := make(map[string]struct{})
		for _, symbol := range builtInEnv.Symbols {
			if symbol.IsUsed && len(symbol.Module) > 0 {
				moduleStr := fmt.Sprintf("\"%s\"", symbol.Module)
				if _, exists := moduleMap[moduleStr]; !exists {
					moduleMap[moduleStr] = struct{}{}
					modules = append(modules, moduleStr)
				}
			}
		}
		if len(modules) > 0 {
			analyzer.write(fmt.Sprintf("import (%s)", strings.Join(modules, "\n")), true)
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
		fmt.Sprintf("%s(append(make([]string, 0, len(os.Args)), os.Args...))", mainFunc.NameSpace),
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
