package main

import (
	"fmt"
)

type TParser struct {
	Tokenizer *TTokenizer
	look      TToken
}

// API:Export
func CreateParser(file string, data string) *TParser {
	parser := new(TParser)
	parser.Tokenizer = CreateTokenizer(file, data)
	return parser
}

func (parser *TParser) matchV(value string) bool {
	return (parser.look.Type == TokenKEY ||
		parser.look.Type == TokenSYM) &&
		parser.look.Value == value
}

func (parser *TParser) matchT(ttype TTokenType) bool {
	return parser.look.Type == ttype
}

func (parser *TParser) acceptV(value string) {
	if parser.matchV(value) {
		parser.look = parser.Tokenizer.Next()
		return
	}
	// Panic | error
	RaiseLanguageCompileError(
		parser.Tokenizer.File,
		parser.Tokenizer.Data,
		fmt.Sprintf("expected %s, got %s", value, parser.look.Value),
		parser.look.Position,
	)
}

func (parser *TParser) acceptT(ttype TTokenType) {
	if parser.matchT(ttype) {
		parser.look = parser.Tokenizer.Next()
		return
	}
	// Panic | error
	RaiseLanguageCompileError(
		parser.Tokenizer.File,
		parser.Tokenizer.Data,
		fmt.Sprintf("expected %s, got %s", GetTokenTypeName(ttype), GetTokenTypeName(parser.look.Type)),
		parser.look.Position,
	)
}

func (parser *TParser) terminal() *TAst {
	if parser.matchT(TokenIDN) {
		node := AstTerminal(
			AstIDN,
			parser.look.Position,
			parser.look.Value,
		)
		parser.acceptT(TokenIDN)
		return node
	} else if parser.matchT(TokenINT) {
		node := AstTerminal(
			AstInt,
			parser.look.Position,
			parser.look.Value,
		)
		parser.acceptT(TokenINT)
		return node
	} else if parser.matchT(TokenNum) {
		node := AstTerminal(
			AstNum,
			parser.look.Position,
			parser.look.Value,
		)
		parser.acceptT(TokenNum)
		return node
	} else if parser.matchT(TokenSTR) {
		node := AstTerminal(
			AstStr,
			parser.look.Position,
			parser.look.Value,
		)
		parser.acceptT(TokenSTR)
		return node
	} else if parser.matchT(TokenKEY) && (parser.matchV(KeyTrue) || parser.matchV(KeyFalse)) {
		node := AstTerminal(
			AstBool,
			parser.look.Position,
			parser.look.Value,
		)
		parser.acceptT(TokenKEY)
		return node
	} else if parser.matchT(TokenKEY) && parser.matchV(KeyNull) {
		node := AstTerminal(
			AstNull,
			parser.look.Position,
			parser.look.Value,
		)
		parser.acceptT(TokenKEY)
		return node
	}
	return nil
}

func (parser *TParser) group() *TAst {
	if parser.matchV("{") {
		return parser.hashmap()
	} else if parser.matchV("[") {
		return parser.array()
	} else if parser.matchV("(") {
		parser.acceptV("(")
		node := parser.mandatoryExpression()
		parser.acceptV(")")
		return node
	}
	return parser.terminal()
}

func (parser *TParser) array() *TAst {
	start := parser.look.Position
	ended := start
	parser.acceptV("[")
	elements := make([]*TAst, 0)
	elementN := parser.expression()
	if elementN != nil {
		elements = append(elements, elementN)
		for parser.matchV(",") {
			parser.acceptV(",")
			elementN = parser.expression()
			if elementN == nil {
				RaiseLanguageCompileError(
					parser.Tokenizer.File,
					parser.Tokenizer.Data,
					"missing expression after comma",
					parser.look.Position,
				)
			}
			elements = append(elements, elementN)
		}
	}
	ended = parser.look.Position
	parser.acceptV("]")
	return AstSingleArray(
		AstArray,
		start.Merge(ended),
		elements,
	)
}

func (parser *TParser) hashmap() *TAst {
	start := parser.look.Position
	ended := start
	parser.acceptV("{")
	keys := make([]*TAst, 0)
	vals := make([]*TAst, 0)
	var keyN *TAst = nil
	var valN *TAst = nil
	keyN = parser.expression()
	if keyN != nil {
		parser.acceptV(":")
		valN = parser.mandatoryExpression()
		keys = append(keys, keyN)
		vals = append(vals, valN)
		for parser.matchV(",") {
			parser.acceptV(",")
			keyN = parser.expression()
			if keyN == nil {
				RaiseLanguageCompileError(
					parser.Tokenizer.File,
					parser.Tokenizer.Data,
					"missing key expression after comma",
					parser.look.Position,
				)
			}
			parser.acceptV(":")
			valN = parser.mandatoryExpression()
			keys = append(keys, keyN)
			vals = append(vals, valN)
		}
	}
	ended = parser.look.Position
	parser.acceptV("}")
	return AstDoubleArray(
		AstHashMap,
		start.Merge(ended),
		keys,
		vals,
	)
}

func (parser *TParser) memberOrCall() *TAst {
	node := parser.group()
	if node == nil {
		return nil
	}
	for parser.matchV(".") || parser.matchV("[") || parser.matchV("(") {
		if parser.matchV(".") {
			parser.acceptV(".")
			member := parser.terminal()
			if member == nil {
				RaiseLanguageCompileError(
					parser.Tokenizer.File,
					parser.Tokenizer.Data,
					"missing member or expression",
					parser.look.Position,
				)
			}
			node = AstDouble(
				AstMember,
				node.Position.Merge(member.Position),
				node,
				member,
			)
		} else if parser.matchV("[") {
			parser.acceptV("[")
			index := parser.mandatoryExpression()
			ended := parser.look.Position
			parser.acceptV("]")
			node = AstDouble(
				AstIndex,
				node.Position.Merge(ended),
				node,
				index,
			)
		} else if parser.matchV("(") {
			parser.acceptV("(")
			arguments := make([]*TAst, 0)
			argN := parser.expression()
			if argN != nil {
				arguments = append(arguments, argN)
				for parser.matchV(",") {
					parser.acceptV(",")
					argN = parser.expression()
					if argN == nil {
						RaiseLanguageCompileError(
							parser.Tokenizer.File,
							parser.Tokenizer.Data,
							"missing expression after comma",
							parser.look.Position,
						)
					}
					arguments = append(arguments, argN)
				}
			}
			ended := parser.look.Position
			parser.acceptV(")")
			node = AstSingleWithArray(
				AstCall,
				node.Position.Merge(ended),
				node,
				arguments,
			)
		}
	}
	return node
}

func (parser *TParser) postfix() *TAst {
	node := parser.memberOrCall()
	if node == nil {
		return nil
	}
	for parser.matchV("++") || parser.matchV("--") {
		opt := parser.look.Value
		parser.acceptT(TokenSYM)
		node = AstUnary(
			GetAstTypeByPostfixOp(opt),
			node.Position,
			node,
			opt,
		)
	}
	return node
}

func (parser *TParser) ifExpression() *TAst {
	if parser.matchV(KeyIf) {
		start := parser.look.Position
		ended := start
		parser.acceptV(KeyIf)
		parser.acceptV("(")
		condition := parser.mandatoryExpression()
		parser.acceptV(")")
		body := parser.expression()
		if body == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing if body",
				parser.look.Position,
			)
		}
		parser.acceptV(KeyElse)
		elseBody := parser.expression()
		if elseBody == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing else body",
				parser.look.Position,
			)
		}
		ended = elseBody.Position
		return AstDouble(
			AstIf,
			start.Merge(ended),
			condition,
			body,
		)
	}
	return parser.postfix()
}

func (parser *TParser) unary() *TAst {
	if parser.matchV("+") || parser.matchV("-") || parser.matchV("!") || parser.matchV("~") {
		opt := parser.look.Value
		parser.acceptT(TokenSYM)
		node := parser.unary()
		if node == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing right-hand expression",
				parser.look.Position,
			)
			return nil
		}
		return AstUnary(
			GetAstTypeByUnaryOp(opt),
			node.Position,
			node,
			opt,
		)
	} else if parser.matchV("++") || parser.matchV("--") {
		opt := parser.look.Value
		parser.acceptT(TokenSYM)
		node := parser.memberOrCall()
		if node == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing right-hand expression",
				parser.look.Position,
			)
			return nil
		}
		return AstUnary(
			GetAstTypeByUnaryOp(opt),
			node.Position,
			node,
			opt,
		)
	} else if parser.matchV(keyAwait) {
		opt := parser.look.Value
		parser.acceptT(TokenKEY)
		node := parser.memberOrCall()
		if node == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing right-hand expression",
				parser.look.Position,
			)
			return nil
		}
		return AstUnary(
			GetAstTypeByUnaryOp(opt),
			node.Position,
			node,
			opt,
		)
	}
	return parser.ifExpression()
}

func (parser *TParser) multiplicative() *TAst {
	lhs := parser.unary()
	if lhs == nil {
		return nil
	}
	for parser.matchV("*") || parser.matchV("/") || parser.matchV("%") {
		opt := parser.look.Value
		parser.acceptT(TokenSYM)
		rhs := parser.unary()
		if rhs == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing right-hand expression",
				lhs.Position,
			)
			return nil
		}
		lhs = AstBinary(
			GetAstTypeByBinaryOp(opt),
			lhs.Position.Merge(rhs.Position),
			lhs,
			rhs,
			opt,
		)
	}
	return lhs
}

func (parser *TParser) additive() *TAst {
	lhs := parser.multiplicative()
	if lhs == nil {
		return nil
	}
	for parser.matchV("+") || parser.matchV("-") {
		opt := parser.look.Value
		parser.acceptT(TokenSYM)
		rhs := parser.multiplicative()
		if rhs == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing right-hand expression",
				lhs.Position,
			)
			return nil
		}
		lhs = AstBinary(
			GetAstTypeByBinaryOp(opt),
			lhs.Position.Merge(rhs.Position),
			lhs,
			rhs,
			opt,
		)
	}
	return lhs
}

func (parser *TParser) shift() *TAst {
	lhs := parser.additive()
	if lhs == nil {
		return nil
	}
	for parser.matchV("<<") || parser.matchV(">>") {
		opt := parser.look.Value
		parser.acceptT(TokenSYM)
		rhs := parser.additive()
		if rhs == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing right-hand expression",
				lhs.Position,
			)
			return nil
		}
		lhs = AstBinary(
			GetAstTypeByBinaryOp(opt),
			lhs.Position.Merge(rhs.Position),
			lhs,
			rhs,
			opt,
		)
	}
	return lhs
}

func (parser *TParser) relational() *TAst {
	lhs := parser.shift()
	if lhs == nil {
		return nil
	}
	for parser.matchV("<") || parser.matchV("<=") ||
		parser.matchV(">") || parser.matchV(">=") {
		opt := parser.look.Value
		parser.acceptT(TokenSYM)
		rhs := parser.shift()
		if rhs == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing right-hand expression",
				lhs.Position,
			)
			return nil
		}
		lhs = AstBinary(
			GetAstTypeByBinaryOp(opt),
			lhs.Position.Merge(rhs.Position),
			lhs,
			rhs,
			opt,
		)
	}
	return lhs
}

func (parser *TParser) equality() *TAst {
	lhs := parser.relational()
	if lhs == nil {
		return nil
	}
	for parser.matchV("==") || parser.matchV("!=") {
		opt := parser.look.Value
		parser.acceptT(TokenSYM)
		rhs := parser.relational()
		if rhs == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing right-hand expression",
				lhs.Position,
			)
			return nil
		}
		lhs = AstBinary(
			GetAstTypeByBinaryOp(opt),
			lhs.Position.Merge(rhs.Position),
			lhs,
			rhs,
			opt,
		)
	}
	return lhs
}

func (parser *TParser) bitwise() *TAst {
	lhs := parser.equality()
	if lhs == nil {
		return nil
	}
	for parser.matchV("&") || parser.matchV("|") || parser.matchV("^") {
		opt := parser.look.Value
		parser.acceptT(TokenSYM)
		rhs := parser.equality()
		if rhs == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing right-hand expression",
				lhs.Position,
			)
			return nil
		}
		lhs = AstBinary(
			GetAstTypeByBinaryOp(opt),
			lhs.Position.Merge(rhs.Position),
			lhs,
			rhs,
			opt,
		)
	}
	return lhs
}

func (parser *TParser) logical() *TAst {
	lhs := parser.bitwise()
	if lhs == nil {
		return nil
	}
	for parser.matchV("&&") || parser.matchV("||") {
		opt := parser.look.Value
		parser.acceptT(TokenSYM)
		rhs := parser.bitwise()
		if rhs == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing right-hand expression",
				lhs.Position,
			)
			return nil
		}
		lhs = AstBinary(
			GetAstTypeByBinaryOp(opt),
			lhs.Position.Merge(rhs.Position),
			lhs,
			rhs,
			opt,
		)
	}
	return lhs
}

func (parser *TParser) simpleAssign() *TAst {
	lhs := parser.logical()
	if lhs == nil {
		return nil
	}
	for parser.matchV("=") {
		opt := parser.look.Value
		parser.acceptT(TokenSYM)
		rhs := parser.logical()
		if rhs == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing right-hand expression",
				lhs.Position,
			)
			return nil
		}
		lhs = AstBinary(
			GetAstTypeByBinaryOp(opt),
			lhs.Position.Merge(rhs.Position),
			lhs,
			rhs,
			opt,
		)
	}
	return lhs
}

func (parser *TParser) mulAssign() *TAst {
	lhs := parser.simpleAssign()
	if lhs == nil {
		return nil
	}
	for parser.matchV("*=") || parser.matchV("/=") || parser.matchV("%=") {
		opt := parser.look.Value
		parser.acceptT(TokenSYM)
		rhs := parser.simpleAssign()
		if rhs == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing right-hand expression",
				lhs.Position,
			)
			return nil
		}
		lhs = AstBinary(
			GetAstTypeByBinaryOp(opt),
			lhs.Position.Merge(rhs.Position),
			lhs,
			rhs,
			opt,
		)
	}
	return lhs
}

func (parser *TParser) addAssign() *TAst {
	lhs := parser.mulAssign()
	if lhs == nil {
		return nil
	}
	for parser.matchV("+=") || parser.matchV("-=") {
		opt := parser.look.Value
		parser.acceptT(TokenSYM)
		rhs := parser.mulAssign()
		if rhs == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing right-hand expression",
				lhs.Position,
			)
			return nil
		}
		lhs = AstBinary(
			GetAstTypeByBinaryOp(opt),
			lhs.Position.Merge(rhs.Position),
			lhs,
			rhs,
			opt,
		)
	}
	return lhs
}

func (parser *TParser) shiftAssign() *TAst {
	lhs := parser.addAssign()
	if lhs == nil {
		return nil
	}
	for parser.matchV("<<=") || parser.matchV(">>=") {
		opt := parser.look.Value
		parser.acceptT(TokenSYM)
		rhs := parser.addAssign()
		if rhs == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing right-hand expression",
				lhs.Position,
			)
			return nil
		}
		lhs = AstBinary(
			GetAstTypeByBinaryOp(opt),
			lhs.Position.Merge(rhs.Position),
			lhs,
			rhs,
			opt,
		)
	}
	return lhs
}

func (parser *TParser) bitAssign() *TAst {
	lhs := parser.shiftAssign()
	if lhs == nil {
		return nil
	}
	for parser.matchV("&=") || parser.matchV("|=") || parser.matchV("^=") {
		opt := parser.look.Value
		parser.acceptT(TokenSYM)
		rhs := parser.shiftAssign()
		if rhs == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing right-hand expression",
				lhs.Position,
			)
			return nil
		}
		lhs = AstBinary(
			GetAstTypeByBinaryOp(opt),
			lhs.Position.Merge(rhs.Position),
			lhs,
			rhs,
			opt,
		)
	}
	return lhs
}

func (parser *TParser) expression() *TAst {
	return parser.bitAssign()
}

func (parser *TParser) mandatoryExpression() *TAst {
	if node := parser.expression(); node != nil {
		return node
	}
	RaiseLanguageCompileError(
		parser.Tokenizer.File,
		parser.Tokenizer.Data,
		"missing expression",
		parser.look.Position,
	)
	return nil
}

func (parser *TParser) baseType() *TAst {
	if parser.matchV("{") {
		start := parser.look.Position
		ended := start
		parser.acceptV("{")
		keyType := parser.typeOrNil()
		if keyType == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing key type",
				parser.look.Position,
			)
		}
		parser.acceptV(":")
		valueType := parser.typeOrNil()
		if valueType == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing value type",
				parser.look.Position,
			)
		}
		ended = parser.look.Position
		parser.acceptV("}")
		return AstDouble(
			AstTypeHashMap,
			start.Merge(ended),
			keyType,
			valueType,
		)
	} else if parser.matchV("[") {
		start := parser.look.Position
		ended := start
		parser.acceptV("[")
		elementType := parser.typeOrNil()
		if elementType == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing element type",
				parser.look.Position,
			)
		}
		ended = parser.look.Position
		parser.acceptV("]")
		return AstSingle(
			AstTypeArray,
			start.Merge(ended),
			elementType,
		)
	} else if parser.matchV("(") {
		start := parser.look.Position
		parser.acceptV("(")
		argumentTypes := make([]*TAst, 0)
		argN := parser.typeOrNil()
		if argN != nil {
			argumentTypes = append(argumentTypes, argN)
			for parser.matchV(",") {
				parser.acceptV(",")
				argN = parser.typeOrNil()
				if argN == nil {
					RaiseLanguageCompileError(
						parser.Tokenizer.File,
						parser.Tokenizer.Data,
						"missing expression after comma",
						parser.look.Position,
					)
				}
				argumentTypes = append(argumentTypes, argN)
			}
		}
		parser.acceptV(")")
		returnNode := parser.typeOrNil()
		if returnNode == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing return type",
				parser.look.Position,
			)
		}
		return AstSingleWithArray(
			AstTypeFunc,
			start.Merge(returnNode.Position),
			returnNode,
			argumentTypes,
		)
	} else if parser.matchT(TokenKEY) && parser.matchV(KeyInt8) {
		node := AstTerminal(
			AstTypeInt8,
			parser.look.Position,
			parser.look.Value,
		)
		parser.acceptT(TokenKEY)
		return node
	} else if parser.matchT(TokenKEY) && parser.matchV(KeyInt16) {
		node := AstTerminal(
			AstTypeInt16,
			parser.look.Position,
			parser.look.Value,
		)
		parser.acceptT(TokenKEY)
		return node
	} else if parser.matchT(TokenKEY) && parser.matchV(KeyInt32) {
		node := AstTerminal(
			AstTypeInt32,
			parser.look.Position,
			parser.look.Value,
		)
		parser.acceptT(TokenKEY)
		return node
	} else if parser.matchT(TokenKEY) && parser.matchV(KeyInt64) {
		node := AstTerminal(
			AstTypeInt64,
			parser.look.Position,
			parser.look.Value,
		)
		parser.acceptT(TokenKEY)
		return node
	} else if parser.matchT(TokenKEY) && parser.matchV(KeyNum) {
		node := AstTerminal(
			AstTypeNum,
			parser.look.Position,
			parser.look.Value,
		)
		parser.acceptT(TokenKEY)
		return node
	} else if parser.matchT(TokenKEY) && parser.matchV(KeyStr) {
		node := AstTerminal(
			AstTypeStr,
			parser.look.Position,
			parser.look.Value,
		)
		parser.acceptT(TokenKEY)
		return node
	} else if parser.matchT(TokenKEY) && parser.matchV(KeyBool) {
		node := AstTerminal(
			AstTypeBool,
			parser.look.Position,
			parser.look.Value,
		)
		parser.acceptT(TokenKEY)
		return node
	} else if parser.matchT(TokenKEY) && parser.matchV(KeyVoid) {
		node := AstTerminal(
			AstTypeVoid,
			parser.look.Position,
			parser.look.Value,
		)
		parser.acceptT(TokenKEY)
		return node
	}
	return parser.terminal()
}

func (parser *TParser) typeOrNil() *TAst {
	dtypeAst := parser.baseType()
	if dtypeAst == nil {
		return nil
	}
	return dtypeAst
}

func (parser *TParser) typing() *TAst {
	dtypeAst := parser.typeOrNil()
	if dtypeAst == nil {
		RaiseLanguageCompileError(
			parser.Tokenizer.File,
			parser.Tokenizer.Data,
			"missing type",
			parser.look.Position,
		)
	}
	return dtypeAst
}

func (parser *TParser) statement() *TAst {
	if parser.matchV(KeyStruct) {
		return parser.structDecl()
	} else if parser.matchV(KeyFunc) {
		return parser.funcDecl()
	} else if parser.matchV(KeyImport) {
		return parser.importDecl()
	} else if parser.matchV(KeyVar) {
		return parser.varDecl()
	} else if parser.matchV(KeyConst) {
		return parser.constDecl()
	} else if parser.matchV(KeyLocal) {
		return parser.localDecl()
	} else if parser.matchV(KeyFor) {
		return parser.forDecl()
	} else if parser.matchV(KeyDo) {
		return parser.doWhileDecl()
	} else if parser.matchV(KeyWhile) {
		return parser.whileDecl()
	} else if parser.matchV(KeyIf) {
		return parser.ifDecl()
	} else if parser.matchV(KeyContinue) {
		return parser.continueStmnt()
	} else if parser.matchV(KeyBreak) {
		return parser.breakStmnt()
	} else if parser.matchV(KeyReturn) {
		return parser.returnStmnt()
	} else if parser.matchV("{") {
		return parser.blockStmnt()
	}
	return parser.expression()
}

func (parser *TParser) structDecl() *TAst {
	start := parser.look.Position
	ended := start
	parser.acceptV(KeyStruct)
	nameAst := parser.terminal()
	if nameAst == nil {
		RaiseLanguageCompileError(
			parser.Tokenizer.File,
			parser.Tokenizer.Data,
			"missing struct name",
			parser.look.Position,
		)
	}
	parser.acceptV("{")
	names := make([]*TAst, 0)
	types := make([]*TAst, 0)
	nameN := parser.terminal()
	if nameN == nil {
		RaiseLanguageCompileError(
			parser.Tokenizer.File,
			parser.Tokenizer.Data,
			"struct must have at least one field",
			parser.look.Position,
		)
	}
	typeN := parser.typing()
	for nameN != nil {
		names = append(names, nameN)
		types = append(types, typeN)
		parser.acceptV(";")
		nameN = parser.terminal()
		if nameN == nil {
			continue
		}
		typeN = parser.typing()
	}
	ended = parser.look.Position
	parser.acceptV("}")
	return AstStructDec(
		AstStruct,
		start.Merge(ended),
		nameAst,
		names,
		types,
	)
}

func (parser *TParser) funcDecl() *TAst {
	start := parser.look.Position
	ended := start
	parser.acceptV(KeyFunc)
	var thisName *TAst
	var thisType *TAst
	isMethod := parser.matchV("(")
	if isMethod {
		parser.acceptV("(")
		thisName = parser.terminal()
		if thisName == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing method name",
				parser.look.Position,
			)
		}
		thisType = parser.typing()
		parser.acceptV(")")
	}
	funcNameAst := parser.terminal()
	if funcNameAst == nil {
		RaiseLanguageCompileError(
			parser.Tokenizer.File,
			parser.Tokenizer.Data,
			"missing function name",
			parser.look.Position,
		)
	}
	parser.acceptV("(")
	names := make([]*TAst, 0)
	types := make([]*TAst, 0)
	// Append thisName and thisType to names and types if isMethod is true
	if isMethod {
		names = append(names, thisName)
		types = append(types, thisType)
	}
	var nameN *TAst = parser.terminal()
	var typeN *TAst
	if nameN != nil {
		typeN = parser.typing()
		names = append(names, nameN)
		types = append(types, typeN)
		for parser.matchV(",") {
			parser.acceptV(",")
			nameN := parser.terminal()
			if nameN == nil {
				RaiseLanguageCompileError(
					parser.Tokenizer.File,
					parser.Tokenizer.Data,
					"missing field name",
					parser.look.Position,
				)
			}
			typeN := parser.typing()
			names = append(names, nameN)
			types = append(types, typeN)
		}
	}
	parser.acceptV(")")
	returnTypeAst := parser.typing()
	parser.acceptV("{")
	children := make([]*TAst, 0)
	childN := parser.statement()
	for childN != nil {
		children = append(children, childN)
		childN = parser.statement()
	}
	ended = parser.look.Position
	parser.acceptV("}")
	funcType := AstFunc
	if isMethod {
		funcType = AstMethod
	}
	return AstFuncDec(
		funcType,
		start.Merge(ended),
		nameN,
		returnTypeAst,
		names,
		types,
		children,
	)
}

func (parser *TParser) importDecl() *TAst {
	start := parser.look.Position
	ended := start
	parser.acceptV(KeyImport)
	parser.acceptV("(")
	names := make([]*TAst, 0)
	nameN := parser.terminal()
	if nameN == nil {
		RaiseLanguageCompileError(
			parser.Tokenizer.File,
			parser.Tokenizer.Data,
			"missing import name",
			parser.look.Position,
		)
	}
	names = append(names, nameN)
	for parser.matchV(",") {
		parser.acceptV(",")
		nameN = parser.terminal()
		if nameN == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing import name after comma",
				parser.look.Position,
			)
		}
		names = append(names, nameN)
	}
	parser.acceptV(")")
	parser.acceptV(KeyFrom)
	path := parser.terminal()
	if path == nil {
		RaiseLanguageCompileError(
			parser.Tokenizer.File,
			parser.Tokenizer.Data,
			"missing import path",
			parser.look.Position,
		)
	}
	ended = parser.look.Position
	parser.acceptV(";")
	return AstSingleWithArray(
		AstImport,
		start.Merge(ended),
		path,
		names,
	)
}

func (parser *TParser) varDecl() *TAst {
	start := parser.look.Position
	ended := start
	parser.acceptV(KeyVar)
	names := make([]*TAst, 0)
	types := make([]*TAst, 0)
	valus := make([]*TAst, 0)
	var nameN *TAst = parser.terminal()
	var typeN *TAst = nil
	var value *TAst = nil
	if nameN == nil {
		RaiseLanguageCompileError(
			parser.Tokenizer.File,
			parser.Tokenizer.Data,
			"missing field name",
			parser.look.Position,
		)
	}
	typeN = parser.typing()
	if parser.matchV("=") {
		parser.acceptV("=")
		value = parser.mandatoryExpression()
	}
	names = append(names, nameN)
	types = append(types, typeN)
	valus = append(valus, value)
	for parser.matchV(",") {
		nameN = parser.terminal()
		if nameN == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing field name after comma",
				parser.look.Position,
			)
		}
		typeN = parser.typing()
		if parser.matchV("=") {
			parser.acceptV("=")
			value = parser.mandatoryExpression()
		}
		names = append(names, nameN)
		types = append(types, typeN)
		valus = append(valus, value)
	}
	parser.acceptV(";")
	return AstVarDec(
		AstVar,
		start.Merge(ended),
		names,
		types,
		valus,
	)
}

func (parser *TParser) constDecl() *TAst {
	start := parser.look.Position
	ended := start
	parser.acceptV(KeyConst)
	names := make([]*TAst, 0)
	types := make([]*TAst, 0)
	valus := make([]*TAst, 0)
	var nameN *TAst = parser.terminal()
	var typeN *TAst = nil
	var value *TAst = nil
	if nameN == nil {
		RaiseLanguageCompileError(
			parser.Tokenizer.File,
			parser.Tokenizer.Data,
			"missing field name",
			parser.look.Position,
		)
	}
	typeN = parser.typing()
	if parser.matchV("=") {
		parser.acceptV("=")
		value = parser.mandatoryExpression()
	}
	names = append(names, nameN)
	types = append(types, typeN)
	valus = append(valus, value)
	for parser.matchV(",") {
		nameN = parser.terminal()
		if nameN == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing field name after comma",
				parser.look.Position,
			)
		}
		typeN = parser.typing()
		if parser.matchV("=") {
			parser.acceptV("=")
			value = parser.mandatoryExpression()
		}
		names = append(names, nameN)
		types = append(types, typeN)
		valus = append(valus, value)
	}
	parser.acceptV(";")
	return AstVarDec(
		AstVar,
		start.Merge(ended),
		names,
		types,
		valus,
	)
}

func (parser *TParser) localDecl() *TAst {
	start := parser.look.Position
	ended := start
	parser.acceptV(KeyLocal)
	names := make([]*TAst, 0)
	types := make([]*TAst, 0)
	valus := make([]*TAst, 0)
	var nameN *TAst = parser.terminal()
	var typeN *TAst = nil
	var value *TAst = nil
	if nameN == nil {
		RaiseLanguageCompileError(
			parser.Tokenizer.File,
			parser.Tokenizer.Data,
			"missing field name",
			parser.look.Position,
		)
	}
	typeN = parser.typing()
	if parser.matchV("=") {
		parser.acceptV("=")
		value = parser.mandatoryExpression()
	}
	names = append(names, nameN)
	types = append(types, typeN)
	valus = append(valus, value)
	for parser.matchV(",") {
		nameN = parser.terminal()
		if nameN == nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"missing field name after comma",
				parser.look.Position,
			)
		}
		typeN = parser.typing()
		if parser.matchV("=") {
			parser.acceptV("=")
			value = parser.mandatoryExpression()
		}
		names = append(names, nameN)
		types = append(types, typeN)
		valus = append(valus, value)
	}
	parser.acceptV(";")
	return AstVarDec(
		AstLocal,
		start.Merge(ended),
		names,
		types,
		valus,
	)
}

func (parser *TParser) forModeDecl() *TAst {
	if parser.matchV(KeyVar) || parser.matchV(KeyConst) || parser.matchV(KeyLocal) {
		return parser.statement()
	} else {
		return parser.expression()
	}
}

func (parser *TParser) forDecl() *TAst {
	start := parser.look.Position
	ended := start
	parser.acceptV(KeyFor)
	var init *TAst = nil
	var cond *TAst = nil
	var mutt *TAst = nil
	isForMode := parser.matchV("(")
	isForIf := false
	if isForMode {
		parser.acceptV("(")
		init = parser.forModeDecl()
		// If nil, accept a semicolon
		if init == nil {
			parser.acceptV(";")
		}
		cond = parser.expression()
		parser.acceptV(";")
		mutt = parser.expression()
		parser.acceptV(")")
	}
	body := parser.statement()
	if body == nil {
		RaiseLanguageCompileError(
			parser.Tokenizer.File,
			parser.Tokenizer.Data,
			"missing body in for loop",
			parser.look.Position,
		)
	}
	if parser.matchV(KeyIf) {
		if isForMode || cond != nil || mutt != nil {
			RaiseLanguageCompileError(
				parser.Tokenizer.File,
				parser.Tokenizer.Data,
				"ambiguous 'if' in for loop",
				parser.look.Position,
			)
		}
		isForIf = true
		parser.acceptV(KeyIf)
		parser.acceptV("(")
		cond = parser.mandatoryExpression()
		parser.acceptV(")")
		parser.acceptV(";")
	}
	forType := AstFor
	if isForMode {
		forType = AstFor
	} else if isForIf {
		forType = AstForIf
	} else {
		RaiseLanguageCompileError(
			parser.Tokenizer.File,
			parser.Tokenizer.Data,
			"invalid 'for' statement",
			parser.look.Position,
		)
	}
	return AstForDec(
		forType,
		start.Merge(ended),
		init,
		cond,
		mutt,
		body,
	)
}

func (parser *TParser) doWhileDecl() *TAst {
	start := parser.look.Position
	ended := start
	parser.acceptV(KeyDo)
	body := parser.statement()
	if body == nil {
		RaiseLanguageCompileError(
			parser.Tokenizer.File,
			parser.Tokenizer.Data,
			"missing body in do while loop",
			parser.look.Position,
		)
	}
	parser.acceptV(KeyWhile)
	parser.acceptV("(")
	condition := parser.mandatoryExpression()
	parser.acceptV(")")
	ended = parser.look.Position
	parser.acceptV(";")
	return AstDouble(
		AstDo,
		start.Merge(ended),
		condition,
		body,
	)
}

func (parser *TParser) whileDecl() *TAst {
	start := parser.look.Position
	ended := start
	parser.acceptV(KeyWhile)
	parser.acceptV("(")
	condition := parser.mandatoryExpression()
	parser.acceptV(")")
	body := parser.statement()
	if body == nil {
		RaiseLanguageCompileError(
			parser.Tokenizer.File,
			parser.Tokenizer.Data,
			"missing body in while loop",
			parser.look.Position,
		)
	}
	return AstDouble(
		AstWhile,
		start.Merge(ended),
		condition,
		body,
	)
}

func (parser *TParser) ifDecl() *TAst {
	start := parser.look.Position
	ended := start
	parser.acceptV(KeyIf)
	parser.acceptV("(")
	cond := parser.mandatoryExpression()
	parser.acceptV(")")
	body := parser.statement()
	if body == nil {
		RaiseLanguageCompileError(
			parser.Tokenizer.File,
			parser.Tokenizer.Data,
			"invalid 'if' statement",
			parser.look.Position,
		)
	}
	var elseValue *TAst = nil
	if parser.matchV(KeyElse) {
		parser.acceptV(KeyElse)
		elseValue = parser.statement()
	}
	return AstTriple(
		AstIf,
		start.Merge(ended),
		cond,
		body,
		elseValue,
	)
}

func (parser *TParser) continueStmnt() *TAst {
	start := parser.look.Position
	ended := start
	parser.acceptV(KeyContinue)
	ended = parser.look.Position
	parser.acceptV(";")
	return CreateAst(AstContinueStmnt, start.Merge(ended))
}

func (parser *TParser) breakStmnt() *TAst {
	start := parser.look.Position
	ended := start
	parser.acceptV(KeyBreak)
	ended = parser.look.Position
	parser.acceptV(";")
	return CreateAst(AstBreakStmnt, start.Merge(ended))
}

func (parser *TParser) returnStmnt() *TAst {
	start := parser.look.Position
	ended := start
	parser.acceptV(KeyReturn)
	expr := parser.expression()
	ended = parser.look.Position
	parser.acceptV(";")
	return AstSingle(
		AstReturnStmnt,
		start.Merge(ended),
		expr,
	)
}

func (parser *TParser) blockStmnt() *TAst {
	start := parser.look.Position
	ended := start
	parser.acceptV("{")
	childrenN := make([]*TAst, 0)
	for childN := parser.statement(); childN != nil; {
		childrenN = append(childrenN, childN)
		childN = parser.statement()
	}
	ended = parser.look.Position
	parser.acceptV("}")
	return AstBlock(
		AstCodeBlock,
		start.Merge(ended),
		childrenN,
	)
}

func (parser *TParser) program() *TAst {
	start := parser.look.Position
	ended := start
	children := make([]*TAst, 0)
	for child := parser.statement(); child != nil; {
		children = append(children, child)
		child = parser.statement()
	}
	ended = parser.look.Position
	parser.acceptT(TokenEOF)
	return AstBlock(
		AstCodeProgram,
		start.Merge(ended),
		children,
	)
}

// API:Export
func (parser *TParser) Parse() *TAst {
	parser.look = parser.Tokenizer.Next()
	return parser.program()
}
