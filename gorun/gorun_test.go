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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func getStdout(t *testing.T, fn func()) string {
	t.Helper()

	orgStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// execute function to be tested
	fn()

	w.Close()
	os.Stdout = orgStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read buf: %v", err)
	}
	return strings.TrimRight(buf.String(), "\n")
}

func TestExec(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	goModFile, err := findTowardRoot(".", "go.mod")
	if err != nil {
		t.Fatal(err)
	}
	goModDir := filepath.Dir(goModFile)
	const outOfGoModDir = "/this/file/maybe/out/of/go/module/dir/59432543040540/src1.go"

	type args struct {
		file string
		fn   string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Run func test1 in package testdata",
			args: args{
				file: "testdata/src1.go",
				fn:   "test1",
			},
			want: "test1",
		},
		{
			name: "absolute file path",
			args: args{
				file: filepath.Join(wd, "testdata/src1.go"),
				fn:   "test1",
			},
			want: "test1",
		},
		{
			name: "file is not found",
			args: args{
				file: "testdata/fileNotFound.go",
				fn:   "test1",
			},
			want:    "open testdata/fileNotFound.go: no such file or directory",
			wantErr: true,
		},
		{
			name: "fn is not found",
			args: args{
				file: "testdata/src1.go",
				fn:   "funcNotFound",
			},
			want:    "file 'testdata/src1.go' has no func 'funcNotFound'",
			wantErr: true,
		},
		{
			name: "file is out of current go module dir",
			args: args{
				file: outOfGoModDir,
				fn:   "test1",
			},
			want:    fmt.Sprintf("file '%s' must be in go module dir '%s'", outOfGoModDir, goModDir),
			wantErr: true,
		},
		{
			name: "file is empty",
			args: args{
				file: "",
				fn:   "test1",
			},
			want:    "file must not be empty",
			wantErr: true,
		},
		{
			name: "fn is empty",
			args: args{
				file: "testdata/src1.go",
				fn:   "",
			},
			want:    "fn must not be empty",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// DO NOT t.Parallel()
			//t.Parallel()

			var err error
			goRun := func() {
				err = Exec(tt.args.file, tt.args.fn)
			}
			got := getStdout(t, goRun)

			if tt.wantErr {
				if err == nil {
					t.Error("Exec() is expected to return error, but no error is returned")
				} else if err.Error() != tt.want {
					t.Errorf("Exec() error = %v, want error %v", err.Error(), tt.want)
				}
			} else {
				if got != tt.want {
					t.Errorf("Exec() = %v, want %v", got, tt.want)
				}
			}

			// Exec() changes working directory temporarily,
			// so check that current directory is not changed
			doneWd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			} else if doneWd != wd {
				t.Errorf("Exec() doesn't clean up. got wd = %s, want wd = %s", doneWd, wd)
			}
		})
	}
}

func Test_findTowardRoot(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	abs := func(path string) string {
		return filepath.Join(wd, path)
	}

	type args struct {
		baseDir string
		file    string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "found in absolute dir path",
			args: args{
				baseDir: abs("testdata/a/b/c"),
				file:    "d.txt",
			},
			want: abs("testdata/a/b/c/d.txt"),
		},
		{
			name: "found in relative dir path",
			args: args{
				baseDir: "testdata/a/b/c",
				file:    "d.txt",
			},
			want: abs("testdata/a/b/c/d.txt"),
		},
		{
			name: "found in parent of absolute dir path",
			args: args{
				baseDir: abs("testdata/a/b/c"),
				file:    "b.txt",
			},
			want: abs("testdata/a/b.txt"),
		},
		{
			name: "found in parent of relative dir path",
			args: args{
				baseDir: "testdata/a/b/c",
				file:    "b.txt",
			},
			want: abs("testdata/a/b.txt"),
		},
		{
			name: "dir with matching name are ignored",
			args: args{
				baseDir: "testdata/a/b/c",
				file:    "b",
			},
			want: abs("testdata/b"),
		},
		{
			name: "file not found",
			args: args{
				baseDir: "testdata/a/b/c",
				file:    "this_file_is_maybe_not_found_78438598435984324532",
			},
			wantErr: true,
		},
		{
			name: "baseDir not found",
			args: args{
				baseDir: "testdata/a/b/c/notFound",
				file:    "b.txt",
			},
			wantErr: true,
		},
		{
			name: "baseDir is empty",
			args: args{
				baseDir: "",
				file:    "b.txt",
			},
			wantErr: true,
		},
		{
			name: "file is empty",
			args: args{
				baseDir: "testdata/a/b/c",
				file:    "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findTowardRoot(tt.args.baseDir, tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("findTowardRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("findTowardRoot() = %v, want %v", got, tt.want)
			}
		})
	}
}
