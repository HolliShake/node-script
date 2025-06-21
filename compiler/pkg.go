package main

import (
	"fmt"
	"go/types"
	"sync"

	"golang.org/x/tools/go/packages"
)

// thread-safe cache with pre-allocated capacity
var (
	packagesCache = make(map[string][]*packages.Package, 32) // pre-allocate for common case
	cacheMutex    sync.RWMutex
)

// loadPackages loads packages with caching.
func loadPackages(path string, full bool) ([]*packages.Package, error) {
	// Fast path: check cache first with read lock
	cacheMutex.RLock()
	cached, ok := packagesCache[path]
	cacheMutex.RUnlock()
	if ok {
		return cached, nil
	}

	// Prepare config based on need - reuse common config settings
	cfg := &packages.Config{
		// Optimized build flags
		BuildFlags: []string{"-gcflags=-N -l"},
	}

	// Set mode flags with bitwise OR for better performance
	if full {
		cfg.Mode = packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedDeps | packages.NeedSyntax | packages.NeedImports
	} else {
		cfg.Mode = packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo
	}

	// Load packages
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	// Only cache non-empty results to save memory
	if len(pkgs) > 0 {
		cacheMutex.Lock()
		packagesCache[path] = pkgs
		cacheMutex.Unlock()
	}

	return pkgs, nil
}

// HasGoPackage returns true if the given import path resolves to at least one Go package.
// Uses minimal loading configuration for speed.
func HasGoPackage(path string) bool {
	pkgs, err := loadPackages(path, false)
	return err == nil && len(pkgs) > 0
}

// GetGoPackages loads Go packages with full type and syntax information.
// Panics on error or if no valid packages are found.
func GetGoPackages(path string) []*packages.Package {
	pkgs, err := loadPackages(path, true)
	if err != nil {
		panic(fmt.Sprintf("failed to load package %q: %v", path, err))
	}

	// Combined validation check to reduce branching
	if packages.PrintErrors(pkgs) > 0 || len(pkgs) == 0 {
		panic(fmt.Sprintf("failed to load valid package(s) for path %q", path))
	}

	// Fast validation loop with early return on error
	for i, pkg := range pkgs {
		if pkg.Types == nil {
			panic(fmt.Sprintf("missing type info for package %q (index %d)", pkg.PkgPath, i))
		}
	}
	return pkgs
}

// PackagesHasName checks if any loaded package defines a top-level object with the given name.
func PackagesHasName(pkgs []*packages.Package, name string) bool {
	return PackagesGetName(pkgs, name) != nil
}

// PackagesGetName returns the first object with the given name found in the package scope.
func PackagesGetName(pkgs []*packages.Package, name string) types.Object {
	// Optimized loop with nil checks first to avoid unnecessary lookups
	for _, pkg := range pkgs {
		if pkg.Types != nil {
			if obj := pkg.Types.Scope().Lookup(name); obj != nil {
				return obj
			}
		}
	}
	return nil
}

// IsGoStruct returns true if the given type is a struct or named struct.
// Fast path for nil check improves performance.
func IsGoStruct(t types.Type) bool {
	if t == nil {
		return false
	}
	_, ok := t.Underlying().(*types.Struct)
	return ok
}
