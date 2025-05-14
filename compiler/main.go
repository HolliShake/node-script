package main

import (
	"fmt"
	"os"
)

func main() {
	// This is a simple Go program that does nothing.
	// You can add your code here to implement the desired functionality.
	goBinding := CreateGo()
	fmt.Println("Go path:", goBinding.GetGo())
	path := ToAbsolutePath("syntax.ns")
	data, err := os.ReadFile(path)
	if err != nil {
		RaiseSystemError(fmt.Sprintf("error reading file %s", path))
	}
	tstate := CreateState()
	parser := CreateParser(path, string(data))
	ast := parser.Parse()
	files := ForwardDeclairation(tstate, path, parser.Tokenizer.Data, ast)

	tstate.SetFile(files)

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
		fmt.Println(analyzer.src)
	}
}
