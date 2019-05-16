# packagen : Generate Go code from packages or files

## Overview [![GoDoc](https://godoc.org/github.com/pierrec/packagen?status.svg)](https://godoc.org/github.com/pierrec/packagen) [![Go Report Card](https://goreportcard.com/badge/github.com/pierrec/packagen)](https://goreportcard.com/report/github.com/pierrec/packagen)

Before Go adopts generics in one form or another (if it ever does!), duplicating code for different types is the 
current way of dealing with  it.
The goal of this package is to be able to generate that code from a well written and tested package, and incidentally 
from any go file.

## Install

```
go get github.com/pierrec/packagen/cmd/packagen
```

## Usage

The main idea is to define a pivot type, eventually with its own set of methods, that works for the implemented algorithm.
That type is then replaced when generating the code with a custom one that works in the same way, implementing the same 
methods if necessary.

The generated code contains the same types, functions etc than the source, but prefixed so that they do not collide
with the code in the same package.

## Example

Source file:
```
type Numbers []int

func (s Numbers) Len() int { return len(s) }
func (s Numbers) Less(i, j int) bool { return s[i] < s[j] }
func (s Numbers) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func Sort(s Numbers) {
    // ... sorting algorithm
}
```

Existing `int32s.go` file:
```
type Int32s []int32

// Int32s implements the methods required by the sort algorithm.
func (s Int32s) Len() int { return len(s) }
func (s Int32s) Less(i, j int) bool { return s[i] < s[j] }
func (s Int32s) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
```

File `int32s_gen.go` generated with `packagen single -o int32s_gen.go -prefix Int32 -mvtype Numbers=Int32s -rmtype Numbers`:
```
// All the code from the source is preserved and prefixed, except for the pivot type and its methods.
func Int32Sort(s Int32s) {
    // ... sorting algorithm
}
```

## Contributing

Contributions welcome via pull requests. Please provide tests.
