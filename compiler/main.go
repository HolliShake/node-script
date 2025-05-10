package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func getExecPath() string {
	executablePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	// Remove filename.ext from the path
	actualPath := filepath.Dir(executablePath)

	_, ferr := os.Stat(actualPath)

	if os.IsNotExist(ferr) {
		panic("Error: go executable path does not exist")
	}

	return filepath.Dir(executablePath)
}

func getGo() string {
	goPath := filepath.Join(getExecPath(), "thirdparty/go/bin/go")
	if _, err := os.Stat(goPath); err == nil {
		return goPath
	}
	if goPath, err := exec.LookPath("go"); err == nil {
		return goPath
	}
	RaiseSystemError("go executable path does not exist")
	return "[invalid]"
}

func writeToCache(path string, data string) {
	cachePath := filepath.Join(getExecPath(), "__cache__")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		os.MkdirAll(cachePath, 0755)
	}
	cacheFile := filepath.Join(cachePath, path)
	os.WriteFile(cacheFile, []byte(data), 0644)
}

func main() {
	// This is a simple Go program that does nothing.
	// You can add your code here to implement the desired functionality.
	fmt.Println("Go path:", getGo())
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
		writeToCache(GetFileNameWithoutExtension(file.Path)+".go", src)
		fmt.Println(analyzer.src)
	}
}
