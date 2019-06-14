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

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

// BundleOption defines the options for the Bundle processor.
type BundleOption struct {
	Log     *log.Logger
	Pkg     string            // Package to be processed
	NewPkg  string            // Name of the resulting package (default=current working dir package)
	Prefix  string            // Prefix for the global identifiers (default=packageName_)
	Types   map[string]string // Map the names of the types to be renamed to their new one
	RmTypes map[string]bool   // Named types to be removed
	Const   map[string]int    // Values for const to be updated
	RmConst map[string]bool   // Constants to be removed
}

// newpkgname returns the set value or a default one.
func (o *BundleOption) newpkgname() (string, error) {
	if o.NewPkg != "" {
		return o.NewPkg, nil
	}
	return localPkgName()
}

// prefix returns the set value or a default one.
func (o *BundleOption) prefix(pkg *packages.Package) string {
	if o.Prefix != "" {
		return o.Prefix
	}
	// Use the source package name as the prefix.
	return pkg.Name + "_"
}

// Bundle packs the package identified by o.PkgName into a single file and writes it to the given io.Writer.
func Bundle(out io.Writer, o BundleOption) error {
	if o.Log != nil {
		o.Log.Printf("Options: %#v\n", o)
		o.Log.Printf("Loading packages with %v\n", o.Pkg)
	}
	pkgs, err := loadPkg(o.Pkg)
	if err != nil {
		return err
	}
	if packages.PrintErrors(pkgs) > 0 {
		return fmt.Errorf("too many errors while loading package %s", o.Pkg)
	}
	if o.Log != nil {
		o.Log.Printf("Found %d packages: %v\n", len(pkgs), pkgs)
	}

	// Build the list of identifiers NOT to be renamed:
	// - renamed types
	// - removed constants
	// - removed types
	ignore := make(map[string]bool)
	for src, tgt := range o.Types {
		if o.RmTypes[src] {
			// Make sure that renamed types that need to be removed are also in the rm list.
			o.RmTypes[tgt] = true
		}
	}
	for src := range o.RmConst {
		ignore[src] = true
	}
	for src := range o.RmTypes {
		ignore[src] = true
	}
	if o.Log != nil {
		o.Log.Printf("No prefix: %v", keysOf(ignore))
	}
	// Rename types in all packages.
	renameID, renameDone := renamer()
	defer renameDone()

	objsToUpdate := map[types.Object]bool{}
	for _, pkg := range pkgs {
		if o.Log != nil {
			o.Log.Printf("Renaming types in %v\n", pkg)
		}
		renamePkg(pkg, o.Types, ignore, objsToUpdate, renameID)
	}

	// Prefix global declarations in all packages.
	for _, pkg := range pkgs {
		if o.Log != nil {
			o.Log.Printf("Prefixing types in %v\n", pkg)
		}
		prefixPkg(pkg, o.prefix(pkg), objsToUpdate, renameID)
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
						if len(o.Const) == 0 && len(o.RmConst) == 0 && len(o.RmTypes) == 0 {
							break
						}
						for _, spec := range decl.Specs {
							v, ok := spec.(*ast.ValueSpec)
							if !ok {
								continue
							}
							if ident, ok := v.Type.(*ast.Ident); ok {
								// Typed constant: remove if its type is to be removed.
								if name := ident.Name; o.RmTypes[name] {
									if o.Log != nil {
										o.Log.Printf("const of type %s discarded", name)
									}
									continue next
								}
							}
							if len(v.Values) == 0 {
								// initial values; or nil
								continue
							}
							// Do not print out the constant if it is defined standalone.
							if len(v.Names) == 1 {
								if name := v.Names[0].Name; o.RmConst[name] {
									// Constant to be completely removed.
									if o.Log != nil {
										o.Log.Printf("const %s discarded", name)
									}
									continue next
								}
								continue
							}
							// If more than one constant, ignore its line (might be part of iota?).
							for i, id := range v.Names {
								name := id.Name
								if o.RmConst[name] {
									// Constant to be ignored.
									id.Name = "_"
									if o.Log != nil {
										o.Log.Printf("const %s ignored", name)
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
							if name := t.Name.Name; o.RmTypes[name] {
								// Type to be removed.
								if o.Log != nil {
									o.Log.Printf("type %s discarded", name)
								}
								continue next
							}
						}
					}
				case *ast.FuncDecl:
					if len(o.RmTypes) == 0 || decl.Recv == nil {
						break
					}
					var id *ast.Ident
					switch t := decl.Recv.List[0].Type.(type) {
					case *ast.StarExpr:
						id = t.X.(*ast.Ident)
					case *ast.Ident:
						id = t
					}
					if id == nil {
						break
					}
					if name := id.Name; o.RmTypes[name] {
						// Type to be removed.
						if o.Log != nil {
							o.Log.Printf("method for type %s discarded", name)
						}
						continue next
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

	// Resolve imports and format the resulting code.
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
