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

package gorun

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/ezotaka/golib/ezast"
)

// Execute func 'fn' in 'file' using 'go test' shell command.
func Exec(goFilePath, fn string) error {
	if goFilePath == "" {
		return errors.New("file must not be empty")
	}
	if fn == "" {
		return errors.New("fn must not be empty")
	}

	absGoFilePath, err := filepath.Abs(goFilePath)
	if err != nil {
		return err
	}

	// Head to root dir and look for the go.mod file
	// if it's found, the root of go module is there
	goModFile, err := findTowardRoot(".", "go.mod")
	if err != nil {
		return fmt.Errorf("gorun.Exec() must be called in go module dir")
	}
	goModDir := filepath.Dir(goModFile)

	// goFilePath must be in go module dir
	if !strings.HasPrefix(absGoFilePath, goModDir) {
		return fmt.Errorf("file '%s' must be in go module dir '%s'", goFilePath, goModDir)
	}

	// create AST of source code
	ast, err := ezast.NewAstFile(goFilePath)
	if err != nil {
		return err
	}

	fd := ast.GetFuncDecl(fn)
	if fd == nil {
		return fmt.Errorf("file '%s' has no func '%s'", goFilePath, fn)
	}

	if len(fd.Type.Params.List) > 0 {
		return fmt.Errorf("func '%s' must have no args", fn)
	}

	// TODO: it does not work if package of goFile path includes func TestMain
	// test code for running func fn
	src := fmt.Sprintf(`package %s
	import (
		"testing"
	)
	
	func TestMain(m *testing.M) {
		%s()
	}
`, ast.Name.Name, fn)

	absDir := path.Dir(absGoFilePath)

	// create temp test file where go file is
	f, err := os.CreateTemp(absDir, "*_test.go")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	if _, err = f.Write([]byte(src)); err != nil {
		return err
	}

	if err = f.Close(); err != nil {
		return err
	}

	// relative package path from go module root
	relDir, err := filepath.Rel(goModDir, absDir)
	if err != nil {
		return err
	}
	pkg := "./" + relDir

	// move to the root of go module, and run 'go test'
	if wd, err := os.Getwd(); err != nil {
		return err
	} else {
		defer os.Chdir(wd) // clean up
	}

	if err = os.Chdir(goModDir); err != nil {
		return err
	}

	if err = runGoTest(pkg, "-v"); err != nil {
		return err
	}

	return nil
}

func findTowardRoot(baseDir, file string) (string, error) {
	if file == "" {
		return "", fmt.Errorf("file must not be empty")
	}
	if f, err := os.Stat(baseDir); os.IsNotExist(err) || !f.IsDir() {
		return "", fmt.Errorf("dir '%s' is not found", baseDir)
	}

	dir := baseDir
	for dir, err := filepath.Abs(dir); err == nil && dir != "/"; {
		path := filepath.Join(dir, file)

		if f, err := os.Stat(path); err == nil && !f.IsDir() {
			if abs, err := filepath.Abs(path); err == nil {
				return abs, nil
			}
		}
		dir = filepath.Join(dir, "..")
	}

	return "", fmt.Errorf("file '%s' is not found", file)
}

// Execute shell command 'go test'.
func runGoTest(args ...string) error {
	if _, err := exec.LookPath("go"); err != nil {
		return err
	}

	w := newGoTestWriter()
	defer w.Close()

	newArgs := append([]string{"test"}, args...)
	cmd := exec.Command("go", newArgs...)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

type goTestWriter struct {
	w *io.PipeWriter
}

func newGoTestWriter() *goTestWriter {
	r, w := io.Pipe()
	go func() {
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			line := sc.Text()
			// ignore test result line which started "ok"
			// TODO: need longer match
			if !strings.HasPrefix(line, "ok") {
				fmt.Println(line)
			}
		}
	}()
	return &goTestWriter{w}
}

func (g *goTestWriter) Write(p []byte) (n int, err error) {
	g.w.Write(p)
	return len(p), nil
}

func (g *goTestWriter) Close() error {
	return g.w.Close()
}
