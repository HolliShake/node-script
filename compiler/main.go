package main

import (
	"fmt"
	"os"
)

func parseArgs() map[string]string {
	args := os.Args[1:]
	result := make(map[string]string)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if len(arg) > 0 && arg[0] == '-' {
			var key string
			if len(arg) > 1 && arg[1] == '-' {
				// Handle --flag format
				key = arg[2:]
			} else {
				// Handle -flag format
				key = arg[1:]
			}

			// Skip empty keys
			if key == "" {
				continue
			}

			// Check if there's a value following this flag
			if i+1 < len(args) && (len(args[i+1]) == 0 || args[i+1][0] != '-') {
				result[key] = args[i+1]
				i++ // Skip the value in the next iteration
			} else {
				// Flag without value, set it to "true"
				result[key] = "true"
			}
		}
	}
	return result
}

func main() {
	// This is a simple Go program that does nothing.
	// You can add your code here to implement the desired functionality.
	goBinding := CreateGo()

	fmt.Println(parseArgs())

	// Read the syntax.ns file.
	path := ToAbsolutePath("./library/syntax.ns")
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
