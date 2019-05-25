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
	"log"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

// SingleOption defines the options for the Single processor.
type SingleOption struct {
	Log        *log.Logger
	Patterns   []string            // Packages to be processed
	NewPkgName string              // Name of the resulting package (default=current working dir package)
	Prefix     string              // Prefix for the global identifiers (default=packageName_)
	Types      map[string]string   // Map the names of the types to be renamed to their new one
	RmTypes    map[string]struct{} // Named types to be removed
	Const      map[string]int      // Values for const to be updated
	RmConst    map[string]struct{} // Constants to be removed
}

// newpkgname returns the set value or a default one.
func (o *SingleOption) newpkgname() (string, error) {
	if o.NewPkgName != "" {
		return o.NewPkgName, nil
	}
	// Use the current working directory package name.
	if o.Log != nil {
		o.Log.Println("Finding local package name")
	}
	pkgs, err := packages.Load(&packages.Config{Mode: packages.LoadFiles}, ".")
	if len(pkgs) > 0 {
		// Be optimistic: even if the local package has errors, return its name.
		if p := pkgs[0].Name; p != "" {
			if o.Log != nil {
				o.Log.Printf("Local package name is %s\n", p)
			}
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
	if o.Log != nil {
		o.Log.Printf("Options: %#v\n", o)
		o.Log.Printf("Loading packages with %v\n", o.Patterns)
	}
	pkgs, err := packages.Load(&packages.Config{Mode: packages.LoadSyntax}, o.Patterns...)
	if err != nil {
		return err
	}
	if packages.PrintErrors(pkgs) > 0 {
		return fmt.Errorf("too many errors while loading package %s", o.Patterns)
	}
	if o.Log != nil {
		o.Log.Printf("Found %d packages: %v\n", len(pkgs), pkgs)
	}

	// Rename types in all packages.
	objsToUpdate := map[types.Object]bool{}
	for _, pkg := range pkgs {
		if o.Log != nil {
			o.Log.Printf("Renaming types in %v\n", pkg)
		}
		renamePkg(pkg, o.Types, o.RmTypes, objsToUpdate)
	}

	// Prefix global declarations in all packages.
	for _, pkg := range pkgs {
		if o.Log != nil {
			o.Log.Printf("Prefixing types in %v\n", pkg)
		}
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
		if o.Log != nil {
			o.Log.Printf("Writing package %v\n", pkg)
		}
		for _, f := range pkg.Syntax {
		next:
			for _, decl := range f.Decls {
				switch decl := decl.(type) {
				case *ast.GenDecl:
					switch decl.Tok {
					case token.IMPORT:
						// Skip imports.
						continue
					case token.CONST:
						// Constants to be updated or removed.
						if len(o.Const) == 0 && len(o.RmConst) == 0 {
							break
						}
						for _, spec := range decl.Specs {
							v, ok := spec.(*ast.ValueSpec)
							if !ok {
								continue
							}
							if len(v.Values) == 0 {
								// initial values; or nil
								continue
							}
							for i, id := range v.Names {
								// Check without the added prefix...
								name := strings.TrimPrefix(id.Name, o.prefix(pkg))

								if _, ok := o.RmConst[name]; ok {
									// Constant to be removed.
									id.Name = "_"
									if o.Log != nil {
										o.Log.Printf("const %s discarded", name)
									}
									continue
								}
								lit, ok := v.Values[i].(*ast.BasicLit)
								if !ok {
									continue
								}
								if n, ok := o.Const[name]; ok {
									// Update the constant value.
									val := lit.Value
									lit.Value = strconv.Itoa(n)
									if o.Log != nil {
										o.Log.Printf("const %s value updated from %s to %s", name, val, lit.Value)
									}
								}
							}
						}
					case token.TYPE:
						if len(o.RmTypes) == 0 {
							break
						}
						for _, spec := range decl.Specs {
							t, ok := spec.(*ast.TypeSpec)
							if !ok {
								continue
							}
							if _, ok := o.RmTypes[t.Name.Name]; ok {
								// Type to be removed.
								if o.Log != nil {
									o.Log.Printf("type %s discarded", t.Name.Name)
								}
								continue next
							}
						}
					}
				case *ast.FuncDecl:
					if len(o.RmTypes) == 0 || decl.Recv == nil {
						break
					}
					if t, ok := decl.Recv.List[0].Type.(*ast.Ident); ok {
						if _, ok := o.RmTypes[t.Name]; ok {
							// Type to be removed.
							if o.Log != nil {
								o.Log.Printf("method for type %s discarded", t.Name)
							}
							continue next
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
	if o.Log != nil {
		o.Log.Printf("Resolving imports\n")
	}
	code, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		// Output without imports.
		_, _ = io.Copy(out, &buf)
		return err
	}
	_, err = io.Copy(out, bytes.NewReader(code))

	return err
}
