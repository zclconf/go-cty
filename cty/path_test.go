package cty_test

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestPathApply(t *testing.T) {
	tests := []struct {
		Start   cty.Value
		Path    cty.Path
		Want    cty.Value
		WantErr string
	}{
		{
			cty.StringVal("hello"),
			nil,
			cty.StringVal("hello"),
			``,
		},
		{
			cty.StringVal("hello"),
			(cty.Path)(nil).Index(cty.StringVal("boop")),
			cty.NilVal,
			`at step 0: not a map type`,
		},
		{
			cty.StringVal("hello"),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.NilVal,
			`at step 0: not a list type`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.StringVal("hello"),
			``,
		},
		{
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.StringVal("hello"),
			``,
		},
		{
			cty.ListValEmpty(cty.String),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.NilVal,
			`at step 0: value does not have given index key`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(1)),
			cty.NilVal,
			`at step 0: value does not have given index key`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)).GetAttr("foo"),
			cty.NilVal,
			`at step 1: not an object type`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.EmptyObjectVal,
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)).GetAttr("foo"),
			cty.NilVal,
			`at step 1: object has no attribute "foo"`,
		},
		{
			cty.NullVal(cty.List(cty.String)),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.NilVal,
			`at step 0: cannot index a null value`,
		},
		{
			cty.NullVal(cty.Map(cty.String)),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.NilVal,
			`at step 0: cannot index a null value`,
		},
		{
			cty.NullVal(cty.EmptyObject),
			(cty.Path)(nil).GetAttr("foo"),
			cty.NilVal,
			`at step 0: cannot access attributes on a null value`,
		},
		{
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{cty.StringVal("hello")}).Mark(2),
			}).Mark(1),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)).Index(cty.NumberIntVal(0)),
			cty.StringVal("hello").Mark(1).Mark(2),
			``,
		},
		{
			cty.TupleVal([]cty.Value{
				cty.ListVal([]cty.Value{cty.StringVal("hello")}).Mark(2),
			}).Mark(1),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)).Index(cty.NumberIntVal(0)),
			cty.StringVal("hello").Mark(1).Mark(2),
			``,
		},
		{
			cty.MapVal(map[string]cty.Value{
				"hello": cty.StringVal("there"),
			}).Mark(1),
			(cty.Path)(nil).Index(cty.StringVal("hello")),
			cty.StringVal("there").Mark(1),
			``,
		},
		{
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("there"),
			}).Mark(1),
			cty.GetAttrPath("hello"),
			cty.StringVal("there").Mark(1),
			``,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello").Mark(1),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.StringVal("hello").Mark(1),
			``,
		},
		{
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello").Mark(1),
			}),
			(cty.Path)(nil).Index(cty.NumberIntVal(0)),
			cty.StringVal("hello").Mark(1),
			``,
		},
		{
			cty.MapVal(map[string]cty.Value{
				"hello": cty.StringVal("there").Mark(1),
			}),
			(cty.Path)(nil).Index(cty.StringVal("hello")),
			cty.StringVal("there").Mark(1),
			``,
		},
		{
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("there").Mark(1),
			}),
			cty.GetAttrPath("hello"),
			cty.StringVal("there").Mark(1),
			``,
		},
		{
			cty.SetVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{"x": cty.StringVal("X1")}),
				cty.ObjectVal(map[string]cty.Value{"x": cty.StringVal("X2")}),
			}),
			cty.IndexPath(cty.ObjectVal(map[string]cty.Value{"x": cty.StringVal("X1")})).GetAttr("x"),
			cty.StringVal("X1"),
			``,
		},
		{
			cty.SetVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{"x": cty.StringVal("X1")}),
				cty.ObjectVal(map[string]cty.Value{"x": cty.StringVal("X2")}),
			}),
			cty.IndexPath(cty.ObjectVal(map[string]cty.Value{"x": cty.StringVal("X3")})).GetAttr("x"),
			cty.NilVal,
			`at step 0: value does not have given element`,
		},
		{
			cty.UnknownVal(cty.Set(cty.Object(map[string]cty.Type{"x": cty.String}))),
			cty.IndexPath(cty.ObjectVal(map[string]cty.Value{"x": cty.StringVal("X3")})).GetAttr("x"),
			cty.UnknownVal(cty.String),
			``,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v %#v", test.Start, test.Path), func(t *testing.T) {
			got, gotErr := test.Path.Apply(test.Start)
			t.Logf("testing path apply\nstart: %#v\npath:  %#v", test.Start, test.Path)

			if test.WantErr != "" {
				if gotErr == nil {
					t.Fatalf("succeeded, but want error\nwant error: %s", test.WantErr)
				}

				if gotErrStr := gotErr.Error(); gotErrStr != test.WantErr {
					t.Fatalf("wrong error\ngot error:  %s\nwant error: %s", gotErrStr, test.WantErr)
				}
				return
			}

			if gotErr != nil {
				t.Fatalf("failed, but want success\ngot error: %s", gotErr.Error())
			}
			if !test.Want.RawEquals(got) {
				t.Fatalf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestPathEquals(t *testing.T) {
	tests := []struct {
		A, B   cty.Path
		Equal  bool
		Prefix bool
	}{
		{
			A:      nil,
			B:      nil,
			Equal:  true,
			Prefix: true,
		},
		{
			A:      cty.Path{},
			B:      cty.Path{},
			Equal:  true,
			Prefix: true,
		},
		{
			A: cty.Path{nil},
			B: cty.Path{cty.GetAttrStep{Name: "attr"}},
		},
		{
			A: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.UnknownVal(cty.String)},
				cty.GetAttrStep{Name: "attr"},
			},
			B: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.StringVal("key")},
				cty.GetAttrStep{Name: "attr"},
			},
		},
		{
			A: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.ListVal([]cty.Value{cty.UnknownVal(cty.String)})},
				cty.GetAttrStep{Name: "attr"},
			},
			B: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.ListVal([]cty.Value{cty.StringVal("known")})},
				cty.GetAttrStep{Name: "attr"},
			},
		},
		{
			A: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.UnknownVal(cty.String)},
			},
			B: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.StringVal("known")},
				cty.GetAttrStep{Name: "attr"},
			},
		},
		{
			A: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.StringVal("known")},
			},
			B: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.StringVal("known")},
				cty.GetAttrStep{Name: "attr"},
			},
		},
		{
			A: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.StringVal("known")},
				cty.GetAttrStep{Name: "attr"},
			},
			B: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.StringVal("known")},
			},
			Prefix: true,
		},
		{
			A: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.UnknownVal(cty.String)},
			},
			B: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.UnknownVal(cty.String)},
			},
			Prefix: true,
			Equal:  true,
		},
		{
			A: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.NumberFloatVal(0)},
				cty.GetAttrStep{Name: "attr"},
			},
			B: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.NumberIntVal(0)},
				cty.GetAttrStep{Name: "attr"},
			},
			Equal:  true,
			Prefix: true,
		},
		{
			A: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.NumberIntVal(1)},
				cty.GetAttrStep{Name: "attr"},
			},
			B: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.NumberIntVal(0)},
				cty.GetAttrStep{Name: "attr"},
			},
		},

		// tests for convenience methods
		{
			A: cty.Path{
				cty.GetAttrStep{Name: "attr"},
			},
			B:      cty.GetAttrPath("attr"),
			Prefix: true,
			Equal:  true,
		},
		{
			A: cty.Path{
				cty.IndexStep{Key: cty.NumberIntVal(0)},
			},
			B:      cty.IndexPath(cty.NumberIntVal(0)),
			Prefix: true,
			Equal:  true,
		},
		{
			A: cty.Path{
				cty.IndexStep{Key: cty.NumberIntVal(0)},
			},
			B:      cty.IndexIntPath(0),
			Prefix: true,
			Equal:  true,
		},
		{
			A: cty.Path{
				cty.IndexStep{Key: cty.StringVal("key")},
			},
			B:      cty.IndexStringPath("key"),
			Prefix: true,
			Equal:  true,
		},
		{
			A: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.NumberIntVal(0)},
			},
			B:      cty.GetAttrPath("attr").IndexInt(0),
			Prefix: true,
			Equal:  true,
		},
		{
			A: cty.Path{
				cty.GetAttrStep{Name: "attr"},
				cty.IndexStep{Key: cty.StringVal("key")},
			},
			B:      cty.GetAttrPath("attr").IndexString("key"),
			Prefix: true,
			Equal:  true,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d-%#v", i, test.A), func(t *testing.T) {
			if test.Equal != test.A.Equals(test.B) {
				t.Fatalf("%#v.Equals(%#v) != %t", test.A, test.B, test.Equal)
			}
			if test.Prefix != test.A.HasPrefix(test.B) {
				t.Fatalf("%#v.HasPrefix(%#v) != %t", test.A, test.B, test.Prefix)
			}
		})
	}
}
