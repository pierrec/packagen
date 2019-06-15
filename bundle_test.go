package packagen

import (
	"bytes"
	"flag"
	"io/ioutil"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
)

var update = flag.Bool("update", false, "update .golden files")

func TestBundle(t *testing.T) {
	for _, tc := range []BundleOption{
		{
			Pkg:    "./testdata/bundle",
			NewPkg: "mvtypes",
			Prefix: "prefix",
			Types:  map[string]string{"S": "X"},
		},
		{
			Pkg:     "./testdata/bundle",
			NewPkg:  "rmtypes",
			Prefix:  "prefix",
			RmTypes: map[string]bool{"S": true},
		},
		{
			Pkg:     "./testdata/bundle",
			NewPkg:  "rmconst",
			Prefix:  "prefix",
			RmConst: map[string]bool{"V": true},
		},
		// Removing a type and its methods but avoid prefixing its references.
		{
			Pkg:     "./testdata/bundle",
			NewPkg:  "mvrmtype",
			Prefix:  "prefix",
			Types:   map[string]string{"A": "A"},
			RmTypes: map[string]bool{"A": true},
		},
	} {
		t.Run(tc.Pkg, func(t *testing.T) {
			c := qt.New(t)

			buf := new(bytes.Buffer)
			err := Bundle(buf, tc)
			c.Assert(err, qt.IsNil)

			fname := filepath.Join("testdata", "bundle_"+tc.NewPkg+".golden")
			if *update {
				t.Log("update golden file")
				if err := ioutil.WriteFile(fname, buf.Bytes(), 0644); err != nil {
					t.Fatalf("failed to update golden file: %s", err)
				}
			}
			result, err := ioutil.ReadFile(fname)
			c.Assert(err, qt.IsNil)
			c.Assert(buf.String(), qt.Equals, string(result))
		})
	}
}
