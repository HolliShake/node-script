package main

type AstType int

const (
	AstIDN             AstType = iota
	AstInt             AstType = iota
	AstNum             AstType = iota
	AstStr             AstType = iota
	AstBool            AstType = iota
	AstNull            AstType = iota
	AstArray           AstType = iota
	AstHashMap         AstType = iota
	AstMember          AstType = iota
	AstNullSafeMember  AstType = iota
	AstIndex           AstType = iota
	AstCall            AstType = iota
	AstPlus            AstType = iota
	AstMinus           AstType = iota
	AstNot             AstType = iota
	AstBitNot          AstType = iota
	AstPlus2           AstType = iota
	AstMinus2          AstType = iota
	AstAllocation      AstType = iota
	AstMul             AstType = iota
	AstDiv             AstType = iota
	AstMod             AstType = iota
	AstAdd             AstType = iota
	AstSub             AstType = iota
	AstShl             AstType = iota
	AstShr             AstType = iota
	AstLt              AstType = iota
	AstLe              AstType = iota
	AstGt              AstType = iota
	AstGe              AstType = iota
	AstEq              AstType = iota
	AstNe              AstType = iota
	AstAnd             AstType = iota
	AstOr              AstType = iota
	AstXor             AstType = iota
	AstLogAnd          AstType = iota
	AstLogOr           AstType = iota
	AstAssign          AstType = iota
	AstBindAssign      AstType = iota
	AstMulAssign       AstType = iota
	AstDivAssign       AstType = iota
	AstModAssign       AstType = iota
	AstAddAssign       AstType = iota
	AstSubAssign       AstType = iota
	AstShlAssign       AstType = iota
	AstShrAssign       AstType = iota
	AstAndAssign       AstType = iota
	AstOrAssign        AstType = iota
	AstXorAssign       AstType = iota
	AstTupleExpression AstType = iota
	AstTypePointer     AstType = iota
	AstTypeInt8        AstType = iota // Typing
	AstTypeInt16       AstType = iota // Typing
	AstTypeInt32       AstType = iota // Typing
	AstTypeInt64       AstType = iota // Typing
	AstTypeNum         AstType = iota // Typing
	AstTypeStr         AstType = iota // Typing
	AstTypeBool        AstType = iota // Typing
	AstTypeVoid        AstType = iota // Typing
	AstTypeError       AstType = iota // Typing
	AstTypeFunc        AstType = iota // Typing
	AstTypeTuple       AstType = iota // Typing
	AstTypeHashMap     AstType = iota // Typing
	AstTypeArray       AstType = iota // Typing
	AstStruct          AstType = iota
	AstMethod          AstType = iota
	AstFunction        AstType = iota
	AstDo              AstType = iota
	AstWhile           AstType = iota
	AstImport          AstType = iota
	AstVar             AstType = iota
	AstLocal           AstType = iota
	AstConst           AstType = iota
	AstFor             AstType = iota
	AstForIf           AstType = iota
	AstIf              AstType = iota
	AstRunStmnt        AstType = iota
	AstContinueStmnt   AstType = iota
	AstBreakStmnt      AstType = iota
	AstReturnStmnt     AstType = iota
	AstCodeBlock       AstType = iota
	AstEmptyStmnt      AstType = iota
	AstExpressionStmnt AstType = iota
	AstCodeProgram     AstType = iota
)

type TAst struct {
	Ttype    AstType
	Position TPosition
	Str0     string
	Flg0     bool
	Ast0     *TAst
	Ast1     *TAst
	Ast2     *TAst
	Ast3     *TAst
	AstArr0  []*TAst
	AstArr1  []*TAst
	AstArr2  []*TAst
}

func CreateAst(ttype AstType, position TPosition) *TAst {
	ast := new(TAst)
	ast.Ttype = ttype
	ast.Position = position
	return ast
}

func AstTerminal(ttype AstType, position TPosition, str0 string) *TAst {
	ast := CreateAst(ttype, position)
	ast.Str0 = str0
	return ast
}

func AstSingle(ttype AstType, position TPosition, Ast0 *TAst) *TAst {
	ast := CreateAst(ttype, position)
	ast.Ast0 = Ast0
	return ast
}

func AstDouble(ttype AstType, position TPosition, Ast0 *TAst, Ast1 *TAst) *TAst {
	ast := CreateAst(ttype, position)
	ast.Ast0 = Ast0
	ast.Ast1 = Ast1
	return ast
}

func AstTriple(ttype AstType, position TPosition, Ast0 *TAst, Ast1 *TAst, Ast2 *TAst) *TAst {
	ast := CreateAst(ttype, position)
	ast.Ast0 = Ast0
	ast.Ast1 = Ast1
	ast.Ast2 = Ast2
	return ast
}

func AstQuad(ttype AstType, position TPosition, Ast0 *TAst, Ast1 *TAst, Ast2 *TAst, Ast3 *TAst) *TAst {
	ast := CreateAst(ttype, position)
	ast.Ast0 = Ast0
	ast.Ast1 = Ast1
	ast.Ast2 = Ast2
	ast.Ast3 = Ast3
	return ast
}

func AstSingleArray(ttype AstType, position TPosition, AstArr0 []*TAst) *TAst {
	ast := CreateAst(ttype, position)
	ast.AstArr0 = AstArr0
	return ast
}

func AstDoubleArray(ttype AstType, position TPosition, AstArr0 []*TAst, AstArr1 []*TAst) *TAst {
	ast := CreateAst(ttype, position)
	ast.AstArr0 = AstArr0
	ast.AstArr1 = AstArr1
	return ast
}

func AstSingleWithArray(ttype AstType, position TPosition, Ast0 *TAst, AstArr0 []*TAst) *TAst {
	ast := CreateAst(ttype, position)
	ast.Ast0 = Ast0
	ast.AstArr0 = AstArr0
	return ast
}

func AstSingleWithDoubleArray(ttype AstType, position TPosition, Ast0 *TAst, AstArr0 []*TAst, AstArr1 []*TAst) *TAst {
	ast := CreateAst(ttype, position)
	ast.Ast0 = Ast0
	ast.AstArr0 = AstArr0
	ast.AstArr1 = AstArr1
	return ast
}

// Custom AST functions for specific constructs

func AstPostfix(ttype AstType, position TPosition, Ast0 *TAst, opt string) *TAst {
	ast := CreateAst(ttype, position)
	ast.Str0 = opt
	ast.Ast0 = Ast0
	return ast
}

func AstUnary(ttype AstType, position TPosition, Ast0 *TAst, opt string) *TAst {
	ast := CreateAst(ttype, position)
	ast.Str0 = opt
	ast.Ast0 = Ast0
	return ast
}

func AstBinary(ttype AstType, position TPosition, lhs *TAst, rhs *TAst, opt string) *TAst {
	ast := CreateAst(ttype, position)
	ast.Str0 = opt
	ast.Ast0 = lhs
	ast.Ast1 = rhs
	return ast
}

func AstStructDec(ttype AstType, position TPosition, name *TAst, fieldNames []*TAst, fieldTypes []*TAst) *TAst {
	ast := CreateAst(ttype, position)
	ast.Ast0 = name
	ast.AstArr0 = fieldNames
	ast.AstArr1 = fieldTypes
	return ast
}

func AstFunctionDec(ttype AstType, position TPosition, name *TAst, returnType *TAst, paramNames []*TAst, paramTypes []*TAst, children []*TAst, panics bool) *TAst {
	ast := CreateAst(ttype, position)
	ast.Flg0 = panics
	ast.Ast0 = name
	ast.Ast1 = returnType
	ast.AstArr0 = paramNames
	ast.AstArr1 = paramTypes
	ast.AstArr2 = children
	return ast
}

func AstVarDec(ttype AstType, position TPosition, names []*TAst, types []*TAst, values []*TAst) *TAst {
	ast := CreateAst(ttype, position)
	ast.AstArr0 = names
	ast.AstArr1 = types
	ast.AstArr2 = values
	return ast
}

func AstForDec(ttype AstType, position TPosition, init *TAst, cond *TAst, post *TAst, body *TAst) *TAst {
	ast := CreateAst(ttype, position)
	ast.Ast0 = init
	ast.Ast1 = cond
	ast.Ast2 = post
	ast.Ast3 = body
	return ast
}

func AstBlock(ttype AstType, position TPosition, children []*TAst) *TAst {
	ast := CreateAst(ttype, position)
	ast.AstArr0 = children
	return ast
}

func IsNoEffectValueNode(node *TAst) bool {
	switch node.Ttype {
	case AstIDN, AstInt, AstNum, AstStr, AstBool, AstNull:
		return true
	case AstIndex, AstMember, AstCall:
		return true
	case AstTupleExpression:
		for _, child := range node.AstArr0 {
			if !IsConstantValueNode(child) {
				return false
			}
		}
		return true
	case AstArray:
		for _, child := range node.AstArr0 {
			if !IsConstantValueNode(child) {
				return false
			}
		}
		return true
	case AstHashMap:
		for index, child := range node.AstArr0 {
			valueNode := node.AstArr1[index]
			if !IsConstantValueNode(child) || !IsConstantValueNode(valueNode) {
				return false
			}
		}
		return true
	case AstAdd, AstSub, AstMul, AstDiv, AstMod, AstShl, AstShr, AstLt, AstLe, AstGt, AstGe, AstEq, AstNe, AstAnd, AstOr, AstXor, AstLogAnd, AstLogOr:
		return IsConstantValueNode(node.Ast0) && IsConstantValueNode(node.Ast1)
	}
	return IsConstantValueNode(node)
}

func IsConstantValueNode(node *TAst) bool {
	switch node.Ttype {
	case AstInt, AstNum, AstStr, AstBool, AstNull:
		return true
	case AstAdd, AstSub, AstMul, AstDiv, AstMod, AstShl, AstShr, AstLt, AstLe, AstGt, AstGe, AstEq, AstNe, AstAnd, AstOr, AstXor, AstLogAnd, AstLogOr:
		return IsConstantValueNode(node.Ast0) && IsConstantValueNode(node.Ast1)
	}
	return false
}

func GetAstTypeByPostfixOp(opt string) AstType {
	switch opt {
	case "++":
		return AstPlus2
	case "--":
		return AstMinus2
	default:
		RaiseSystemError("not implemented!")
	}
	return -1
}

func GetAstTypeByUnaryOp(opt string) AstType {
	switch opt {
	case "+":
		return AstPlus
	case "-":
		return AstMinus
	case "!":
		return AstNot
	case "~":
		return AstBitNot
	default:
		RaiseSystemError("invalid or not implemented unary operator!")
	}
	return -1
}

func GetAstTypeByBinaryOp(opt string) AstType {
	switch opt {
	case "*":
		return AstMul
	case "/":
		return AstDiv
	case "%":
		return AstMod
	case "+":
		return AstAdd
	case "-":
		return AstSub
	case "<<":
		return AstShl
	case ">>":
		return AstShr
	case "<":
		return AstLt
	case "<=":
		return AstLe
	case ">":
		return AstGt
	case ">=":
		return AstGe
	case "==":
		return AstEq
	case "!=":
		return AstNe
	case "&":
		return AstAnd
	case "|":
		return AstOr
	case "^":
		return AstXor
	case "&&":
		return AstLogAnd
	case "||":
		return AstLogOr
	case "=":
		return AstAssign
	case ":=":
		return AstBindAssign
	case "*=":
		return AstMulAssign
	case "/=":
		return AstDivAssign
	case "%=":
		return AstModAssign
	case "+=":
		return AstAddAssign
	case "-=":
		return AstSubAssign
	case "<<=":
		return AstShlAssign
	case ">>=":
		return AstShrAssign
	case "&=":
		return AstAndAssign
	case "|=":
		return AstOrAssign
	case "^=":
		return AstXorAssign
	default:
		RaiseSystemError("invalid or not implemented binary operator!")
	}
	return -1
}
