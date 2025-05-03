package main

type TTokenType int

const (
	TokenIDN TTokenType = iota
	TokenKEY TTokenType = iota
	TokenINT TTokenType = iota
	TokenNum TTokenType = iota
	TokenSTR TTokenType = iota
	TokenSYM TTokenType = iota
	TokenEOF TTokenType = iota
)

type TToken struct {
	Type     TTokenType
	Value    string
	Position TPosition
}

func GetTokenTypeName(tokenType TTokenType) string {
	switch tokenType {
	case TokenIDN:
		return "identifier"
	case TokenKEY:
		return "keyword"
	case TokenINT:
		return "integer"
	case TokenNum:
		return "number"
	case TokenSTR:
		return "string"
	case TokenSYM:
		return "symbol"
	case TokenEOF:
		return "end of file"
	default:
		return "unknown"
	}
}
