package main

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func HasGoPackage(path string) bool {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedSyntax | packages.NeedImports,
	}

	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return false
	}

	return len(pkgs) > 0
}

func GetGoPackages(path string) []*packages.Package {
	if !HasGoPackage(path) {
		panic(fmt.Sprintf("package %s not found", path))
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedSyntax | packages.NeedImports,
	}

	pkgs, err := packages.Load(cfg, path)

	if err != nil {
		fmt.Printf("Error loading package %s: %v\n", path, err)
		panic(err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		panic(fmt.Sprintf("package %s load failed", path))
	}

	if len(pkgs) == 0 {
		panic(fmt.Sprintf("no packages loaded for %s", path))
	}

	// Check if Types is nil before returning to avoid nil pointer dereference
	for i, pkg := range pkgs {
		if pkg == nil {
			panic(fmt.Sprintf("package at index %d is nil for %s", i, path))
		}
		if pkg.Types == nil {
			panic(fmt.Sprintf("Types info is nil for package %s at index %d", pkg.PkgPath, i))
		}
	}

	return pkgs
}

func PackagesHasName(pkgs []*packages.Package, name string) bool {
	for _, pkg := range pkgs {
		obj := pkg.Types.Scope().Lookup(name)
		if obj != nil {
			return true
		}
	}
	return false
}

func PackagesGetName(pkgs []*packages.Package, name string) types.Object {
	for _, pkg := range pkgs {
		obj := pkg.Types.Scope().Lookup(name)
		if obj != nil {
			return obj
		}
	}
	return nil
}
