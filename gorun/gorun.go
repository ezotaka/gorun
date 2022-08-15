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
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/ezotaka/golib/ezast"
)

// Execute func 'fn' in 'file' using 'go run' shell command.
//
// Func 'fn' is treated as if it were the main function of the main package
func Exec(file, fn string) error {
	if file == "" {
		return errors.New("file must not be empty")
	}
	if fn == "" {
		return errors.New("fn must not be empty")
	}

	ast, err := convert(file, fn)
	if err != nil {
		return err
	}

	tmpFile, cleaner, err := ast.SaveTemp("main.go")
	defer cleaner()
	if err != nil {
		return err
	}

	err = goRun(tmpFile)
	if err != nil {
		return err
	}

	return nil
}

// Convert the source code so that it can be executed with the 'go run' command
func convert(file, fn string) (*ezast.AstFile, error) {
	ast, err := ezast.NewAstFile(file)
	if err != nil {
		return nil, err
	}

	if !ast.ContainsFunc(fn) {
		return nil, fmt.Errorf("file '%s' has no func '%s'", file, fn)
	}

	err = ast.ChangePackage("main")
	if err != nil {
		return nil, err
	}

	err = ast.SwapSimpleFuncs(fn, "main", false)
	if err != nil {
		return nil, err
	}

	return ast, nil
}

// Execute shell command 'go run [file]'.
func goRun(file string) error {
	if _, err := exec.LookPath("go"); err != nil {
		return err
	}

	cmd := exec.Command("go", "run", file)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
