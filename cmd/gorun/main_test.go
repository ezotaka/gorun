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
	"testing"

	"github.com/ezotaka/gorun/ezast"
)

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
