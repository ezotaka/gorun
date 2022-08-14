// Copyright (c) 2022 Takatomo Ezo
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package ezast

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/ast/astutil"
)

// ast.File with token.FileSet
type AstFile struct {
	*ast.File
	fset *token.FileSet
}

// Convert to minimized source code text.
func (a AstFile) String() string {
	pp := &printer.Config{}
	buf := bytes.NewBufferString("")

	pp.Fprint(buf, a.fset, a.File)
	return buf.String()
}

// Save AST as source code to temp file.
//
// SaveTemp returns filePath, temp file cleaner func , and error.
// Cleaner func is not always nil, so you may call it any time.
func (a AstFile) SaveTemp() (file string, cleaner func(), err error) {
	dir, err := os.MkdirTemp("", "gorun")
	if err != nil {
		cleaner = func() {}
		return
	}
	cleaner = func() { os.RemoveAll(dir) }

	file = filepath.Join(dir, "main.go")
	err = os.WriteFile(file, []byte(a.String()), 0666)
	return
}

// NewAstFile parses the source code of a single Go source file and returns
// the corresponding ast.File node
func NewAstFile(file string) (*AstFile, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.Mode(0))

	if err != nil {
		return nil, err
	}

	return &AstFile{
		fset: fset,
		File: f,
	}, nil
}

// Change package name
func (a *AstFile) ChangePackage(newPkg string) error {
	if newPkg == "" {
		return errors.New("package name must be not empty")
	}
	if a.File.Name.Name != newPkg {
		a.File.Name = &ast.Ident{Name: newPkg}
	}
	return nil
}

// GetFuncDecl returns declaration of func named fn.
//
// If there is no declaration, it returns nil.
func (a *AstFile) GetFuncDecl(fn string) *ast.FuncDecl {
	for _, d := range a.File.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok {
			if fd.Name.Name == fn {
				return fd
			}
		}
	}
	return nil
}

// ContainsFunc returns true if AST contains func fn
func (a *AstFile) ContainsFunc(fn string) bool {
	return a.GetFuncDecl(fn) != nil
}

// Swap func name of fn1 and one of f2.
// * Funcs must not have arguments and return values.
//
// If strict is true, fn1 and fn2 must be present.
// When either is not present, SwapSimpleFuncs returns error.
//
// If strict is false, only present func is renamed, SwapSimpleFuncs returns no errors.
func (a *AstFile) SwapSimpleFuncs(fn1, fn2 string, strict bool) error {
	if strict {
		if !a.ContainsFunc(fn1) || !a.ContainsFunc(fn2) {
			return fmt.Errorf("func '%s' or func '%s' is not found", fn1, fn2)
		}
	}

	hasArgsOrReturns := func(fn string) bool {
		fd := a.GetFuncDecl(fn)
		if fd == nil {
			return false
		}
		return len(fd.Type.Params.List) > 0 || fd.Type.Results != nil
	}

	if hasArgsOrReturns(fn1) || hasArgsOrReturns(fn2) {
		return fmt.Errorf("func '%s' and '%s' must not have arguments and return values", fn1, fn2)
	}

	if fn1 == fn2 {
		return nil
	}

	return processFuncDecl(a, func(cr *astutil.Cursor, d *ast.FuncDecl) bool {
		if d.Name.Name == fn1 {
			d.Name = &ast.Ident{
				Name: fn2,
			}
		} else if d.Name.Name == fn2 {
			d.Name = &ast.Ident{
				Name: fn1,
			}
		}
		return false
	})
}

func processFuncDecl(a *AstFile, fn func(*astutil.Cursor, *ast.FuncDecl) bool) error {
	if a == nil {
		return errors.New("nil astFile")
	}

	astutil.Apply(a.File, func(cr *astutil.Cursor) bool {
		if d, ok := cr.Node().(*ast.FuncDecl); ok {
			return fn(cr, d)
		}
		return true
	}, nil)

	return nil
}

// Assert two ASTs are structurally equivalent
func AssertEqualAstFile(want, got *ast.File) bool {
	if want == nil {
		return got == nil
	}
	if got == nil {
		return want == nil
	}

	w := &AstFile{fset: token.NewFileSet(), File: want}
	g := &AstFile{fset: token.NewFileSet(), File: got}
	return w.String() == g.String()
}
