// This file is the reference implementation of the Slice type.
// It is implemented for type Item, which can be substituted with any type implementing the Less() method.
package slice

type Item int // this type will get removed for generated code

func (n Item) Less(p Item) bool {
	return n < p
}

type Slice []Item // this type and its methods will be implemented and renamed with the relevant chosen type

func (s Slice) Min() Item {
	var m Item
	for _, v := range s {
		if v := Item(v); v.Less(m) {
			m = v
		}
	}
	return m
}

func (s Slice) Max() Item {
	var m Item
	for _, v := range s {
		if v := Item(v); m.Less(v) {
			m = v
		}
	}
	return m
}
