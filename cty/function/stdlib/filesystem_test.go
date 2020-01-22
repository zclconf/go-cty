package stdlib

import (
	"fmt"
	"path/filepath"
	"testing"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/zclconf/go-cty/cty"
)

func TestFileExists(t *testing.T) {
	tests := []struct {
		Path cty.Value
		Want cty.Value
		Err  bool
	}{
		{
			cty.StringVal("testdata/hello.txt"),
			cty.BoolVal(true),
			false,
		},
		{
			cty.StringVal(""), // empty path
			cty.BoolVal(false),
			true,
		},
		{
			cty.StringVal("testdata/missing"),
			cty.BoolVal(false),
			false, // no file exists
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("FileExists(\".\", %#v)", test.Path), func(t *testing.T) {
			got, err := FileExists(".", test.Path)

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestFileSet(t *testing.T) {
	tests := []struct {
		Path    cty.Value
		Pattern cty.Value
		Want    cty.Value
		Err     bool
	}{
		{
			cty.StringVal("."),
			cty.StringVal("testdata*"),
			cty.SetValEmpty(cty.String),
			false,
		},
		{
			cty.StringVal("."),
			cty.StringVal("testdata"),
			cty.SetValEmpty(cty.String),
			false,
		},
		{
			cty.StringVal("."),
			cty.StringVal("{testdata,missing}"),
			cty.SetValEmpty(cty.String),
			false,
		},
		{
			cty.StringVal("."),
			cty.StringVal("testdata/missing"),
			cty.SetValEmpty(cty.String),
			false,
		},
		{
			cty.StringVal("."),
			cty.StringVal("testdata/missing*"),
			cty.SetValEmpty(cty.String),
			false,
		},
		{
			cty.StringVal("."),
			cty.StringVal("*/missing"),
			cty.SetValEmpty(cty.String),
			false,
		},
		{
			cty.StringVal("."),
			cty.StringVal("**/missing"),
			cty.SetValEmpty(cty.String),
			false,
		},
		{
			cty.StringVal("."),
			cty.StringVal("testdata/*.txt"),
			cty.SetVal([]cty.Value{
				cty.StringVal("testdata/hello.txt"),
			}),
			false,
		},
		{
			cty.StringVal("."),
			cty.StringVal("testdata/hello.txt"),
			cty.SetVal([]cty.Value{
				cty.StringVal("testdata/hello.txt"),
			}),
			false,
		},
		{
			cty.StringVal("."),
			cty.StringVal("testdata/hello.???"),
			cty.SetVal([]cty.Value{
				cty.StringVal("testdata/hello.txt"),
			}),
			false,
		},
		{
			cty.StringVal("."),
			cty.StringVal("testdata/hello*"),
			cty.SetVal([]cty.Value{
				cty.StringVal("testdata/hello.tmpl"),
				cty.StringVal("testdata/hello.txt"),
			}),
			false,
		},
		{
			cty.StringVal("."),
			cty.StringVal("testdata/hello.{tmpl,txt}"),
			cty.SetVal([]cty.Value{
				cty.StringVal("testdata/hello.tmpl"),
				cty.StringVal("testdata/hello.txt"),
			}),
			false,
		},
		{
			cty.StringVal("."),
			cty.StringVal("*/hello.txt"),
			cty.SetVal([]cty.Value{
				cty.StringVal("testdata/hello.txt"),
			}),
			false,
		},
		{
			cty.StringVal("."),
			cty.StringVal("*/*.txt"),
			cty.SetVal([]cty.Value{
				cty.StringVal("testdata/hello.txt"),
			}),
			false,
		},
		{
			cty.StringVal("."),
			cty.StringVal("*/hello*"),
			cty.SetVal([]cty.Value{
				cty.StringVal("testdata/hello.tmpl"),
				cty.StringVal("testdata/hello.txt"),
			}),
			false,
		},
		{
			cty.StringVal("."),
			cty.StringVal("**/hello*"),
			cty.SetVal([]cty.Value{
				cty.StringVal("testdata/hello.tmpl"),
				cty.StringVal("testdata/hello.txt"),
			}),
			false,
		},
		{
			cty.StringVal("."),
			cty.StringVal("**/hello.{tmpl,txt}"),
			cty.SetVal([]cty.Value{
				cty.StringVal("testdata/hello.tmpl"),
				cty.StringVal("testdata/hello.txt"),
			}),
			false,
		},
		{
			cty.StringVal("."),
			cty.StringVal("["),
			cty.SetValEmpty(cty.String),
			true,
		},
		{
			cty.StringVal("."),
			cty.StringVal("\\"),
			cty.SetValEmpty(cty.String),
			true,
		},
		{
			cty.StringVal("testdata"),
			cty.StringVal("missing"),
			cty.SetValEmpty(cty.String),
			false,
		},
		{
			cty.StringVal("testdata"),
			cty.StringVal("missing*"),
			cty.SetValEmpty(cty.String),
			false,
		},
		{
			cty.StringVal("testdata"),
			cty.StringVal("*.txt"),
			cty.SetVal([]cty.Value{
				cty.StringVal("hello.txt"),
			}),
			false,
		},
		{
			cty.StringVal("testdata"),
			cty.StringVal("hello.txt"),
			cty.SetVal([]cty.Value{
				cty.StringVal("hello.txt"),
			}),
			false,
		},
		{
			cty.StringVal("testdata"),
			cty.StringVal("hello.???"),
			cty.SetVal([]cty.Value{
				cty.StringVal("hello.txt"),
			}),
			false,
		},
		{
			cty.StringVal("testdata"),
			cty.StringVal("hello*"),
			cty.SetVal([]cty.Value{
				cty.StringVal("hello.tmpl"),
				cty.StringVal("hello.txt"),
			}),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("FileSet(\".\", %#v, %#v)", test.Path, test.Pattern), func(t *testing.T) {
			got, err := FileSet(".", test.Path, test.Pattern)

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestBasename(t *testing.T) {
	tests := []struct {
		Path cty.Value
		Want cty.Value
		Err  bool
	}{
		{
			cty.StringVal("testdata/hello.txt"),
			cty.StringVal("hello.txt"),
			false,
		},
		{
			cty.StringVal("hello.txt"),
			cty.StringVal("hello.txt"),
			false,
		},
		{
			cty.StringVal(""),
			cty.StringVal("."),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Basename(%#v)", test.Path), func(t *testing.T) {
			got, err := Basename(test.Path)

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestDirname(t *testing.T) {
	tests := []struct {
		Path cty.Value
		Want cty.Value
		Err  bool
	}{
		{
			cty.StringVal("testdata/hello.txt"),
			cty.StringVal("testdata"),
			false,
		},
		{
			cty.StringVal("testdata/foo/hello.txt"),
			cty.StringVal("testdata/foo"),
			false,
		},
		{
			cty.StringVal("hello.txt"),
			cty.StringVal("."),
			false,
		},
		{
			cty.StringVal(""),
			cty.StringVal("."),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Dirname(%#v)", test.Path), func(t *testing.T) {
			got, err := Dirname(test.Path)

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestPathExpand(t *testing.T) {
	homePath, err := homedir.Dir()
	if err != nil {
		t.Fatalf("Error getting home directory: %v", err)
	}

	tests := []struct {
		Path cty.Value
		Want cty.Value
		Err  bool
	}{
		{
			cty.StringVal("~/test-file"),
			cty.StringVal(filepath.Join(homePath, "test-file")),
			false,
		},
		{
			cty.StringVal("~/another/test/file"),
			cty.StringVal(filepath.Join(homePath, "another/test/file")),
			false,
		},
		{
			cty.StringVal("/root/file"),
			cty.StringVal("/root/file"),
			false,
		},
		{
			cty.StringVal("/"),
			cty.StringVal("/"),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Dirname(%#v)", test.Path), func(t *testing.T) {
			got, err := Pathexpand(test.Path)

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
