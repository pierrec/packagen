// DO NOT EDIT Code automatically generated.
// Generated by: go run github.com/pierrec/packagen/cmd/packagen -v bundle -nogen -o slice_int64.go -prefix Int64 -mvtype Item=Int64 -rmtype Item slice.go
package slice

type Int64Slice []Int64 // this type and its methods will be implemented and renamed with the relevant chosen type

func (s Int64Slice) Min() Int64 {
	var m Int64
	for _, v := range s {
		if v := Int64(v); v.Less(m) {
			m = v
		}
	}
	return m
}
func (s Int64Slice) Max() Int64 {
	var m Int64
	for _, v := range s {
		if v := Int64(v); m.Less(v) {
			m = v
		}
	}
	return m
}
