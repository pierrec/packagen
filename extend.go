package packagen

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"log"

	"golang.org/x/tools/go/packages"
)

// ExtendOption defines the options in use when extending a type.
type ExtendOption struct {
	Log          *log.Logger
	SrcPkg       string            // Package of the source type
	Src          string            // Name of the struct type to be used as source
	DstPkg       string            // Package of the destination type
	Dst          string            // Name of the struct type to be extended
	Fields       map[string]string // Map field name to the new type
	FieldPrefix  string            // Prefix to be used for the added fields
	MethodPrefix string            // Prefix to be used for method names
}

// ExtendStruct adds fields and methods from one struct to another.
// It returns the name of the file where the destination struct is located
// and writes the new destination content to out and its new methods (if any) to methods.
func ExtendStruct(out, methods io.Writer, o ExtendOption) (string, error) {
	dstPkg, dstType, dstStruct, err := lookupStruct(o.DstPkg, o.Dst)
	if err != nil {
		return "", err
	}
	srcPkg, _, srcStruct, err := lookupStruct(o.SrcPkg, o.Src)
	if err != nil {
		return "", err
	}

	// Restore modified identifiers names.
	renameID, renameDone := renamer()
	defer renameDone()

	// Find the destination file.
	var dstFile *ast.File
fileLoop:
	for _, file := range dstPkg.Syntax {
		for _, decl := range file.Decls {
			if decl, ok := decl.(*ast.GenDecl); ok && decl.Tok == token.TYPE {
				for _, spec := range decl.Specs {
					if t, ok := spec.(*ast.TypeSpec); ok && t == dstType {
						dstFile = file
						break fileLoop
					}
				}
			}
		}
	}

	// Extend the destination type.
	for _, decl := range dstFile.Decls {
		decl, ok := decl.(*ast.GenDecl)
		if !ok || decl.Tok != token.TYPE {
			continue
		}
		for _, spec := range decl.Specs {
			t, ok := spec.(*ast.TypeSpec)
			if !ok || t != dstType {
				continue
			}
			// Extend the type with new fields.
			for _, field := range srcStruct.Fields.List {
				if field.Names == nil {
					// Embedded type, ignore.
					continue
				}
				name := field.Names[0].Name
				field.Names[0].Name = o.FieldPrefix + name
				newtype, ok := o.Fields[name]
				if !ok {
					// Do not change the type if none defined for this field.
					continue
				}
				switch e := field.Type.(type) {
				case *ast.Ident:
					renameID(e, newtype)
				case *ast.SliceExpr:
					if id, ok := e.X.(*ast.Ident); ok {
						renameID(id, newtype)
					}
				case *ast.ArrayType:
					if id, ok := e.Elt.(*ast.Ident); ok {
						renameID(id, newtype)
					}
				}
			}
			dstStruct.Fields.List = append(dstStruct.Fields.List, srcStruct.Fields.List...)
		}
	}

	// Write the new file content.
	if err := printNode(out, dstPkg.Fset, dstFile); err != nil {
		return "", err
	}

	// Write new methods.
	if _, err = fmt.Fprintf(methods, "package %s\n\n", dstPkg.Name); err != nil {
		return "", err
	}
	for _, file := range srcPkg.Syntax {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil {
				continue
			}
			// Got a method.
			var id *ast.Ident
			switch t := fn.Recv.List[0].Type.(type) {
			case *ast.StarExpr:
				id = t.X.(*ast.Ident)
			case *ast.Ident:
				id = t
			}
			if id == nil || id.Name != o.Src {
				// Invalid receiver?? or not for the selected target type.
				continue
			}
			renameID(fn.Name, o.MethodPrefix+fn.Name.Name)
			if err := printNode(methods, srcPkg.Fset, decl); err != nil {
				return "", err
			}
		}
	}

	return dstFile.Name.Name, nil
}

func lookupStruct(pname, tname string) (p *packages.Package, t *ast.TypeSpec, s *ast.StructType, err error) {
	pkgs, err := loadPkg(pname)
	if err != nil {
		return
	}
	p = pkgs[0]

	// Look for the target type.
	t = lookupType(p, tname)
	if t == nil {
		err = fmt.Errorf("target type %q not found in package %s", tname, pname)
		return
	}
	s, ok := t.Type.(*ast.StructType)
	if !ok {
		err = fmt.Errorf("target type %q is not a struct", tname)
		return
	}
	return
}

// lookupType returns the spec for the type with name.
func lookupType(pkg *packages.Package, name string) *ast.TypeSpec {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			d, ok := decl.(*ast.GenDecl)
			if !ok || d.Tok != token.TYPE {
				continue
			}
			for _, spec := range d.Specs {
				if t, ok := spec.(*ast.TypeSpec); ok && t.Name.Name == name {
					return t
				}
			}
		}
	}
	return nil
}
