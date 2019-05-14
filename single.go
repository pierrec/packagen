package packagen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

// SingleOption defines the options for the Single processor.
type SingleOption struct {
	Patterns   []string            // Packages to be processed
	NewPkgName string              // Name of the resulting package (default=current working dir package)
	Prefix     string              // Prefix for the global identifiers (default=packageName_)
	Types      map[string]string   // Map the names of the types to be renamed to their new one
	Const      map[string]int      // Values for const to be updated
	RmType     map[string]struct{} // Named types to be removed
}

// newpkgname returns the set value or a default one.
func (o *SingleOption) newpkgname() (string, error) {
	if o.NewPkgName != "" {
		return o.NewPkgName, nil
	}
	// Use the current working directory package name.
	pkgs, err := packages.Load(&packages.Config{Mode: packages.LoadFiles}, ".")
	if len(pkgs) > 0 {
		// Be optimistic: even if the local package has errors, return its name.
		if p := pkgs[0].Name; p != "" {
			return p, nil
		}
	}
	if err == nil {
		err = fmt.Errorf("cannot define new package name")
	}
	return "", err
}

// prefix returns the set value or a default one.
func (o *SingleOption) prefix(pkg *packages.Package) string {
	if o.Prefix != "" {
		return o.Prefix
	}
	// Use the source package name as the prefix.
	return pkg.Name + "_"
}

// Single packs the package identified by o.PkgName into a single file and writes it to the given io.Writer.
func Single(out io.Writer, o SingleOption) error {
	pkgs, err := packages.Load(&packages.Config{Mode: packages.LoadSyntax}, o.Patterns...)
	if err != nil {
		return err
	}
	if packages.PrintErrors(pkgs) > 0 {
		return fmt.Errorf("too many errors while loading package %s", o.Patterns)
	}

	// Rename types in all packages.
	objsToUpdate := map[types.Object]bool{}
	for _, pkg := range pkgs {
		renamePkg(pkg, o.Types, objsToUpdate)
	}

	// Prefix global declarations in all packages.
	for _, pkg := range pkgs {
		prefixPkg(pkg, o.prefix(pkg), objsToUpdate)
	}

	// Build the single file package.
	var buf bytes.Buffer
	newName, err := o.newpkgname()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(&buf, "package %s\n\n", newName)
	if err != nil {
		return err
	}
	//TODO test package is missing?
	for _, pkg := range pkgs {
		for _, f := range pkg.Syntax {
		next:
			for _, decl := range f.Decls {
				if decl, ok := decl.(*ast.GenDecl); ok {
					switch decl.Tok {
					case token.IMPORT:
						// Skip imports.
						continue
					case token.CONST:
						for _, spec := range decl.Specs {
							v, ok := spec.(*ast.ValueSpec)
							if !ok {
								continue
							}
							for i, id := range v.Names {
								lit, ok := v.Values[i].(*ast.BasicLit)
								if !ok {
									continue
								}
								// Check without the added prefix...
								name := strings.TrimPrefix(id.Name, o.prefix(pkg))
								if n, ok := o.Const[name]; ok {
									lit.Value = strconv.Itoa(n)
								}
							}
						}
					case token.TYPE:
						for _, spec := range decl.Specs {
							t, ok := spec.(*ast.TypeSpec)
							if !ok {
								continue
							}
							if _, ok := o.RmType[t.Name.Name]; ok {
								// Type to be removed.
								continue next
							}
						}
					}
				}
				err := format.Node(&buf, pkg.Fset, &printer.CommentedNode{Node: decl})
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(&buf, "\n")
				if err != nil {
					return err
				}
			}
		}
	}

	// Resolved imports and format the resulting code.
	code, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, bytes.NewReader(code))

	return err
}
