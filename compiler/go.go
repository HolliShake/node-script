package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type TGoBinding struct {
	// Cache paths to avoid repeated calculations
	execPathCache  string
	cachePath      string
	goPathCache    string
	goFmtPathCache string
	// Mutex to protect concurrent access to cache fields
	mu sync.RWMutex
}

func CreateGo() *TGoBinding {
	return &TGoBinding{}
}

func (g *TGoBinding) getExecPath() (string, error) {
	// Check cache with read lock first
	g.mu.RLock()
	if path := g.execPathCache; path != "" {
		g.mu.RUnlock()
		return path, nil
	}
	g.mu.RUnlock()

	executablePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	actualPath := filepath.Dir(executablePath)

	if _, err := os.Stat(actualPath); os.IsNotExist(err) {
		return "", errors.New("go executable path does not exist")
	}

	// Update cache with write lock
	g.mu.Lock()
	g.execPathCache = actualPath
	g.mu.Unlock()

	return actualPath, nil
}

func (g *TGoBinding) getCachePath() (string, error) {
	// Check cache with read lock first
	g.mu.RLock()
	if path := g.cachePath; path != "" {
		g.mu.RUnlock()
		return path, nil
	}
	g.mu.RUnlock()

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cachePath := filepath.Join(dir, "__cache__")

	// Update cache with write lock
	g.mu.Lock()
	g.cachePath = cachePath
	g.mu.Unlock()

	return cachePath, nil
}

// API:Export
func (g *TGoBinding) GetGo() (string, error) {
	// Check cache with read lock first
	g.mu.RLock()
	if path := g.goPathCache; path != "" {
		g.mu.RUnlock()
		return path, nil
	}
	g.mu.RUnlock()

	goPath, err := g.getExecPath()
	if err != nil {
		return "", err
	}

	// Check bundled Go first
	bundledGoPath := filepath.Join(goPath, "thirdparty/go/bin/go")
	if _, err := os.Stat(bundledGoPath); err == nil {
		g.mu.Lock()
		g.goPathCache = bundledGoPath
		g.mu.Unlock()
		return bundledGoPath, nil
	}

	// Then check system Go
	if systemGoPath, err := exec.LookPath("go"); err == nil {
		g.mu.Lock()
		g.goPathCache = systemGoPath
		g.mu.Unlock()
		return systemGoPath, nil
	}

	return "", errors.New("go executable path does not exist")
}

// API:Export
func (g *TGoBinding) GetGoFmt() (string, error) {
	// Check cache with read lock first
	g.mu.RLock()
	if path := g.goFmtPathCache; path != "" {
		g.mu.RUnlock()
		return path, nil
	}
	g.mu.RUnlock()

	goPath, err := g.getExecPath()
	if err != nil {
		return "", err
	}

	// Check bundled gofmt first
	bundledGoFmtPath := filepath.Join(goPath, "thirdparty/go/bin/gofmt")
	if _, err := os.Stat(bundledGoFmtPath); err == nil {
		g.mu.Lock()
		g.goFmtPathCache = bundledGoFmtPath
		g.mu.Unlock()
		return bundledGoFmtPath, nil
	}

	// Then check system gofmt
	if systemGoFmtPath, err := exec.LookPath("gofmt"); err == nil {
		g.mu.Lock()
		g.goFmtPathCache = systemGoFmtPath
		g.mu.Unlock()
		return systemGoFmtPath, nil
	}

	return "", errors.New("gofmt executable path does not exist")
}

// API:Export
func (g *TGoBinding) InitGoModToCache() (bool, error) {
	goPath, err := g.GetGo()
	if err != nil {
		return false, err
	}

	cachePath, err := g.getCachePath()
	if err != nil {
		return false, err
	}

	modulePath := "script/main"
	goModPath := filepath.Join(cachePath, "go.mod")

	// Check if the cache directory exists.
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		if err := os.MkdirAll(cachePath, 0755); err != nil {
			return false, err
		}
	}

	// Check if the module already exists.
	if _, err := os.Stat(goModPath); !os.IsNotExist(err) {
		return true, nil
	}

	cmd := exec.Command(goPath, "mod", "init", modulePath)
	cmd.Dir = cachePath

	if output, err := cmd.CombinedOutput(); err != nil {
		return false, errors.New(string(output))
	}

	return true, nil
}

// API:Export
func (g *TGoBinding) GoExecFmt(file string) (bool, error) {
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

	if _, err := cmd.CombinedOutput(); err != nil {
		return false, err
	}

	return true, nil
}

// API:Export
func (g *TGoBinding) GoRunCache() (string, error) {
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
func (g *TGoBinding) GoCompileCache(scriptPath string, output string) (bool, error) {
	goPath, err := g.GetGo()
	if err != nil {
		return false, err
	}

	cachePath, err := g.getCachePath()
	if err != nil {
		return false, err
	}

	cmd := exec.Command(goPath, "build", "-o", fmt.Sprintf("%s.exe", output), cachePath)
	cmd.Dir = scriptPath

	if _, err := cmd.CombinedOutput(); err != nil {
		return false, err
	}

	return true, nil
}

// API:Export
func (g *TGoBinding) Generate(file string, data string) (bool, error) {
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
	return g.GoExecFmt(cacheFile)
}
