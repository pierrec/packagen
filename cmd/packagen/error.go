package main

type errorString string

func (e errorString) Error() string {
	return string(e)
}

const (
	errMissingPkg   errorString = "missing package name"
	errTooManyPkg   errorString = "too many packages"
	errMustGenerate errorString = "must be called by go generate"
)
