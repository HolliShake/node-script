package main

import (
	"fmt"
	"os"
)

func main() {
	// This is a simple Go program that does nothing.
	// You can add your code here to implement the desired functionality.
	fmt.Println("Hello, World!")

	path := ToAbsolutePath("syntax.ns")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	tstate := CreateState()
	parser := CreateParser(path, string(data))
	ast := parser.Parse()
	ForwardDeclairation(tstate, path, parser.Tokenizer.Data, ast)
}
