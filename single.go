package packagen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"go/types"
	"golang.org/x/tools/imports"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

// SingleOption defines the options for the Single processor.
type SingleOption struct {
	PkgName    string              // Package to be processed
	NewPkgName string              // Name of the resulting package
	Types      map[string]string   // Map the names of the types to be renamed to their new one
	Const      map[string]int      // Values for const to be updated
	RmType     map[string]struct{} // Named types to be removed
	Prefix     string              // Prefix for the global identifiers
}

// Single packs the package identified by o.PkgName into a single file and writes it to the given io.Writer.
func Single(out io.Writer, o SingleOption) error {
	if o.NewPkgName == "" {
		o.NewPkgName = filepath.Base(o.PkgName)
	}

	cfg := &packages.Config{Mode: packages.LoadSyntax}
	pkgs, err := packages.Load(cfg, o.PkgName)
	if err != nil {
		return err
	}
	if packages.PrintErrors(pkgs) > 0 {
		return fmt.Errorf("too many errors while loading package %s", o.PkgName)
	}

	// Rename types in all packages.
	objsToUpdate := map[types.Object]bool{}
	for _, pkg := range pkgs {
		renamePkg(pkg, o.Types, objsToUpdate)
	}

	// Prefix global declarations in all packages.
	for _, pkg := range pkgs {
		prefixPkg(pkg, o.Prefix, objsToUpdate)
	}

	// Build the single file package.
	var buf bytes.Buffer
	_, _ = fmt.Fprintf(&buf, "package %s\n\n", o.NewPkgName)

	//TODO test package is missing?
	for _, pkg := range pkgs {
		for _, f := range pkg.Syntax {
		next:
			for _, decl := range f.Decls {
				decl, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}
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
							name := strings.TrimPrefix(id.Name, o.Prefix)
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
				err := format.Node(&buf, pkg.Fset, &printer.CommentedNode{Node: decl})
				if err != nil {
					return err
				}
				_, _ = fmt.Fprint(&buf, "\n")
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
