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
		panic("NodeError: node executable path does not exist")
	}

	return filepath.Dir(executablePath)
}

func ExecuteFile(filePath string, args ...string) {
	localNodePath := getExecPath() + "/thirdparty/node/bin/node"

	nodeArgs := append(
		[]string{filePath},
		args...,
	)
	cmd := exec.Command(localNodePath, nodeArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
