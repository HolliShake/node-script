package main

import (
	"os"
	"path/filepath"
	"strings"
)

func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		RaiseSystemError(err)
	}

	return info.IsDir()
}

func IsAbsolutePath(path string) bool {
	// Check if the path is absolute
	return filepath.IsAbs(path)
}

func GetFileName(path string) string {
	return filepath.Base(path)
}

func GetFileNameWithoutExtension(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func GetFileExtension(path string) string {
	return filepath.Ext(path)
}

func GetDir(path string) string {
	if !IsAbsolutePath(path) {
		RaiseSystemError("path must be an absolute path")
	}
	return filepath.Dir(path)
}

func ToAbsolutePath(path string) string {
	if IsAbsolutePath(path) && IsDir(path) {
		return path
	}
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		RaiseSystemError(err)
	}
	// Join the current directory with the relative path
	absolutePath, err := filepath.Abs(filepath.Join(currentDir, path))
	if err != nil {
		RaiseSystemError(err)
	}

	return absolutePath
}

func ResolvePath(currentDir string, relativePath string) string {
	if !(IsAbsolutePath(currentDir) && IsDir(currentDir)) {
		RaiseSystemError("currentDir must be an absolute path")
	}
	// Check if the path is already absolute
	if IsAbsolutePath(relativePath) {
		return relativePath
	}
	finalPath := ""
	if strings.HasPrefix(relativePath, "./") {
		finalPath = filepath.Join(currentDir, strings.TrimPrefix(relativePath, "./"))
	} else if strings.HasPrefix(relativePath, "../") {
		withDepth := relativePath
		current := currentDir
		for strings.HasPrefix(withDepth, "../") {
			withDepth = strings.TrimPrefix(withDepth, "../")
			current = filepath.Dir(current)
		}
		finalPath = filepath.Join(current, withDepth)
	} else {
		RaiseSystemError("invalid path, path must be in a form of relative path")
	}
	return finalPath
}
