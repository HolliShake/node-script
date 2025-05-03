package main

import (
	"fmt"
	"unicode"
)

// API:Export
type TTokenizer struct {
	File string
	Data []rune
	size int
	look rune
	indx int
	line int
	colm int
}

// API:Export
func CreateTokenizer(file string, data string) *TTokenizer {
	tokenizer := new(TTokenizer)
	tokenizer.File = file
	tokenizer.Data = []rune(data)
	tokenizer.size = len(tokenizer.Data)
	if len(tokenizer.Data) > 0 {
		tokenizer.look = tokenizer.Data[0]
	} else {
		tokenizer.look = -1
	}
	tokenizer.indx = 0
	tokenizer.line = 1
	tokenizer.colm = 1
	return tokenizer
}

func (tokenizer *TTokenizer) forward() {
	if tokenizer.look == '\n' {
		tokenizer.line++
		tokenizer.colm = 1
	} else {
		tokenizer.colm++
	}
	tokenizer.indx++
	if tokenizer.indx >= len(tokenizer.Data) {
		tokenizer.look = -1
	} else {
		tokenizer.look = tokenizer.Data[tokenizer.indx]
	}
}

// API:Export
func (tokenizer *TTokenizer) IsEof() bool {
	return tokenizer.indx >= tokenizer.size
}

func (tokenizer *TTokenizer) isWht() bool {
	return unicode.IsSpace(tokenizer.look)
}

func (tokenizer *TTokenizer) isIdn() bool {
	return unicode.IsLetter(tokenizer.look) ||
		tokenizer.look == '_'
}

func (tokenizer *TTokenizer) isNum() bool {
	return unicode.IsDigit(tokenizer.look)
}

func (tokenizer *TTokenizer) isHex() bool {
	return unicode.IsDigit(tokenizer.look) ||
		(tokenizer.look >= 'a' && tokenizer.look <= 'f') ||
		(tokenizer.look >= 'A' && tokenizer.look <= 'F')
}

func (tokenizer *TTokenizer) isOct() bool {
	return tokenizer.look >= '0' && tokenizer.look <= '7'
}

func (tokenizer *TTokenizer) isBin() bool {
	return tokenizer.look == '0' || tokenizer.look == '1'
}

func (tokenizer *TTokenizer) isStr() bool {
	return tokenizer.look == '"'
}

func (tokenizer *TTokenizer) ignWht() {
	for !tokenizer.IsEof() && tokenizer.isWht() {
		tokenizer.forward()
	}
}

func (tokenizer *TTokenizer) getIdn() TToken {
	value := ""
	position := InitPositionFromLineAndColm(
		tokenizer.line,
		tokenizer.colm,
	)
	if !tokenizer.isIdn() {
		RaiseSystemError(fmt.Sprintf("invalid identifier start %s", string(tokenizer.look)))
	}
	for !tokenizer.IsEof() && (tokenizer.isIdn() || (tokenizer.isNum() && len(value) > 0)) {
		value += string(tokenizer.look)
		tokenizer.forward()
	}
	ttype := TokenIDN
	if IsKeyword(value) {
		ttype = TokenKEY
	}
	return TToken{
		Type:     ttype,
		Value:    value,
		Position: position,
	}
}

func (tokenizer *TTokenizer) getNum() TToken {
	value := ""
	position := InitPositionFromLineAndColm(
		tokenizer.line,
		tokenizer.colm,
	)
	if !tokenizer.isNum() {
		RaiseSystemError(fmt.Sprintf("invalid number start %s", string(tokenizer.look)))
	}
	for !tokenizer.IsEof() && tokenizer.isNum() {
		value += string(tokenizer.look)
		tokenizer.forward()
	}
	if value == "0" {
		switch tokenizer.look {
		case 'x', 'X':
			value += string(tokenizer.look)
			tokenizer.forward()
			if !tokenizer.isHex() {
				RaiseSystemError(fmt.Sprintf("invalid hex number %s", value))
			}
			for !tokenizer.IsEof() && tokenizer.isHex() {
				value += string(tokenizer.look)
				tokenizer.forward()
			}
		case 'o', 'O':
			value += string(tokenizer.look)
			tokenizer.forward()
			if !tokenizer.isOct() {
				RaiseSystemError(fmt.Sprintf("invalid oct number %s", value))
			}
			for !tokenizer.IsEof() && tokenizer.isOct() {
				value += string(tokenizer.look)
				tokenizer.forward()
			}
		case 'b', 'B':
			value += string(tokenizer.look)
			tokenizer.forward()
			if !tokenizer.isBin() {
				RaiseSystemError(fmt.Sprintf("invalid bin number %s", value))
			}
			for !tokenizer.IsEof() && tokenizer.isBin() {
				value += string(tokenizer.look)
				tokenizer.forward()
			}
		}
		if value != "0" {
			return TToken{
				Type:     TokenINT,
				Value:    value,
				Position: position,
			}
		}
	}

	ttype := TokenINT

	if tokenizer.look == '.' {
		ttype = TokenNum
		value += string(tokenizer.look)
		tokenizer.forward()
		if !tokenizer.isNum() {
			RaiseSystemError(fmt.Sprintf("invalid number %s", value))
		}
		for !tokenizer.IsEof() && tokenizer.isNum() {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
	}

	if tokenizer.look == 'e' || tokenizer.look == 'E' {
		ttype = TokenNum
		value += string(tokenizer.look)
		tokenizer.forward()
		if tokenizer.look == '+' || tokenizer.look == '-' {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
		if !tokenizer.isNum() {
			RaiseSystemError(fmt.Sprintf("invalid number %s", value))
		}
		for !tokenizer.IsEof() && tokenizer.isNum() {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
	}
	return TToken{
		Type:     ttype,
		Value:    value,
		Position: position,
	}
}

func (tokenizer *TTokenizer) getStr() TToken {
	value := ""
	position := InitPositionFromLineAndColm(
		tokenizer.line,
		tokenizer.colm,
	)
	if !tokenizer.isStr() {
		RaiseSystemError(fmt.Sprintf("invalid string start %s", string(tokenizer.look)))
	}
	op := tokenizer.isStr()
	cl := false
	tokenizer.forward()
	cl = tokenizer.isStr()
	for !tokenizer.IsEof() && !(op && cl) {
		if tokenizer.look == '\n' {
			break
		}
		if tokenizer.look == '\\' {
			tokenizer.forward()
			if tokenizer.look == 'b' {
				value += "\b"
			} else if tokenizer.look == 'f' {
				value += "\f"
			} else if tokenizer.look == 'n' {
				value += "\n"
			} else if tokenizer.look == 'r' {
				value += "\r"
			} else if tokenizer.look == 't' {
				value += "\t"
			} else if tokenizer.look == '\\' {
				value += "\\"
			} else if tokenizer.look == '"' {
				value += "\""
			} else if tokenizer.look == '\'' {
				value += "'"
			} else {
				value += string(tokenizer.look)
			}
		} else {
			value += string(tokenizer.look)
		}
		tokenizer.forward()
		cl = tokenizer.isStr()
	}
	if !(op && cl) {
		RaiseSystemError(fmt.Sprintf("string not properly closed %s", value))
	}
	tokenizer.forward()
	return TToken{
		Type:     TokenSTR,
		Value:    value,
		Position: position,
	}
}

func (tokenizer *TTokenizer) getSym() TToken {
	value := ""
	position := InitPositionFromLineAndColm(
		tokenizer.line,
		tokenizer.colm,
	)
	switch tokenizer.look {
	case '(',
		')',
		'{',
		'}',
		'[',
		']',
		':',
		';',
		',',
		'.':
		value += string(tokenizer.look)
		tokenizer.forward()
	case '?':
		value += string(tokenizer.look)
		tokenizer.forward()
		if tokenizer.look == '.' {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
	case '*':
		value += string(tokenizer.look)
		tokenizer.forward()
		if tokenizer.look == '=' {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
	case '/':
		value += string(tokenizer.look)
		tokenizer.forward()
		if tokenizer.look == '/' {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
	case '%':
		value += string(tokenizer.look)
		tokenizer.forward()
		if tokenizer.look == '=' {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
	case '+':
		value += string(tokenizer.look)
		tokenizer.forward()
		if tokenizer.look == '+' ||
			tokenizer.look == '=' {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
	case '-':
		value += string(tokenizer.look)
		tokenizer.forward()
		if tokenizer.look == '-' ||
			tokenizer.look == '=' {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
	case '<':
		value += string(tokenizer.look)
		tokenizer.forward()
		if tokenizer.look == '<' {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
		if tokenizer.look == '=' {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
	case '>':
		value += string(tokenizer.look)
		tokenizer.forward()
		if tokenizer.look == '>' {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
		if tokenizer.look == '=' {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
	case '=':
		value += string(tokenizer.look)
		tokenizer.forward()
		if tokenizer.look == '=' {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
	case '!':
		value += string(tokenizer.look)
		tokenizer.forward()
		if tokenizer.look == '=' {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
	case '&':
		value += string(tokenizer.look)
		tokenizer.forward()
		if tokenizer.look == '&' ||
			tokenizer.look == '=' {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
	case '|':
		value += string(tokenizer.look)
		tokenizer.forward()
		if tokenizer.look == '|' ||
			tokenizer.look == '=' {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
	case '^':
		value += string(tokenizer.look)
		tokenizer.forward()
		if tokenizer.look == '=' {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
	case '~':
		value += string(tokenizer.look)
		tokenizer.forward()
		if tokenizer.look == '=' {
			value += string(tokenizer.look)
			tokenizer.forward()
		}
	default:
		RaiseSystemError(fmt.Sprintf("invalid symbol %s", string(tokenizer.look)))
	}
	return TToken{
		Type:     TokenSYM,
		Value:    value,
		Position: position,
	}
}

func (tokenizer *TTokenizer) getEof() TToken {
	position := InitPositionFromLineAndColm(
		tokenizer.line,
		tokenizer.colm,
	)
	return TToken{
		Type:     TokenEOF,
		Value:    "<eof/>",
		Position: position,
	}
}

// API:Export
func (tokenizer *TTokenizer) Next() TToken {
	for !tokenizer.IsEof() {
		if tokenizer.isWht() {
			tokenizer.ignWht()
		} else if tokenizer.isIdn() {
			return tokenizer.getIdn()
		} else if tokenizer.isNum() {
			return tokenizer.getNum()
		} else if tokenizer.isStr() {
			return tokenizer.getStr()
		} else {
			return tokenizer.getSym()
		}
	}
	return tokenizer.getEof()
}
