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
	"os"
	"strings"
	"testing"

	"github.com/ezotaka/golib/ezast"
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
		})
	}
}

func Test_convert(t *testing.T) {
	type args struct {
		file string
		fn   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "change package name to main, and swap test1 func and main func",
			args: args{
				file: "testdata/src1.go",
				fn:   "test1",
			},
			want: "testdata/ans1.go",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, _ := convert(tt.args.file, tt.args.fn)
			want, _ := ezast.NewAstFile(tt.want)
			if !ezast.AssertEqualAstFile(got.File, want.File) {
				t.Errorf("convert() = %v, want %v", got.String(), want.String())
			}
		})
	}
}
