package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/pierrec/cmdflag"
	"github.com/pierrec/packagen"
)

func init() {
	cli.MustAdd(cmdflag.Application{
		Name:  "extend",
		Descr: "extend a type with fields and methods from another one",
		Args:  "",
		Err:   flag.ExitOnError,
		Init: func(set *flag.FlagSet) cmdflag.Handler {
			o := packagen.ExtendOption{Log: newLogger()}

			set.StringVar(&o.SrcPkg, "pkg", "", "source package name")
			set.StringVar(&o.Src, "src", "", "source type name")
			set.StringVar(&o.Dst, "tgt", "", "extended type name")
			set.StringVar(&o.FieldPrefix, "fprefix", "", "field prefix")
			set.StringVar(&o.MethodPrefix, "mprefix", "", "method prefix")

			var fields string
			set.StringVar(&fields, "fields", "",
				fmt.Sprintf("list of field names to their type: name%ctype[%c ...]",
					typeSep, listSep))

			return func(args ...string) (_ int, err error) {
				o.Fields, err = toMapString(fields)
				if err != nil {
					return
				}
				var buf, methods bytes.Buffer
				fname, err := packagen.ExtendStruct(&buf, &methods, o)
				if err != nil {
					return
				}
				// Write the updated type.
				if err = extendWriteFile(fname, &buf); err != nil {
					return
				}
				// Write the type methods.
				if methods.Len() > 0 {
					mname := fmt.Sprintf("%s_gen.go", strings.TrimSuffix(fname, ".go"))
					err = extendWriteFile(mname, &methods)
				}
				return
			}
		},
	})
}

func extendWriteFile(name string, buf *bytes.Buffer) error {
	out, err := initOutput(name, false, false)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, buf); err != nil {
		return err
	}
	return out.Close()
}
