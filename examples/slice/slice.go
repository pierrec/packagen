// This file is the reference implementation of the Slice type.
// It is implemented for type Number, which can be substituted with any Go builtin number type:
// int, int64, int32, int16, int8 or their float/uint counterparts.
package slice

type Number int // this type will get removed for generated code

type Slice []Number // this type and its methods will be implemented and renamed with the relevant chosen type

func (s Slice) Min() Number {
	var m Number
	for _, v := range s {
		if v := Number(v); v < m {
			m = v
		}
	}
	return m
}

func (s Slice) Max() Number {
	var m Number
	for _, v := range s {
		if v := Number(v); v > m {
			m = v
		}
	}
	return m
}
