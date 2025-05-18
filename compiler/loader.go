package main

import (
	"dev/types"
)

func Load(env *TEnv) {
	// Define only the neccessary symbols|global variables

	// Define the println function
	DefineSymbol(
		env,
		"println",
		"fmt.Println",
		"fmt",
		types.TFunc(true, []*types.TPair{types.CreatePair("value", types.TAny())}, types.TVoid()),
	)

	// Define the print function
	DefineSymbol(
		env,
		"print",
		"fmt.Print",
		"fmt",
		types.TFunc(true, []*types.TPair{types.CreatePair("value", types.TAny())}, types.TVoid()),
	)

	// Define the append function
	DefineSymbol(
		env,
		"append",
		"append",
		"",
		types.TFunc(
			true,
			[]*types.TPair{
				types.CreatePair("slice", types.TArray(types.TAny())),
				types.CreatePair("value", types.TAny()),
			},
			types.TArray(types.TAny()),
		),
	)

	// Define the panic function
	DefineSymbol(
		env,
		"panic",
		"panic",
		"",
		types.TFunc(true, []*types.TPair{types.CreatePair("value", types.TAny())}, types.TVoid()),
	)
}
