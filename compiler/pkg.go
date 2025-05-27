package main

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func GetPackages(path string) []*packages.Package {
	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedSyntax, Tests: false}

	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		panic(err)
	}
	return pkgs
}

func PackagesHasName(pkgs []*packages.Package, name string) bool {
	for _, pkg := range pkgs {
		fmt.Println("Checking package:", pkg.PkgPath)
		obj := pkg.Types.Scope().Lookup(name)
		if obj != nil {
			return true
		}
	}
	return false
}

func PackagesGetName(pkgs []*packages.Package, name string) types.Object {
	for _, pkg := range pkgs {
		fmt.Println("Checking package:", pkg.PkgPath)
		obj := pkg.Types.Scope().Lookup(name)
		if obj != nil {
			return obj
		}
	}
	return nil
}
