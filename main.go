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
	"strings"

	"github.com/ezotaka/gorun/gorun"
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

	err := gorun.Exec(*file, *fn)
	if err != nil {
		fmt.Println(err)
	}
}
