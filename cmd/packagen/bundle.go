package main

import (
	"flag"
	"fmt"

	"github.com/pierrec/cmdflag"
	"github.com/pierrec/packagen"
)

func init() {
	cli.MustAdd(cmdflag.Application{
		Name:  "bundle",
		Descr: "bundle all files in a package",
		Args:  "package to be processed",
		Err:   flag.ExitOnError,
		Init: func(set *flag.FlagSet) cmdflag.Handler {
			o := packagen.BundleOption{Log: newLogger()}
			var nogen bool
			set.BoolVar(&nogen, "nogen", false, "do not add the generate directive")

			set.StringVar(&o.NewPkg, "newpkg", "",
				"new package name (default=current working dir package)")
			set.StringVar(&o.Prefix, "prefix", "",
				"prefix used to rename declarations (default=packageName_)")

			var mvtype string
			set.StringVar(&mvtype, "mvtype", "",
				fmt.Sprintf("list of named types to be renamed: old%cnew[%c ...]", typeSep, listSep))

			var rmtype string
			set.StringVar(&rmtype, "rmtype", "",
				fmt.Sprintf("list of named types to be removed: typename[%c ...]", listSep))

			var upconst string
			set.StringVar(&upconst, "const", "",
				fmt.Sprintf("list of integer constants to be updated: constname%cinteger[%c ...]", typeSep, listSep))

			var rmconst string
			set.StringVar(&rmconst, "rmconst", "",
				fmt.Sprintf("list of constants to be discarded: constname[%c ...]", typeSep))

			var outfile string
			set.StringVar(&outfile, "o", "", "write output to `file` (default=standard output)")

			return func(args ...string) (_ int, err error) {
				switch len(args) {
				case 1:
				case 0:
					err = errMissingPkg
					return
				default:
					err = errTooManyPkg
					return
				}
				o.Types, err = toMapString(mvtype)
				if err != nil {
					return
				}
				o.Const, err = toMapInt(upconst)
				if err != nil {
					return
				}
				o.Pkg = args[0]
				o.RmTypes = toMapBool(rmtype)
				o.RmConst = toMapBool(rmconst)

				out, err := initOutput(outfile, true, nogen)
				if err != nil {
					return
				}
				err = packagen.Bundle(out, o)
				if err != nil {
					return
				}
				return len(args), out.Close()
			}
		},
	})
}
