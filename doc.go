// Package packagen generate Go source code from a source package.
// It is largely inspired by https://github.com/golang/tools/cmd/bundle.
//
// Transformations can be applied to the resulting AST before outputting it.
//
// Limitations:
// - no asm, no cgo, no build tags
package packagen
