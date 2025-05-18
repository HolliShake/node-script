package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

type TGo struct {
	// Cache paths to avoid repeated calculations
	execPathCache  string
	cachePath      string
	goPathCache    string
	goFmtPathCache string
}

func CreateGo() *TGo {
	return &TGo{}
}

func (g *TGo) getExecPath() (string, error) {
	// Return cached path if available
	if g.execPathCache != "" {
		return g.execPathCache, nil
	}

	executablePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	// Remove filename.ext from the path
	actualPath := filepath.Dir(executablePath)

	_, ferr := os.Stat(actualPath)
	if os.IsNotExist(ferr) {
		return "", errors.New("go executable path does not exist")
	}

	// Cache the result
	g.execPathCache = actualPath
	return actualPath, nil
}

func (g *TGo) getCachePath() (string, error) {
	// Return cached path if available
	if g.cachePath != "" {
		return g.cachePath, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cachePath := filepath.Join(dir, "__cache__")

	// Cache the result
	g.cachePath = cachePath
	return cachePath, nil
}

// API:Export
func (g *TGo) GetGo() (string, error) {
	// Return cached path if available
	if g.goPathCache != "" {
		return g.goPathCache, nil
	}

	goPath, err := g.getExecPath()
	if err != nil {
		return "", err
	}

	// Check bundled Go first
	bundledGoPath := filepath.Join(goPath, "thirdparty/go/bin/go")
	if _, err := os.Stat(bundledGoPath); err == nil {
		g.goPathCache = bundledGoPath
		return bundledGoPath, nil
	}

	// Then check system Go
	if systemGoPath, err := exec.LookPath("go"); err == nil {
		g.goPathCache = systemGoPath
		return systemGoPath, nil
	}

	return "", errors.New("go executable path does not exist")
}

// API:Export
func (g *TGo) GetGoFmt() (string, error) {
	// Return cached path if available
	if g.goFmtPathCache != "" {
		return g.goFmtPathCache, nil
	}

	goPath, err := g.getExecPath()
	if err != nil {
		return "", err
	}

	// Check bundled gofmt first
	bundledGoFmtPath := filepath.Join(goPath, "thirdparty/go/bin/gofmt")
	if _, err := os.Stat(bundledGoFmtPath); err == nil {
		g.goFmtPathCache = bundledGoFmtPath
		return bundledGoFmtPath, nil
	}

	// Then check system gofmt
	if systemGoFmtPath, err := exec.LookPath("gofmt"); err == nil {
		g.goFmtPathCache = systemGoFmtPath
		return systemGoFmtPath, nil
	}

	return "", errors.New("gofmt executable path does not exist")
}

// API:Export
func (g *TGo) InitGoModToCache() (bool, error) {
	goPath, err := g.GetGo()
	if err != nil {
		return false, err
	}
	cachePath, err := g.getCachePath()
	if err != nil {
		return false, err
	}
	modulePath := "script/main"

	// Check if the cache directory exists.
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		if err := os.MkdirAll(cachePath, 0755); err != nil {
			return false, err
		}
	}

	// Check if the module already exists.
	if _, err := os.Stat(filepath.Join(cachePath, "go.mod")); !os.IsNotExist(err) {
		return true, nil
	}

	cmd := exec.Command(goPath, "mod", "init", modulePath)
	cmd.Dir = cachePath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, errors.New(string(output))
	}
	return true, nil
}

// API:Export
func (g *TGo) GoExecFmt(file string) (bool, error) {
	goPath, err := g.GetGoFmt()
	if err != nil {
		return false, err
	}
	cachePath, err := g.getCachePath()
	if err != nil {
		return false, err
	}

	cmd := exec.Command(goPath, "-w", file)
	cmd.Dir = cachePath

	_, err = cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	return true, nil
}

// API:Export
func (g *TGo) GoRunCache() (string, error) {
	goPath, err := g.GetGo()
	if err != nil {
		return "", err
	}
	cachePath, err := g.getCachePath()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(goPath, "run", cachePath)
	cmd.Dir = cachePath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(string(output))
	}
	return string(output), nil
}

// API:Export
func (g *TGo) GoRunCompileCache() (bool, error) {
	goPath, err := g.GetGo()
	if err != nil {
		return false, err
	}
	cachePath, err := g.getCachePath()
	if err != nil {
		return false, err
	}

	cmd := exec.Command(goPath, "build", "-o", "main.exe", cachePath)
	cmd.Dir = cachePath

	_, err = cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	return true, nil
}

// API:Export
func (g *TGo) Generate(file string, data string) (bool, error) {
	cachePath, err := g.getCachePath()
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		if err := os.MkdirAll(cachePath, 0755); err != nil {
			return false, err
		}
	}

	cacheFile := filepath.Join(cachePath, file)
	if err := os.WriteFile(cacheFile, []byte(data), 0644); err != nil {
		return false, err
	}

	// Gofmt the file
	_, err = g.GoExecFmt(cacheFile)
	return err == nil, err
}
