// DO NOT EDIT Code automatically generated.
//go:generate go run github.com/pierrec/packagen/cmd/packagen -v single -o slice_int64.go -prefix Int64 -mvtype Number=int64 -rmtype Number slice.go
package slice

type Int64Slice []int64 // this type and its methods will be implemented and renamed with the relevant chosen type

func (s Int64Slice) Min() int64 {
	var m int64
	for _, v := range s {
		if v := int64(v); v < m {
			m = v
		}
	}
	return m
}
func (s Int64Slice) Max() int64 {
	var m int64
	for _, v := range s {
		if v := int64(v); v > m {
			m = v
		}
	}
	return m
}
