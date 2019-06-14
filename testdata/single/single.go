package single

type (
	A string
	L = A
)

const (
	V = iota
	V1
	V2
)

const (
	C  = L("abc")
	CA = A("xyz")
)

type S struct {
	V A
}

func (s S) String() string {
	return string(s.V)
}

func (s *S) GoString() string {
	return string(s.V)
}

type AS struct {
	V S
}

func (as AS) String() string {
	return as.V.String()
}

func (as *AS) GoString() string {
	return as.V.String()
}
