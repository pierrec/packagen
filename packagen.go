package packagen

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/pierrec/packagen/internal/par"
	"golang.org/x/tools/go/packages"
)

// renamePkg renames all used type names with the ones in names for the given package.
func renamePkg(pkg *packages.Package, names map[string]string, ignore map[string]bool, objsToUpdate map[types.Object]bool) {
	info := pkg.TypesInfo
	for id, obj := range info.Defs {
		if _, ok := ignore[id.Name]; ok {
			objsToUpdate[obj] = false
		}
		if newname, ok := names[id.Name]; ok {
			objsToUpdate[obj] = false
			if !ignore[id.Name] {
				// Exclude types to be removed.
				id.Name = newname
			}
		}
	}
	for id, obj := range info.Uses {
		if !ignore[id.Name] {
			objsToUpdate[obj] = false
		}
		if newname, ok := names[id.Name]; ok {
			objsToUpdate[obj] = false
			id.Name = newname
		}
	}
}

// prefixPkg prefixes all global identifiers (types, variables, functions).
func prefixPkg(pkg *packages.Package, prefix string, objsToUpdate map[types.Object]bool) {
	info := pkg.TypesInfo
	// Contains all the objects to be renamed.
	// Copied from https://github.com/golang/tools/blob/master/cmd/bundle/main.go:210
	var rename func(from types.Object)
	rename = func(from types.Object) {
		if _, ok := objsToUpdate[from]; ok {
			// Ignore objects that are already updated.
			return
		}
		objsToUpdate[from] = true

		// Renaming a type that is used as an embedded field
		// requires renaming the field too. e.g.
		// 	type T int // if we rename this to U..
		// 	var s struct {T}
		// 	print(s.T) // ...this must change too
		if _, ok := from.(*types.TypeName); !ok {
			return
		}
		for id, obj := range info.Uses {
			if obj == from {
				if field := info.Defs[id]; field != nil {
					rename(field)
				}
			}
		}
	}

	// Populate the map with the objects to be prefixed.
	// Only the ones in the top package scope need to be prefixed.
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		rename(scope.Lookup(name))
	}

	// Prefix the objects.
	for id, obj := range info.Defs {
		if objsToUpdate[obj] {
			id.Name = prefix + obj.Name()
		}
	}
	for id, obj := range info.Uses {
		if objsToUpdate[obj] {
			id.Name = prefix + obj.Name()
		}
	}
}

func keysOf(m map[string]bool) []string {
	s := make([]string, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	return s
}

// localPkgName attempts to determine the name of the package in the current directory.
// If any non empty name is found, it is used even in case of an error while loading it.
func localPkgName() (string, error) {
	// Use the current working directory package name.
	pkgs, err := loadPkg(".")
	if len(pkgs) > 0 {
		// Be optimistic: even if the local package has errors, return its name.
		if p := pkgs[0].Name; p != "" {
			return p, nil
		}
	}
	if err != nil {
		return "", err
	}
	return "", fmt.Errorf("cannot define new package name")
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
		pkgs, err := packages.Load(&packages.Config{Mode: packages.LoadSyntax}, patterns...)
		return result{pkgs, err}
	}).(result)
	return res.pkgs, res.err
}
