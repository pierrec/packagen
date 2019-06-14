package packagen

import (
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/pierrec/packagen/internal/par"
	"golang.org/x/tools/go/packages"
)

var localPkgNameCache par.Cache

// localPkgName attempts to determine the name of the package in the current directory.
func localPkgName() (string, error) {
	type result struct {
		name string
		err  error
	}
	if res, ok := localPkgNameCache.Get("").(*result); ok {
		return res.name, res.err
	}
	res := localPkgNameCache.Do("", func() interface{} {
		if gofile := os.Getenv("OSFILE"); gofile != "" {
			// Fast path.
			f, err := parser.ParseFile(token.NewFileSet(), gofile, nil, parser.PackageClauseOnly)
			if err != nil {
				return &result{err: err}
			}
			return &result{name: f.Name.Name}
		}
		// Use the current working directory package name.
		pkgs, err := loadPkg(".")
		if err != nil {
			return &result{err: err}
		}
		return &result{name: pkgs[0].Name}
	}).(*result)
	return res.name, res.err
}

// Cache the results of packages.Load as getting them is expensive.
var pkgCache par.Cache

// loadPkg loads the packages matching the patterns, as per golang.org/x/tools/go/packages.Load().
// The result is cached and returned upon subsequent calls.
func loadPkg(patterns ...string) ([]*packages.Package, error) {
	type result struct {
		pkgs []*packages.Package
		err  error
	}
	key := strings.Join(patterns, " ")
	res := pkgCache.Do(key, func() interface{} {
		// Only declare the minimum load modes.
		mode := packages.NeedName |
			packages.NeedImports | packages.NeedDeps |
			packages.NeedTypes | packages.NeedTypesSizes |
			packages.NeedSyntax | packages.NeedTypesInfo

		pkgs, err := packages.Load(&packages.Config{Mode: mode}, patterns...)
		return &result{pkgs, err}
	}).(*result)
	return res.pkgs, res.err
}
