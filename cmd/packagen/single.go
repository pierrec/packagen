package main

import (
	"flag"
	"fmt"
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
		Args:  "",
		Init: func(set *flag.FlagSet) cmdflag.Initializer {
			var o packagen.SingleOption
			set.StringVar(&o.PkgName, "pkg", "", "package name to be packed")
			set.StringVar(&o.NewPkgName, "newpkg", "", "new package name")
			var mvtype string
			set.StringVar(&mvtype, "mvtype", "",
				fmt.Sprintf("list of named mvtype to be renamed: old%cnew[%c ...]",
					typeSeparator, listSeparator))
			var rmtype string
			set.StringVar(&rmtype, "rmtype", "",
				fmt.Sprintf("list of named mvtype to be removed: typename[%c ...]",
					listSeparator))
			set.StringVar(&o.Prefix, "prefix", "", "prefix used to rename declarations")
			var upconst string
			set.StringVar(&upconst, "const", "",
				fmt.Sprintf("list of interger consts to be updated: constname%cinteger[%c ...]",
					typeSeparator, listSeparator))

			return func(args ...string) (err error) {
				o.RmType = buildRemove(rmtype)
				o.Types, err = toMapString(mvtype)
				if err != nil {
					return err
				}
				o.Const, err = toMapInt(upconst)
				if err != nil {
					return err
				}
				return packagen.Single(os.Stdout, o)
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
