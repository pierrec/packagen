package packagen

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestExtendStruct(t *testing.T) {
	for _, tc := range []ExtendOption{
		{
			SrcPkg:       "./testdata/extend/src",
			Src:          "Data",
			DstPkg:       "./testdata/extend/dst",
			Dst:          "ExData",
			Fields:       map[string]string{"i": "int32", "is": "uint32"},
			FieldPrefix:  "field_",
			MethodPrefix: "method_",
		},
	} {
		t.Run(tc.SrcPkg, func(t *testing.T) {
			c := qt.New(t)

			var buf, methods bytes.Buffer
			_, err := ExtendStruct(&buf, &methods, tc)
			c.Assert(err, qt.IsNil)

			fname := filepath.Join("testdata", "extend_"+tc.Dst+".golden")
			mname := filepath.Join("testdata", "extend_"+tc.Dst+"_methods.golden")
			if *update {
				t.Log("update golden file")
				if err := ioutil.WriteFile(fname, buf.Bytes(), 0644); err != nil {
					t.Fatalf("failed to update golden file: %s", err)
				}
				if err := ioutil.WriteFile(mname, methods.Bytes(), 0644); err != nil {
					t.Fatalf("failed to update golden file: %s", err)
				}
			}
			result, err := ioutil.ReadFile(fname)
			c.Assert(err, qt.IsNil)
			c.Assert(buf.String(), qt.Equals, string(result))

			result, err = ioutil.ReadFile(mname)
			c.Assert(err, qt.IsNil)
			c.Assert(methods.String(), qt.Equals, string(result))
		})
	}
}
