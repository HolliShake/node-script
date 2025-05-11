package main

const (
	// Keywords
	KeyStruct   = "struct"
	KeyDefine   = "define"
	KeyImport   = "import"
	KeyFrom     = "from"
	KeyVar      = "var"
	KeyLocal    = "local"
	KeyConst    = "const"
	KeyFor      = "for"
	KeyDo       = "do"
	KeyWhile    = "while"
	KeyIf       = "if"
	KeyElse     = "else"
	KeySwitch   = "switch"
	KeyCase     = "case"
	KeyDefault  = "default"
	KeyBreak    = "break"
	KeyContinue = "continue"
	KeyReturn   = "return"
	KeyTrue     = "true"
	KeyFalse    = "false"
	KeyNull     = "null"
	keyAwait    = "await"
	KeyInt8     = "i8"   // Typing
	KeyInt16    = "i16"  // Typing
	KeyInt32    = "i32"  // Typing
	KeyInt64    = "i64"  // Typing
	KeyNum      = "num"  // Typing
	KeyStr      = "str"  // Typing
	KeyBool     = "bool" // Typing
	KeyVoid     = "void" // Typing
)

var Keywords = []string{
	KeyStruct,
	KeyDefine,
	KeyImport,
	KeyFrom,
	KeyVar,
	KeyLocal,
	KeyConst,
	KeyFor,
	KeyDo,
	KeyWhile,
	KeyIf,
	KeyElse,
	KeySwitch,
	KeyCase,
	KeyDefault,
	KeyBreak,
	KeyContinue,
	KeyReturn,
	KeyTrue,
	KeyFalse,
	KeyNull,
	KeyInt8,
	KeyInt16,
	KeyInt32,
	KeyInt64,
	KeyNum,
	KeyStr,
	KeyBool,
	KeyVoid,
}

func IsKeyword(str string) bool {
	for _, keyword := range Keywords {
		if str == keyword {
			return true
		}
	}
	return false
}
