package packagen

import (
	"go/types"
	"golang.org/x/tools/go/packages"
)

// renamePkg renames all used type names with the ones in names for the given package.
func renamePkg(pkg *packages.Package, names map[string]string, objsToUpdate map[types.Object]bool) {
	info := pkg.TypesInfo
	for id, obj := range info.Defs {
		if newname, ok := names[id.Name]; ok {
			objsToUpdate[obj] = false
			id.Name = newname
		}
	}
	for id, obj := range info.Uses {
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
