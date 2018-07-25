package parseutil

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ErrTooManyPackages is returned when there is more than one package in a
// directory where there should only be one Go package.
var ErrTooManyPackages = errors.New("more than one package found in a directory")

// PackageAST returns the AST of the package at the given path.
func PackageAST(path string) (pkg *ast.Package, err error) {
	return parseAndFilterPackages(path, func(k string, v *ast.Package) bool {
		return !strings.HasSuffix(k, "_test")
	})
}

// PackageTestAST returns the AST of the test package at the given path.
func PackageTestAST(path string) (pkg *ast.Package, err error) {
	return parseAndFilterPackages(path, func(k string, v *ast.Package) bool {
		return strings.HasSuffix(k, "_test")
	})
}

type packageFilter func(string, *ast.Package) bool

// filteredPackages filters the parsed packages and then makes sure there is only
// one left.
func parseAndFilterPackages(path string, filter packageFilter) (pkg *ast.Package, err error) {
	srcDir, err := DefaultGoPath.Abs(path)
	if err != nil {
		goRoot := os.Getenv("GOROOT")
		if goRoot == "" {
			goRoot = "/usr/local/go"
		}
		rootPath := filepath.Join(goRoot, "src", path)
		_, err = os.Stat(rootPath)
		if err != nil {
			return nil, fmt.Errorf("package %s not found in GOPATH or GOROOT", path)
		}
		srcDir = rootPath
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, srcDir, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	pkgs = filterPkgs(pkgs, filter)

	dirName := filepath.Base(srcDir)

	if len(pkgs) == 1 {
		for _, p := range pkgs {
			pkg = p
		}
	} else {
		if p, ok := pkgs[dirName]; ok {
			pkg = p
		} else {
			return nil, ErrTooManyPackages
		}
	}

	return
}

func filterPkgs(pkgs map[string]*ast.Package, filter packageFilter) map[string]*ast.Package {
	filtered := make(map[string]*ast.Package)
	for k, v := range pkgs {
		if filter(k, v) {
			filtered[k] = v
		}
	}

	return filtered
}
