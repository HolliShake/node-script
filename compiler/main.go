package main

import (
	"fmt"
	"os"
	"time"
)

const (
	FLAG_COMPILE = "compile"
	FLAG_OUT     = "out"
	FLAG_RUN     = "run"
)

func parseArgs() map[string]string {
	args := os.Args[1:]
	result := make(map[string]string)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if len(arg) == 0 || arg[0] != '-' {
			continue
		}

		// Get key (strip leading dashes)
		var key string
		if arg[1] == '-' {
			key = arg[2:] // --flag format
		} else {
			key = arg[1:] // -flag format
		}

		if key == "" {
			continue
		}

		// Check if there's a value following this flag
		if i+1 < len(args) && (len(args[i+1]) == 0 || args[i+1][0] != '-') {
			result[key] = args[i+1]
			i++ // Skip the value in the next iteration
		} else {
			result[key] = "true" // Flag without value
		}
	}
	return result
}

func showHelp() {
	fmt.Println("Parrot Script")
	fmt.Println("Author: Philipp Andrew Roa Redondo")
	fmt.Println("License: GNU GENERAL PUBLIC LICENSE V3")
	fmt.Println("Available options:")
	fmt.Println("  --compile <file>  Compile the specified file")
	fmt.Println("  --out     <file>  Specify the output file for compilation")
	fmt.Println("  --run     <file>  Run the specified file")
	fmt.Println("\nExamples:")
	fmt.Println("  To compile:  parrot --compile myfile.ns --out output")
	fmt.Println("  To run:      parrot --run myfile.ns")
}

func processArgs(goBinding *TGoBinding, args map[string]string) {
	if compileFile := args[FLAG_COMPILE]; compileFile != "true" && compileFile != "" {
		output := args[FLAG_OUT]
		if output == "true" || output == "" {
			RaiseSystemError("error compiling, output file is not specified")
		}
		compile(goBinding, compileFile, output)
	} else if runFile := args[FLAG_RUN]; runFile != "true" && runFile != "" {
		run(goBinding, runFile)
	} else {
		showHelp()
	}
}

func processFile(goBinding *TGoBinding, path string) {
	// Read the file
	absPath := ToAbsolutePath(path)
	data, err := os.ReadFile(absPath)
	if err != nil {
		RaiseSystemError(fmt.Sprintf("error reading file %s", absPath))
	}

	// Generate the array code
	startTime := time.Now()

	// Start a goroutine to print elapsed seconds
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				elapsed := time.Since(startTime)
				fmt.Printf("Compiling... %d seconds elapsed\r", int(elapsed.Seconds()))
			case <-done:
				return
			}
		}
	}()

	// Create the state
	tstate := CreateState()

	// Parse the file
	parser := CreateParser(absPath, string(data))
	ast := parser.Parse()

	// Forward declaration
	files := ForwardDeclairation(tstate, absPath, parser.Tokenizer.Data, ast)
	tstate.SetFile(files)

	// Initialize Go module
	_, modInitErr := goBinding.InitGoModToCache()
	if modInitErr != nil {
		RaiseSystemError(modInitErr.Error())
	}

	// Analyze the files
	for _, file := range files {
		analyzer := CreateAnalyzer(tstate, file)
		src := analyzer.Analyze()
		ok, err := goBinding.Generate(GetFileNameWithoutExtension(file.Path)+".go", src)
		if err != nil || !ok {
			RaiseSystemError(fmt.Sprintf("error generating file %s: %s", file.Path, err))
		}
	}

	// Generate arrays
	arrayCode := tstate.GenerateArrays()
	ok, err := goBinding.Generate("arrays.go", arrayCode)

	// Signal the goroutine to stop
	done <- true

	// Print final message
	elapsed := time.Since(startTime)
	fmt.Printf("Arrays generation completed in %d seconds\n", int(elapsed.Seconds()))

	if err != nil || !ok {
		fmt.Println(err)
		RaiseSystemError(fmt.Sprintf("error generating array.go: %s", err))
	}

	// Collect and free memory
	CollectAndFree()
}

func compile(goBinding *TGoBinding, scriptPath string, output string) {
	processFile(goBinding, scriptPath)

	// Compile the cache
	ok, err := goBinding.GoCompileCache(GetDir(ToAbsolutePath(scriptPath)), output)
	if err != nil {
		RaiseSystemError(fmt.Sprintf("error compiling cache: %s", err))
	}
	if !ok {
		RaiseSystemError("error compiling cache")
	}
	fmt.Println("compiled cache")
}

func run(goBinding *TGoBinding, scriptPath string) {
	processFile(goBinding, scriptPath)

	// Run the cache
	out, err := goBinding.GoRunCache()
	if err != nil {
		RaiseSystemError(fmt.Sprintf("error running Go program: %s", err))
	}
	fmt.Print(out)
}

func main() {
	// Parse and process arguments
	processArgs(CreateGo(), parseArgs())
}
