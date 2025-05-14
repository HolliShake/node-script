package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

type TGo struct {
}

func CreateGo() *TGo {
	return &TGo{}
}

func (g *TGo) getExecPath() string {
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

func (g *TGo) getCachePath() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return filepath.Join(dir, "__cache__")
}

// API:Export
func (g *TGo) GetGo() string {
	goPath := filepath.Join(g.getExecPath(), "thirdparty/go/bin/go")
	if _, err := os.Stat(goPath); err == nil {
		return goPath
	}
	if goPath, err := exec.LookPath("go"); err == nil {
		return goPath
	}
	RaiseSystemError("go executable path does not exist")
	return "[invalid]"
}

// API:Export
func (g *TGo) Generate(file string, data string) (bool, error) {
	cachePath := g.getCachePath()
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		os.MkdirAll(cachePath, 0755)
	}
	cacheFile := filepath.Join(cachePath, file)
	err := os.WriteFile(cacheFile, []byte(data), 0644)
	return err == nil, err
}
