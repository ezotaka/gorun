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

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ezotaka/gorun/ezast"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	app := kingpin.New("gorun", "Run any func in any file as main func")

	file := app.Arg("goFilePath", "Path to go file to run").Required().String()
	fn := app.Arg("funcName", "Name of func to run").Required().String()

	kingpin.MustParse(app.Parse(os.Args[1:]))

	// .go extension can be omitted
	if !strings.HasSuffix(*file, ".go") {
		*file = *file + ".go"
	}

	ast, err := convert(*file, *fn)
	if err != nil {
		fmt.Println(err)
		return
	}

	tmpFile, cleaner, err := ast.SaveTemp()
	defer cleaner()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = goRun(tmpFile)
	if err != nil {
		fmt.Println(err)
		return
	}
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

func goRun(path string) error {
	if _, err := exec.LookPath("go"); err != nil {
		return err
	}

	cmd := exec.Command("go", "run", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
