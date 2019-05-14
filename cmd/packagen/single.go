package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/pierrec/cmdflag"
	"github.com/pierrec/packagen"
)

const (
	listSeparator = ','
	typeSeparator = '='
)

func init() {
	cli.MustAdd(cmdflag.Application{
		Name:  "single",
		Descr: "create a single file from a package",
		Args:  "<list of patterns matching the packages to be processed>",
		Init: func(set *flag.FlagSet) cmdflag.Initializer {
			var o packagen.SingleOption
			if verbose {
				o.Log = newLogger()
			}

			set.StringVar(&o.NewPkgName, "newpkg", "",
				"new package name (default=current working dir package)")
			set.StringVar(&o.Prefix, "prefix", "",
				"prefix used to rename declarations (default=packageName_)")

			var mvtype string
			set.StringVar(&mvtype, "mvtype", "",
				fmt.Sprintf("list of named types to be renamed: old%cnew[%c ...]",
					typeSeparator, listSeparator))

			var rmtype string
			set.StringVar(&rmtype, "rmtype", "",
				fmt.Sprintf("list of named types to be removed: typename[%c ...]",
					listSeparator))

			var upconst string
			set.StringVar(&upconst, "const", "",
				fmt.Sprintf("list of interger consts to be updated: constname%cinteger[%c ...]",
					typeSeparator, listSeparator))

			var outfile string
			set.StringVar(&outfile, "o", "", "write output to `file` (default=standard output)")

			return func(args ...string) (err error) {
				o.Patterns = args
				o.RmTypes = buildRemove(rmtype)
				o.Types, err = toMapString(mvtype)
				if err != nil {
					return err
				}
				o.Const, err = toMapInt(upconst)
				if err != nil {
					return err
				}
				var out io.Writer
				if outfile == "" {
					// Buffer standard output.
					buf := bufio.NewWriter(os.Stdout)
					defer buf.Flush()
					out = buf
				} else {
					f, err := os.Create(outfile)
					if err != nil {
						return err
					}
					defer f.Close()
					out = f
				}
				// File header.
				_, err = fmt.Fprintf(out, "// DO NOT EDIT Code automatically generated.\n")
				if err != nil {
					return err
				}
				_, err = fmt.Fprintf(out, "//go:generate go run github.com/pierrec/packagen/cmd/packagen %s\n",
					strings.Join(os.Args[1:], " "))
				if err != nil {
					return err
				}

				return packagen.Single(out, o)
			}
		},
	})
}

func buildRemove(src string) map[string]struct{} {
	m := map[string]struct{}{}
	if src != "" {
		for _, s := range strings.Split(src, string(listSeparator)) {
			m[s] = struct{}{}
		}
	}
	return m
}

func toMapString(src string) (map[string]string, error) {
	m := map[string]string{}
	if src == "" {
		return m, nil
	}

	for _, kv := range strings.Split(src, string(listSeparator)) {
		i := strings.IndexByte(kv, typeSeparator)
		if i < 0 {
			return nil, fmt.Errorf("missing separator %c in %s", typeSeparator, kv)
		}
		m[kv[:i]] = kv[i+1:]
	}

	return m, nil
}

func toMapInt(src string) (map[string]int, error) {
	m := map[string]int{}
	if src == "" {
		return m, nil
	}

	for _, kv := range strings.Split(src, string(listSeparator)) {
		i := strings.IndexByte(kv, typeSeparator)
		if i < 0 {
			return nil, fmt.Errorf("missing separator %c in %s", typeSeparator, kv)
		}
		n, err := strconv.Atoi(kv[i+1:])
		if err != nil {
			return nil, err
		}
		m[kv[:i]] = n
	}

	return m, nil
}
