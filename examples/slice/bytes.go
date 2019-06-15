//go:generate go run github.com/pierrec/packagen/cmd/packagen -v bundle -nogen -o slice_bytes.go -prefix Bytes -mvtype Item=Bytes -rmtype Item slice.go
package slice

import "bytes"

type Bytes []byte

func (n Bytes) Less(p Bytes) bool {
	return bytes.Compare(n, p) < 0
}
