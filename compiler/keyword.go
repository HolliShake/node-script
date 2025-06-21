package main

const (
	// Keywords
	KeyStruct   = "struct"
	KeyFunction = "function"
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
	KeyRun      = "run"
	KeyContinue = "continue"
	KeyBreak    = "break"
	KeyReturn   = "return"
	KeyPanics   = "panics"
	KeyTrue     = "true"
	KeyFalse    = "false"
	KeyNull     = "null"
	KeyNew      = "new"
	KeyInt8     = "i8"    // Typing
	KeyInt16    = "i16"   // Typing
	KeyInt32    = "i32"   // Typing
	KeyInt64    = "i64"   // Typing
	KeyNum      = "num"   // Typing
	KeyStr      = "str"   // Typing
	KeyBool     = "bool"  // Typing
	KeyVoid     = "void"  // Typing
	KeyError    = "error" // Typing
)

var Keywords = []string{
	KeyStruct,
	KeyFunction,
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
	KeyRun,
	KeyContinue,
	KeyBreak,
	KeyReturn,
	KeyPanics,
	KeyTrue,
	KeyFalse,
	KeyNull,
	KeyNew,
	KeyInt8,
	KeyInt16,
	KeyInt32,
	KeyInt64,
	KeyNum,
	KeyStr,
	KeyBool,
	KeyVoid,
	KeyError,
}

func IsKeyword(str string) bool {
	for _, keyword := range Keywords {
		if str == keyword {
			return true
		}
	}
	return false
}
