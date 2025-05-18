package main

import (
	"fmt"
	"os"
)

func main() {
	// This is a simple Go program that does nothing.
	// You can add your code here to implement the desired functionality.
	goBinding := CreateGo()

	// Read the syntax.ns file.
	path := ToAbsolutePath("syntax.ns")
	data, err := os.ReadFile(path)

	if err != nil {
		RaiseSystemError(fmt.Sprintf("error reading file %s", path))
	}

	// Create the state.
	tstate := CreateState()

	// Parse the syntax.ns file.
	parser := CreateParser(path, string(data))
	ast := parser.Parse()

	// Forward declaration.
	files := ForwardDeclairation(tstate, path, parser.Tokenizer.Data, ast)
	tstate.SetFile(files)

	// Initialize the Go module.
	_, modInitErr := goBinding.InitGoModToCache()
	if modInitErr != nil {
		RaiseSystemError(string(modInitErr.Error()))
	}

	// Analyze the syntax.ns file.
	for _, file := range files {
		analyzer := CreateAnalyzer(tstate, file)
		src := analyzer.Analyze()
		ok, err := goBinding.Generate(GetFileNameWithoutExtension(file.Path)+".go", src)
		if err != nil {
			RaiseSystemError(fmt.Sprintf("error generating file %s", path))
		}
		if !ok {
			RaiseSystemError(fmt.Sprintf("error generating file %s", path))
		}
	}
	// Collect and free the memory.
	CollectAndFree()

	// Run the Go program.
	out, err := goBinding.GoRunCache()
	if err != nil {
		RaiseSystemError(fmt.Sprintf("error running Go program: %s", err))
	}
	fmt.Print(out)
}
