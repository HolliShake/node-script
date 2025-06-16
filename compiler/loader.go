package main

import (
	"dev/types"
	"strings"
)

const (
	MODULE_FMT    = "fmt"
	MODULE_OS     = "os"
	MODULE_GLOBAL = ""
)

const (
	FMT_PRINTLN  = "Println"
	FMT_PRINT    = "Print"
	GLOBAL_PANIC = "panic"
)

func Load(env *TEnv) {
	// Define only the neccessary symbols|global variables

	// Define the println function
	DefineSymbol(
		env,
		strings.ToLower(FMT_PRINTLN),
		MODULE_FMT+"."+FMT_PRINTLN,
		MODULE_FMT,
		types.TFunc(true, []*types.TPair{types.CreatePair("value", types.TAny())}, types.TVoid(), false),
	)

	// Define the print function
	DefineSymbol(
		env,
		strings.ToLower(FMT_PRINT),
		MODULE_FMT+"."+FMT_PRINT,
		MODULE_FMT,
		types.TFunc(true, []*types.TPair{types.CreatePair("value", types.TAny())}, types.TVoid(), false),
	)

	// Define the panic function
	DefineSymbol(
		env,
		strings.ToLower(GLOBAL_PANIC),
		GLOBAL_PANIC,
		MODULE_GLOBAL,
		types.TFunc(true, []*types.TPair{types.CreatePair("value", types.TAny())}, types.TVoid(), true),
	)
}
