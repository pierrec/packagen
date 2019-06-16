package packagen

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"go/types"
	"io"

	"golang.org/x/tools/go/packages"
)

// renamePkg renames all used type names with the ones in names for the given package.
func renamePkg(pkg *packages.Package, names map[string]string, ignore map[string]bool,
	objsToUpdate map[types.Object]bool, renameID func(*ast.Ident, string)) {
	info := pkg.TypesInfo
	for id, obj := range info.Defs {
		if ignore[id.Name] {
			objsToUpdate[obj] = false
		}
		if newname, ok := names[id.Name]; ok {
			objsToUpdate[obj] = false
			if !ignore[id.Name] {
				// Exclude types to be removed.
				renameID(id, newname)
			}
		}
	}
	for id, obj := range info.Uses {
		if ignore[id.Name] {
			objsToUpdate[obj] = false
		}
		if newname, ok := names[id.Name]; ok {
			objsToUpdate[obj] = false
			renameID(id, newname)
		}
	}
}

// prefixPkg prefixes all global identifiers (types, variables, functions).
func prefixPkg(pkg *packages.Package, prefix string,
	objsToUpdate map[types.Object]bool, renameID func(*ast.Ident, string)) {
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
			renameID(id, prefix+obj.Name())
		}
	}
	for id, obj := range info.Uses {
		if objsToUpdate[obj] {
			renameID(id, prefix+obj.Name())
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

// renamer is to be used when identifiers are renamed/prefixed etc in an ast tree.
// It returns the function to be used for renaming and the handler to be called once
// the use of the ast tree is complete, to restore the identifiers original name.
func renamer() (rename func(*ast.Ident, string), done func()) {
	m := map[*ast.Ident]string{}
	return func(id *ast.Ident, name string) {
			m[id] = id.Name
			id.Name = name
		}, func() {
			for id, name := range m {
				id.Name = name
			}
		}
}

func printNode(out io.Writer, fset *token.FileSet, node interface{}) error {
	err := format.Node(out, fset, &printer.CommentedNode{Node: node})
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(out, "\n")
	return err
}
