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

The available command line options are:
  - single <list of patterns matching the packages to be processed>
    - -const **list of integer constants to be updated** (constname=integer[, ...])
    - -mvtype **list of named types to be renamed** (old=new[, ...])
    - -newpkg **new package name** (default=current working dir package)
    - -nogen - do not add the generate directive
    - -o **file** - write output to file (default=standard output)
    - -prefix **prefix used to rename declarations** (default=packageName_)
    - -rmtype **list of named types to be removed** (typename[, ...])


## Example

Given a sorting algorithm implemented in the package `domain/user/sort`, generate the code for another integer type 
with the following command:

`packagen single -o int32s_gen.go -prefix Int32 -mvtype Numbers=Int32s -rmtype Numbers domain/user/sort`

Source package:
```
package sort

type Numbers []int

func (s Numbers) Len() int { return len(s) }
func (s Numbers) Less(i, j int) bool { return s[i] < s[j] }
func (s Numbers) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func Sort(s Numbers) {
    // ... sorting algorithm
}
```

In the new package `domain/user/myapp`, there is an existing `int32s.go` file:
```
package myapp

type Int32s []int32

// Int32s implements the methods required by the sort algorithm.
func (s Int32s) Len() int { return len(s) }
func (s Int32s) Less(i, j int) bool { return s[i] < s[j] }
func (s Int32s) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
```

The generated file `int32s_gen.go` contains:
```
//go:generate go run github/pierrec/packagen/cmd/packagen single -o int32s_gen.go -prefix Int32 -mvtype Numbers=Int32s -rmtype Numbers domain/user/sort

package myapp

// Note that the package name has been set automatically.
// Also, the go:generate directive is added by default, but this can be disabled.

// All the code from the source is preserved and top level declarations are prefixed, except for the pivot type and its 
// methods which have been removed.
func Int32Sort(s Int32s) {
    // ... sorting algorithm
}
```

## Contributing

Contributions welcome via pull requests. Please provide tests.
