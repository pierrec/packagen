//go:generate go run github.com/pierrec/packagen/cmd/packagen -v bundle -nogen -o slice_int64.go -prefix Int64 -mvtype Item=Int64 -rmtype Item slice.go
package slice

type Int64 int64

func (n Int64) Less(p Int64) bool {
	return n < p
}
