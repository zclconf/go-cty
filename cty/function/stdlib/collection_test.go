package stdlib

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestHasIndex(t *testing.T) {
	tests := []struct {
		Collection cty.Value
		Key        cty.Value
		Want       cty.Value
	}{
		{
			cty.ListValEmpty(cty.Number),
			cty.NumberIntVal(2),
			cty.False,
		},
		{
			cty.ListVal([]cty.Value{cty.True}),
			cty.NumberIntVal(0),
			cty.True,
		},
		{
			cty.ListVal([]cty.Value{cty.True}),
			cty.StringVal("hello"),
			cty.False,
		},
		{
			cty.MapValEmpty(cty.Bool),
			cty.StringVal("hello"),
			cty.False,
		},
		{
			cty.MapVal(map[string]cty.Value{"hello": cty.True}),
			cty.StringVal("hello"),
			cty.True,
		},
		{
			cty.EmptyTupleVal,
			cty.StringVal("hello"),
			cty.False,
		},
		{
			cty.EmptyTupleVal,
			cty.NumberIntVal(0),
			cty.False,
		},
		{
			cty.TupleVal([]cty.Value{cty.True}),
			cty.NumberIntVal(0),
			cty.True,
		},
		{
			cty.ListValEmpty(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.UnknownVal(cty.List(cty.Bool)),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.ListValEmpty(cty.Number),
			cty.DynamicVal,
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Bool),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("HasIndex(%#v,%#v)", test.Collection, test.Key), func(t *testing.T) {
			got, err := HasIndex(test.Collection, test.Key)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestChunklist(t *testing.T) {
	tests := []struct {
		List cty.Value
		Len  cty.Value
		Want cty.Value
		Err  string
	}{
		{
			cty.ListValEmpty(cty.String),
			cty.NumberIntVal(2),
			cty.ListValEmpty(cty.List(cty.String)),
			``,
		},
		{
			cty.UnknownVal(cty.List(cty.String)),
			cty.NumberIntVal(2),
			cty.UnknownVal(cty.List(cty.List(cty.String))),
			``,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("a"),
			}),
			cty.NumberIntVal(2),
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("a"),
				}),
			}),
			``,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("a").Mark("b"),
			}),
			cty.NumberIntVal(2),
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("a").Mark("b"),
				}),
			}),
			``,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("a"),
			}).Mark("a"),
			cty.NumberIntVal(2),
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("a"),
				}),
			}).Mark("a"),
			``,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("a").Mark("b"),
			}).Mark("a"),
			cty.NumberIntVal(2),
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("a").Mark("b"),
				}),
			}).Mark("a"),
			``,
		},
		{
			cty.ListVal([]cty.Value{
				cty.UnknownVal(cty.String),
			}),
			cty.NumberIntVal(2),
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.UnknownVal(cty.String),
				}),
			}),
			``,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
			}),
			cty.NumberIntVal(2),
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("a"),
					cty.StringVal("b"),
				}),
			}),
			``,
		},
		{ // Multiple result elements, one shorter
			cty.ListVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
				cty.StringVal("c"),
			}),
			cty.NumberIntVal(2),
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("a"),
					cty.StringVal("b"),
				}),
				cty.ListVal([]cty.Value{
					cty.StringVal("c"),
				}),
			}),
			``,
		},
		{ // Multiple result elements, all "full"
			cty.ListVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
				cty.StringVal("c"),
				cty.StringVal("d"),
				cty.StringVal("e"),
				cty.StringVal("f"),
			}),
			cty.NumberIntVal(2),
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("a"),
					cty.StringVal("b"),
				}),
				cty.ListVal([]cty.Value{
					cty.StringVal("c"),
					cty.StringVal("d"),
				}),
				cty.ListVal([]cty.Value{
					cty.StringVal("e"),
					cty.StringVal("f"),
				}),
			}),
			``,
		},
		{ // We treat length zero as infinite length
			cty.ListVal([]cty.Value{
				cty.StringVal("a"),
			}),
			cty.Zero,
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("a"),
				}),
			}),
			``,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("a"),
			}).Mark("a"),
			cty.Zero,
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("a"),
				}),
			}).Mark("a"),
			``,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("a"),
			}),
			cty.Zero.Mark("a"),
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("a"),
				}),
			}).Mark("a"),
			``,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("a").Mark("b"),
			}),
			cty.Zero,
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("a").Mark("b"),
				}),
			}),
			``,
		},
		{
			cty.ListValEmpty(cty.String),
			cty.NumberIntVal(-1),
			cty.NilVal,
			`the size argument must be positive`,
		},
		{
			cty.ListValEmpty(cty.String),
			cty.PositiveInfinity,
			cty.NilVal,
			`invalid size: value must be a whole number, between -9223372036854775808 and 9223372036854775807`,
		},
		{
			cty.ListValEmpty(cty.String),
			cty.NumberFloatVal(1.5),
			cty.NilVal,
			`invalid size: value must be a whole number, between -9223372036854775808 and 9223372036854775807`,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Chunklist(%#v, %#v)", test.List, test.Len), func(t *testing.T) {
			got, err := Chunklist(test.List, test.Len)
			if test.Err != "" {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				if got, want := err.Error(), test.Err; got != want {
					t.Fatalf("wrong error\ngot:  %s\nwant: %s", got, want)
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

func TestContains(t *testing.T) {
	listOfStrings := cty.ListVal([]cty.Value{
		cty.StringVal("the"),
		cty.StringVal("quick"),
		cty.StringVal("brown"),
		cty.StringVal("fox"),
	})
	listOfInts := cty.ListVal([]cty.Value{
		cty.NumberIntVal(1),
		cty.NumberIntVal(2),
		cty.NumberIntVal(3),
		cty.NumberIntVal(4),
	})
	listWithUnknown := cty.ListVal([]cty.Value{
		cty.StringVal("the"),
		cty.StringVal("quick"),
		cty.StringVal("brown"),
		cty.UnknownVal(cty.String),
	})

	tests := []struct {
		List  cty.Value
		Value cty.Value
		Want  cty.Value
		Err   bool
	}{
		{
			listOfStrings,
			cty.StringVal("the"),
			cty.BoolVal(true),
			false,
		},
		{
			listWithUnknown,
			cty.StringVal("the"),
			cty.BoolVal(true),
			false,
		},
		{
			listWithUnknown,
			cty.StringVal("orange"),
			cty.UnknownVal(cty.Bool),
			false,
		},
		{
			listOfStrings,
			cty.StringVal("penguin"),
			cty.BoolVal(false),
			false,
		},
		{
			listOfInts,
			cty.NumberIntVal(1),
			cty.BoolVal(true),
			false,
		},
		{
			listOfInts,
			cty.NumberIntVal(42),
			cty.BoolVal(false),
			false,
		},
		{ // And now we mix and match
			listOfInts,
			cty.StringVal("1"),
			cty.BoolVal(false),
			false,
		},
		{ // Check a list with an unknown value
			cty.ListVal([]cty.Value{
				cty.UnknownVal(cty.String),
				cty.StringVal("quick"),
				cty.StringVal("brown"),
				cty.StringVal("fox"),
			}),
			cty.StringVal("quick"),
			cty.BoolVal(true),
			false,
		},
		{
			cty.ListVal([]cty.Value{
				cty.UnknownVal(cty.String),
				cty.StringVal("brown"),
				cty.StringVal("fox"),
			}),
			cty.StringVal("quick"),
			cty.UnknownVal(cty.Bool),
			false,
		},
		{ // set val
			cty.SetVal([]cty.Value{
				cty.StringVal("quick"),
				cty.StringVal("brown"),
				cty.StringVal("fox"),
			}),
			cty.StringVal("quick"),
			cty.BoolVal(true),
			false,
		},
		{
			cty.SetVal([]cty.Value{
				cty.UnknownVal(cty.String),
				cty.StringVal("brown"),
				cty.StringVal("fox"),
			}),
			cty.StringVal("quick"),
			cty.UnknownVal(cty.Bool),
			false,
		},
		{ // nested unknown
			cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.UnknownVal(cty.String),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.StringVal("b"),
			}),
			cty.UnknownVal(cty.Bool),
			false,
		},
		{ // tuple val
			cty.TupleVal([]cty.Value{
				cty.StringVal("quick"),
				cty.StringVal("brown"),
				cty.NumberIntVal(3),
			}),
			cty.NumberIntVal(3),
			cty.BoolVal(true),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("contains(%#v, %#v)", test.List, test.Value), func(t *testing.T) {
			got, err := Contains(test.List, test.Value)

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

func TestMerge(t *testing.T) {
	tests := []struct {
		Values []cty.Value
		Want   cty.Value
		Err    bool
	}{
		{
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("b"),
				}),
				cty.MapVal(map[string]cty.Value{
					"c": cty.StringVal("d"),
				}),
			},
			cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("b"),
				"c": cty.StringVal("d"),
			}),
			false,
		},
		{ // handle unknowns
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.UnknownVal(cty.String),
				}),
				cty.MapVal(map[string]cty.Value{
					"c": cty.StringVal("d"),
				}),
			},
			cty.MapVal(map[string]cty.Value{
				"a": cty.UnknownVal(cty.String),
				"c": cty.StringVal("d"),
			}),
			false,
		},
		{ // handle null map
			[]cty.Value{
				cty.NullVal(cty.Map(cty.String)),
				cty.MapVal(map[string]cty.Value{
					"c": cty.StringVal("d"),
				}),
			},
			cty.MapVal(map[string]cty.Value{
				"c": cty.StringVal("d"),
			}),
			false,
		},
		{ // all inputs are null
			[]cty.Value{
				cty.NullVal(cty.Map(cty.String)),
				cty.NullVal(cty.Object(map[string]cty.Type{
					"a": cty.List(cty.String),
				})),
			},
			cty.EmptyObjectVal,
			false,
		},
		{ // single null input
			[]cty.Value{
				cty.MapValEmpty(cty.String),
			},
			cty.MapValEmpty(cty.String),
			false,
		},
		{ // handle null object
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"c": cty.StringVal("d"),
				}),
				cty.NullVal(cty.Object(map[string]cty.Type{
					"a": cty.List(cty.String),
				})),
			},
			cty.ObjectVal(map[string]cty.Value{
				"c": cty.StringVal("d"),
			}),
			false,
		},
		{ // handle unknowns
			[]cty.Value{
				cty.UnknownVal(cty.Map(cty.String)),
				cty.MapVal(map[string]cty.Value{
					"c": cty.StringVal("d"),
				}),
			},
			cty.UnknownVal(cty.Map(cty.String)),
			false,
		},
		{ // handle dynamic unknown
			[]cty.Value{
				cty.UnknownVal(cty.DynamicPseudoType),
				cty.MapVal(map[string]cty.Value{
					"c": cty.StringVal("d"),
				}),
			},
			cty.DynamicVal,
			false,
		},
		{ // merge with conflicts is ok, last in wins
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("b"),
					"c": cty.StringVal("d"),
				}),
				cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("x"),
				}),
			},
			cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("x"),
				"c": cty.StringVal("d"),
			}),
			false,
		},
		{ // only accept maps
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("b"),
					"c": cty.StringVal("d"),
				}),
				cty.ListVal([]cty.Value{
					cty.StringVal("a"),
					cty.StringVal("x"),
				}),
			},
			cty.NilVal,
			true,
		},

		{ // argument error, for a null type
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("b"),
				}),
				cty.NullVal(cty.String),
			},
			cty.NilVal,
			true,
		},
		{ // merge maps of maps
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.MapVal(map[string]cty.Value{
						"b": cty.StringVal("c"),
					}),
				}),
				cty.MapVal(map[string]cty.Value{
					"d": cty.MapVal(map[string]cty.Value{
						"e": cty.StringVal("f"),
					}),
				}),
			},
			cty.MapVal(map[string]cty.Value{
				"a": cty.MapVal(map[string]cty.Value{
					"b": cty.StringVal("c"),
				}),
				"d": cty.MapVal(map[string]cty.Value{
					"e": cty.StringVal("f"),
				}),
			}),
			false,
		},
		{ // map of lists
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.ListVal([]cty.Value{
						cty.StringVal("b"),
						cty.StringVal("c"),
					}),
				}),
				cty.MapVal(map[string]cty.Value{
					"d": cty.ListVal([]cty.Value{
						cty.StringVal("e"),
						cty.StringVal("f"),
					}),
				}),
			},
			cty.MapVal(map[string]cty.Value{
				"a": cty.ListVal([]cty.Value{
					cty.StringVal("b"),
					cty.StringVal("c"),
				}),
				"d": cty.ListVal([]cty.Value{
					cty.StringVal("e"),
					cty.StringVal("f"),
				}),
			}),
			false,
		},
		{ // merge map of various kinds
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.ListVal([]cty.Value{
						cty.StringVal("b"),
						cty.StringVal("c"),
					}),
				}),
				cty.MapVal(map[string]cty.Value{
					"d": cty.MapVal(map[string]cty.Value{
						"e": cty.StringVal("f"),
					}),
				}),
			},
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.ListVal([]cty.Value{
					cty.StringVal("b"),
					cty.StringVal("c"),
				}),
				"d": cty.MapVal(map[string]cty.Value{
					"e": cty.StringVal("f"),
				}),
			}),
			false,
		},
		{ // merge objects of various shapes
			[]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.ListVal([]cty.Value{
						cty.StringVal("b"),
					}),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"d": cty.DynamicVal,
				}),
			},
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.ListVal([]cty.Value{
					cty.StringVal("b"),
				}),
				"d": cty.DynamicVal,
			}),
			false,
		},
		{ // merge maps and objects
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.ListVal([]cty.Value{
						cty.StringVal("b"),
					}),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"d": cty.NumberIntVal(2),
				}),
			},
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.ListVal([]cty.Value{
					cty.StringVal("b"),
				}),
				"d": cty.NumberIntVal(2),
			}),
			false,
		},
		{ // attr a type and value is overridden
			[]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.ListVal([]cty.Value{
						cty.StringVal("b"),
					}),
					"b": cty.StringVal("b"),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"e": cty.StringVal("f"),
					}),
				}),
			},
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.ObjectVal(map[string]cty.Value{
					"e": cty.StringVal("f"),
				}),
				"b": cty.StringVal("b"),
			}),
			false,
		},
		{ // argument error: non map type
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.ListVal([]cty.Value{
						cty.StringVal("b"),
						cty.StringVal("c"),
					}),
				}),
				cty.ListVal([]cty.Value{
					cty.StringVal("d"),
					cty.StringVal("e"),
				}),
			},
			cty.NilVal,
			true,
		},
		{ // Empty maps are allowed in merge
			[]cty.Value{
				cty.MapValEmpty(cty.String),
				cty.MapValEmpty(cty.String),
			},
			cty.MapValEmpty(cty.String),
			false,
		},
		{ // Preserve marks from chosen elements
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("a").Mark("first"),
					"c": cty.StringVal("c"),
					"d": cty.StringVal("d").Mark("first"),
				}),
				cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("a"),
					"b": cty.StringVal("b").Mark("second"),
					"c": cty.StringVal("c").Mark("second"),
				}),
			},
			cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("a"),
				"b": cty.StringVal("b").Mark("second"),
				"c": cty.StringVal("c").Mark("second"),
				"d": cty.StringVal("d").Mark("first"),
			}),
			false,
		},
		{ // Marks on the collections must be merged, even if empty
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("a"),
				}).Mark("first"),
				cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("a"),
					"b": cty.StringVal("b"),
				}).Mark("second"),
				cty.MapValEmpty(cty.String).Mark("third"),
			},
			cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("a"),
				"b": cty.StringVal("b"),
			}).WithMarks(cty.NewValueMarks("first", "second", "third")),
			false,
		},
		{ // Similar test but where all args are the same object type
			[]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("a"),
					"b": cty.NullVal(cty.String),
				}).Mark("first"),
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("A"),
					"b": cty.StringVal("B"),
				}).Mark("second"),
			},
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.StringVal("A"),
				"b": cty.StringVal("B"),
			}).WithMarks(cty.NewValueMarks("first", "second")),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("merge(%#v)", test.Values), func(t *testing.T) {
			got, err := Merge(test.Values...)

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

func TestIndex(t *testing.T) {
	tests := []struct {
		Collection cty.Value
		Key        cty.Value
		Want       cty.Value
	}{
		{
			cty.ListVal([]cty.Value{cty.True}),
			cty.NumberIntVal(0),
			cty.True,
		},
		{
			cty.MapVal(map[string]cty.Value{"hello": cty.True}),
			cty.StringVal("hello"),
			cty.True,
		},
		{
			cty.TupleVal([]cty.Value{cty.True, cty.StringVal("hello")}),
			cty.NumberIntVal(0),
			cty.True,
		},
		{
			cty.TupleVal([]cty.Value{cty.True, cty.StringVal("hello")}),
			cty.NumberIntVal(1),
			cty.StringVal("hello"),
		},
		{
			cty.ListValEmpty(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.UnknownVal(cty.List(cty.Bool)),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.ListValEmpty(cty.Number),
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{
			cty.MapValEmpty(cty.Number),
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.StringVal("hello"),
			cty.DynamicVal,
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.DynamicVal,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Index(%#v,%#v)", test.Collection, test.Key), func(t *testing.T) {
			got, err := Index(test.Collection, test.Key)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestLength(t *testing.T) {
	tests := []struct {
		Collection cty.Value
		Want       cty.Value
	}{
		{
			cty.ListValEmpty(cty.Number),
			cty.NumberIntVal(0),
		},
		{
			cty.ListVal([]cty.Value{cty.True}),
			cty.NumberIntVal(1),
		},
		{
			cty.SetValEmpty(cty.Number),
			cty.NumberIntVal(0),
		},
		{
			cty.SetVal([]cty.Value{cty.True}),
			cty.NumberIntVal(1),
		},
		{
			cty.SetVal([]cty.Value{cty.True, cty.False}),
			cty.NumberIntVal(2),
		},
		{
			cty.SetVal([]cty.Value{cty.True, cty.UnknownVal(cty.Bool)}),
			cty.UnknownVal(cty.Number), // Don't know if the unknown in the input represents cty.True or cty.False
		},
		{
			cty.SetVal([]cty.Value{cty.UnknownVal(cty.Bool)}),
			cty.NumberIntVal(1), // Will be one regardless of what value the unknown in the input is representing
		},
		{
			cty.MapValEmpty(cty.Bool),
			cty.NumberIntVal(0),
		},
		{
			cty.MapVal(map[string]cty.Value{"hello": cty.True}),
			cty.NumberIntVal(1),
		},
		{
			cty.EmptyTupleVal,
			cty.NumberIntVal(0),
		},
		{
			cty.TupleVal([]cty.Value{cty.True}),
			cty.NumberIntVal(1),
		},
		{
			cty.UnknownVal(cty.List(cty.Bool)),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{ // Marked collections return a marked length
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
				cty.StringVal("world"),
			}).Mark("secret"),
			cty.NumberIntVal(2).Mark("secret"),
		},
		{ // Marks on values in unmarked collections do not propagate
			cty.ListVal([]cty.Value{
				cty.StringVal("hello").Mark("a"),
				cty.StringVal("world").Mark("b"),
			}),
			cty.NumberIntVal(2),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Length(%#v)", test.Collection), func(t *testing.T) {
			got, err := Length(test.Collection)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestLookup(t *testing.T) {
	tests := []struct {
		Collection cty.Value
		Key        cty.Value
		Default    cty.Value
		Want       cty.Value
	}{
		{
			cty.MapValEmpty(cty.String),
			cty.StringVal("baz"),
			cty.StringVal("foo"),
			cty.StringVal("foo"),
		},
		{
			cty.MapVal(map[string]cty.Value{
				"foo": cty.StringVal("bar"),
			}),
			cty.StringVal("foo"),
			cty.StringVal("nope"),
			cty.StringVal("bar"),
		},
		{ // successful marked collection lookup returns marked value
			cty.MapVal(map[string]cty.Value{
				"boop": cty.StringVal("beep"),
			}).Mark("a"),
			cty.StringVal("boop"),
			cty.StringVal("nope"),
			cty.StringVal("beep").Mark("a"),
		},
		{ // apply collection marks to unknown return vaue
			cty.MapVal(map[string]cty.Value{
				"boop": cty.StringVal("beep"),
				"frob": cty.UnknownVal(cty.String),
			}).Mark("a"),
			cty.StringVal("boop"),
			cty.StringVal("nope"),
			cty.UnknownVal(cty.String).Mark("a"),
		},
		{ // propagate collection marks to default when returning
			cty.MapVal(map[string]cty.Value{
				"boop": cty.StringVal("beep"),
			}).Mark("a"),
			cty.StringVal("frob"),
			cty.StringVal("nope").Mark("b"),
			cty.StringVal("nope").WithMarks(cty.NewValueMarks("a", "b")),
		},
		{ // on unmarked collection, return only marks from found value
			cty.MapVal(map[string]cty.Value{
				"boop": cty.StringVal("beep").Mark("a"),
				"frob": cty.StringVal("honk").Mark("b"),
			}),
			cty.StringVal("frob"),
			cty.StringVal("nope").Mark("c"),
			cty.StringVal("honk").Mark("b"),
		},
		{ // on unmarked collection, return default exactly on missing
			cty.MapVal(map[string]cty.Value{
				"boop": cty.StringVal("beep").Mark("a"),
				"frob": cty.StringVal("honk").Mark("b"),
			}),
			cty.StringVal("squish"),
			cty.StringVal("nope").Mark("c"),
			cty.StringVal("nope").Mark("c"),
		},
		{ // retain marks on default if converted
			cty.MapVal(map[string]cty.Value{
				"boop": cty.StringVal("beep").Mark("a"),
				"frob": cty.StringVal("honk").Mark("b"),
			}),
			cty.StringVal("squish"),
			cty.NumberIntVal(5).Mark("c"),
			cty.StringVal("5").Mark("c"),
		},
		{ // propagate marks from key
			cty.MapVal(map[string]cty.Value{
				"boop": cty.StringVal("beep"),
				"frob": cty.StringVal("honk"),
			}),
			cty.StringVal("boop").Mark("a"),
			cty.StringVal("nope"),
			cty.StringVal("beep").Mark("a"),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Lookup(%#v,%#v,%#v)", test.Collection, test.Key, test.Default), func(t *testing.T) {
			got, err := Lookup(test.Collection, test.Key, test.Default)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestElement(t *testing.T) {
	listOfStrings := cty.ListVal([]cty.Value{
		cty.StringVal("the"),
		cty.StringVal("quick"),
		cty.StringVal("brown"),
		cty.StringVal("fox"),
	})
	listOfInts := cty.ListVal([]cty.Value{
		cty.NumberIntVal(1),
		cty.NumberIntVal(2),
		cty.NumberIntVal(3),
		cty.NumberIntVal(4),
	})
	listWithUnknown := cty.ListVal([]cty.Value{
		cty.StringVal("the"),
		cty.StringVal("quick"),
		cty.StringVal("brown"),
		cty.UnknownVal(cty.String),
	})
	listWithMarks := cty.ListVal([]cty.Value{
		cty.StringVal("the"),
		cty.StringVal("quick"),
		cty.StringVal("brown").Mark("fox"),
		cty.UnknownVal(cty.String),
	})

	tests := []struct {
		List  cty.Value
		Index cty.Value
		Want  cty.Value
		Err   bool
	}{
		{
			listOfStrings,
			cty.NumberIntVal(2),
			cty.StringVal("brown"),
			false,
		},
		{ // index greater than length(list)
			listOfStrings,
			cty.NumberIntVal(5),
			cty.StringVal("quick"),
			false,
		},
		{ // list of lists
			cty.ListVal([]cty.Value{listOfStrings, listOfStrings}),
			cty.NumberIntVal(0),
			listOfStrings,
			false,
		},
		{
			listOfStrings,
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.String),
			false,
		},
		{
			listOfInts,
			cty.NumberIntVal(2),
			cty.NumberIntVal(3),
			false,
		},
		{
			listWithUnknown,
			cty.NumberIntVal(2),
			cty.StringVal("brown"),
			false,
		},
		{
			listWithUnknown,
			cty.NumberIntVal(3),
			cty.UnknownVal(cty.String),
			false,
		},
		{ // preserve marks
			listWithMarks,
			cty.NumberIntVal(2),
			cty.StringVal("brown").Mark("fox"),
			false,
		},
		{ // marked items
			listWithMarks,
			cty.NumberIntVal(1),
			cty.StringVal("quick"),
			false,
		},
		{ // The entire list is marked
			listWithMarks.Mark("thewholeshebang"),
			cty.NumberIntVal(2),
			cty.StringVal("brown").WithMarks(cty.NewValueMarks("thewholeshebang", "fox")),
			false,
		},
		{
			listOfStrings,
			cty.NumberIntVal(-1),
			cty.DynamicVal,
			true, // index cannot be a negative number
		},
		{
			listOfStrings,
			cty.StringVal("brown"), // definitely not an index
			cty.DynamicVal,
			true,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Element(%#v,%#v)", test.List, test.Index), func(t *testing.T) {
			got, err := Element(test.List, test.Index)

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

func TestCoalesceList(t *testing.T) {
	tests := map[string]struct {
		Values []cty.Value
		Want   cty.Value
		Err    bool
	}{
		"returns first list if non-empty": {
			[]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("a"),
					cty.StringVal("b"),
				}),
				cty.ListVal([]cty.Value{
					cty.StringVal("c"),
					cty.StringVal("d"),
				}),
			},
			cty.ListVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
			}),
			false,
		},
		"returns second list if first is empty": {
			[]cty.Value{
				cty.ListValEmpty(cty.String),
				cty.ListVal([]cty.Value{
					cty.StringVal("c"),
					cty.StringVal("d"),
				}),
			},
			cty.ListVal([]cty.Value{
				cty.StringVal("c"),
				cty.StringVal("d"),
			}),
			false,
		},
		"return type is dynamic, not unified": {
			[]cty.Value{
				cty.ListValEmpty(cty.String),
				cty.ListVal([]cty.Value{
					cty.NumberIntVal(3),
					cty.NumberIntVal(4),
				}),
			},
			cty.ListVal([]cty.Value{
				cty.NumberIntVal(3),
				cty.NumberIntVal(4),
			}),
			false,
		},
		"works with tuples": {
			[]cty.Value{
				cty.EmptyTupleVal,
				cty.TupleVal([]cty.Value{
					cty.StringVal("c"),
					cty.StringVal("d"),
				}),
			},
			cty.TupleVal([]cty.Value{
				cty.StringVal("c"),
				cty.StringVal("d"),
			}),
			false,
		},
		"unknown arguments": {
			[]cty.Value{
				cty.UnknownVal(cty.List(cty.String)),
				cty.ListVal([]cty.Value{
					cty.StringVal("c"),
					cty.StringVal("d"),
				}),
			},
			cty.DynamicVal,
			false,
		},
		"null arguments": {
			[]cty.Value{
				cty.NullVal(cty.List(cty.String)),
				cty.ListVal([]cty.Value{
					cty.StringVal("c"),
					cty.StringVal("d"),
				}),
			},
			cty.ListVal([]cty.Value{
				cty.StringVal("c"),
				cty.StringVal("d"),
			}),
			false,
		},
		"all null arguments": {
			[]cty.Value{
				cty.NullVal(cty.List(cty.String)),
				cty.NullVal(cty.List(cty.String)),
			},
			cty.NilVal,
			true,
		},
		"invalid arguments": {
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{"a": cty.True}),
				cty.ObjectVal(map[string]cty.Value{"b": cty.False}),
			},
			cty.NilVal,
			true,
		},
		"no arguments": {
			[]cty.Value{},
			cty.NilVal,
			true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := CoalesceList(test.Values...)

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

func TestValues(t *testing.T) {
	tests := []struct {
		Collection cty.Value
		Want       cty.Value
		Err        string
	}{
		{
			cty.MapValEmpty(cty.String),
			cty.ListValEmpty(cty.String),
			``,
		},
		{
			cty.MapValEmpty(cty.String).Mark("a"),
			cty.ListValEmpty(cty.String).Mark("a"),
			``,
		},
		{
			cty.NullVal(cty.Map(cty.String)),
			cty.NilVal,
			`argument must not be null`,
		},
		{
			cty.UnknownVal(cty.Map(cty.String)),
			cty.UnknownVal(cty.List(cty.String)),
			``,
		},
		{
			cty.MapVal(map[string]cty.Value{"hello": cty.StringVal("world")}),
			cty.ListVal([]cty.Value{cty.StringVal("world")}),
			``,
		},
		{ // The map itself is not marked, just an inner element.
			cty.MapVal(map[string]cty.Value{"hello": cty.StringVal("world").Mark("a")}),
			cty.ListVal([]cty.Value{cty.StringVal("world").Mark("a")}),
			``,
		},
		{ // The entire map is marked, so the resulting list is also marked.
			cty.MapVal(map[string]cty.Value{"hello": cty.StringVal("world")}).Mark("a"),
			cty.ListVal([]cty.Value{cty.StringVal("world")}).Mark("a"),
			``,
		},
		{ // Marked both inside and outside.
			cty.MapVal(map[string]cty.Value{"hello": cty.StringVal("world").Mark("a")}).Mark("a"),
			cty.ListVal([]cty.Value{cty.StringVal("world").Mark("a")}).Mark("a"),
			``,
		},
		{
			cty.ObjectVal(map[string]cty.Value{"hello": cty.StringVal("world")}),
			cty.TupleVal([]cty.Value{cty.StringVal("world")}),
			``,
		},
		{
			cty.EmptyObjectVal,
			cty.EmptyTupleVal,
			``,
		},
		{
			cty.EmptyObjectVal.Mark("a"),
			cty.EmptyTupleVal.Mark("a"),
			``,
		},
		{
			cty.NullVal(cty.EmptyObject),
			cty.NilVal,
			`argument must not be null`,
		},
		{
			cty.UnknownVal(cty.EmptyObject),
			cty.UnknownVal(cty.EmptyTuple),
			``,
		},
		{
			cty.UnknownVal(cty.Object(map[string]cty.Type{"a": cty.String})),
			cty.UnknownVal(cty.Tuple([]cty.Type{cty.String})),
			``,
		},
		{ // The object itself is not marked, just an inner attribute value.
			cty.ObjectVal(map[string]cty.Value{"hello": cty.StringVal("world").Mark("a")}),
			cty.TupleVal([]cty.Value{cty.StringVal("world").Mark("a")}),
			``,
		},
		{ // The entire object is marked, so the resulting tuple is also marked.
			cty.ObjectVal(map[string]cty.Value{"hello": cty.StringVal("world")}).Mark("a"),
			cty.TupleVal([]cty.Value{cty.StringVal("world")}).Mark("a"),
			``,
		},
		{ // Marked both inside and outside.
			cty.ObjectVal(map[string]cty.Value{"hello": cty.StringVal("world").Mark("a")}).Mark("a"),
			cty.TupleVal([]cty.Value{cty.StringVal("world").Mark("a")}).Mark("a"),
			``,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Values(%#v)", test.Collection), func(t *testing.T) {
			got, err := Values(test.Collection)
			if test.Err != "" {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				if got, want := err.Error(), test.Err; got != want {
					t.Fatalf("wrong error\ngot:  %s\nwant: %s", got, want)
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

func TestZipMap(t *testing.T) {
	tests := []struct {
		Keys   cty.Value
		Values cty.Value
		Want   cty.Value
		Err    string
	}{
		// Lists of values (map result)
		{
			cty.ListValEmpty(cty.String),
			cty.ListValEmpty(cty.String),
			cty.MapValEmpty(cty.String),
			``,
		},
		{
			cty.ListVal([]cty.Value{cty.StringVal("bleep")}),
			cty.ListVal([]cty.Value{cty.StringVal("bloop")}),
			cty.MapVal(map[string]cty.Value{
				"bleep": cty.StringVal("bloop"),
			}),
			``,
		},
		{
			cty.ListVal([]cty.Value{cty.StringVal("bleep"), cty.StringVal("beep")}),
			cty.ListVal([]cty.Value{cty.StringVal("bloop"), cty.StringVal("boop")}),
			cty.MapVal(map[string]cty.Value{
				"beep":  cty.StringVal("boop"),
				"bleep": cty.StringVal("bloop"),
			}),
			``,
		},
		{
			cty.UnknownVal(cty.List(cty.String)),
			cty.UnknownVal(cty.List(cty.String)),
			cty.UnknownVal(cty.Map(cty.String)),
			``,
		},
		{
			cty.UnknownVal(cty.List(cty.String)),
			cty.ListValEmpty(cty.String),
			cty.UnknownVal(cty.Map(cty.String)),
			``,
		},
		{
			cty.ListValEmpty(cty.String),
			cty.UnknownVal(cty.List(cty.String)),
			cty.UnknownVal(cty.Map(cty.String)),
			``,
		},
		{
			cty.ListVal([]cty.Value{cty.StringVal("bleep")}).Mark("a"),
			cty.ListVal([]cty.Value{cty.StringVal("bloop")}),
			cty.MapVal(map[string]cty.Value{
				"bleep": cty.StringVal("bloop"),
			}).Mark("a"),
			``,
		},
		{
			cty.ListVal([]cty.Value{cty.StringVal("bleep")}),
			cty.ListVal([]cty.Value{cty.StringVal("bloop")}).Mark("b"),
			cty.MapVal(map[string]cty.Value{
				"bleep": cty.StringVal("bloop"),
			}).Mark("b"),
			``,
		},
		{
			cty.ListVal([]cty.Value{cty.StringVal("bleep")}).Mark("a"),
			cty.ListVal([]cty.Value{cty.StringVal("bloop")}).Mark("b"),
			cty.MapVal(map[string]cty.Value{
				"bleep": cty.StringVal("bloop"),
			}).Mark("a").Mark("b"),
			``,
		},
		{
			// cty map keys don't have individual marks, so marks on elements
			// in the keys list aggregate with the resulting map as a whole.
			cty.ListVal([]cty.Value{cty.StringVal("bleep").Mark("a")}),
			cty.ListVal([]cty.Value{cty.StringVal("bloop")}),
			cty.MapVal(map[string]cty.Value{
				"bleep": cty.StringVal("bloop"),
			}).Mark("a"),
			``,
		},
		{
			// cty map _values_ can have individual marks, so individual
			// elements in the values list should have their marks preserved.
			cty.ListVal([]cty.Value{cty.StringVal("bleep")}),
			cty.ListVal([]cty.Value{cty.StringVal("bloop").Mark("a")}),
			cty.MapVal(map[string]cty.Value{
				"bleep": cty.StringVal("bloop").Mark("a"),
			}),
			``,
		},
		{
			cty.ListVal([]cty.Value{cty.StringVal("boop")}),
			cty.ListValEmpty(cty.String),
			cty.NilVal,
			`number of keys (1) does not match number of values (0)`,
		},
		{
			cty.ListValEmpty(cty.String),
			cty.ListVal([]cty.Value{cty.StringVal("boop")}),
			cty.NilVal,
			`number of keys (0) does not match number of values (1)`,
		},

		// Tuple of values (object result)
		{
			cty.ListValEmpty(cty.String),
			cty.EmptyTupleVal,
			cty.EmptyObjectVal,
			``,
		},
		{
			cty.ListVal([]cty.Value{cty.StringVal("bleep")}),
			cty.TupleVal([]cty.Value{cty.StringVal("bloop")}),
			cty.ObjectVal(map[string]cty.Value{
				"bleep": cty.StringVal("bloop"),
			}),
			``,
		},
		{
			cty.ListVal([]cty.Value{cty.StringVal("bleep"), cty.StringVal("beep")}),
			cty.TupleVal([]cty.Value{cty.StringVal("bloop"), cty.StringVal("boop")}),
			cty.ObjectVal(map[string]cty.Value{
				"beep":  cty.StringVal("boop"),
				"bleep": cty.StringVal("bloop"),
			}),
			``,
		},
		{
			cty.UnknownVal(cty.List(cty.String)),
			cty.UnknownVal(cty.EmptyTuple),
			cty.DynamicVal,
			``,
		},
		{
			cty.UnknownVal(cty.List(cty.String)),
			cty.EmptyTupleVal,
			cty.DynamicVal,
			``,
		},
		{
			cty.ListValEmpty(cty.String),
			cty.UnknownVal(cty.EmptyTuple),
			cty.UnknownVal(cty.EmptyObject),
			``,
		},
		{
			cty.ListVal([]cty.Value{cty.StringVal("bleep")}).Mark("a"),
			cty.TupleVal([]cty.Value{cty.StringVal("bloop")}),
			cty.ObjectVal(map[string]cty.Value{
				"bleep": cty.StringVal("bloop"),
			}).Mark("a"),
			``,
		},
		{
			cty.ListVal([]cty.Value{cty.StringVal("bleep")}),
			cty.TupleVal([]cty.Value{cty.StringVal("bloop")}).Mark("b"),
			cty.ObjectVal(map[string]cty.Value{
				"bleep": cty.StringVal("bloop"),
			}).Mark("b"),
			``,
		},
		{
			cty.ListVal([]cty.Value{cty.StringVal("bleep")}).Mark("a"),
			cty.TupleVal([]cty.Value{cty.StringVal("bloop")}).Mark("b"),
			cty.ObjectVal(map[string]cty.Value{
				"bleep": cty.StringVal("bloop"),
			}).Mark("a").Mark("b"),
			``,
		},
		{
			// cty object attributes don't have individual marks, so marks on
			// elements in the keys list aggregate with the resulting object as
			// a whole.
			cty.ListVal([]cty.Value{cty.StringVal("bleep").Mark("a")}),
			cty.TupleVal([]cty.Value{cty.StringVal("bloop")}),
			cty.ObjectVal(map[string]cty.Value{
				"bleep": cty.StringVal("bloop"),
			}).Mark("a"),
			``,
		},
		{
			// cty attribute _values_ can have individual marks, so individual
			// elements in the values list should have their marks preserved.
			cty.ListVal([]cty.Value{cty.StringVal("bleep")}),
			cty.TupleVal([]cty.Value{cty.StringVal("bloop").Mark("a")}),
			cty.ObjectVal(map[string]cty.Value{
				"bleep": cty.StringVal("bloop").Mark("a"),
			}),
			``,
		},
		{
			cty.ListVal([]cty.Value{cty.StringVal("boop")}),
			cty.EmptyTupleVal,
			cty.NilVal,
			`number of keys (1) does not match number of values (0)`,
		},
		{
			cty.ListValEmpty(cty.String),
			cty.TupleVal([]cty.Value{cty.StringVal("boop")}),
			cty.NilVal,
			`number of keys (0) does not match number of values (1)`,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("ZipMap(%#v, %#v)", test.Keys, test.Values), func(t *testing.T) {
			got, err := Zipmap(test.Keys, test.Values)
			if test.Err != "" {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				if got, want := err.Error(), test.Err; got != want {
					t.Fatalf("wrong error\ngot:  %s\nwant: %s", got, want)
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

func TestKeys(t *testing.T) {
	tests := []struct {
		Collection cty.Value
		Want       cty.Value
		Err        string
	}{
		{
			cty.MapValEmpty(cty.String),
			cty.ListValEmpty(cty.String),
			``,
		},
		{
			cty.MapValEmpty(cty.String).Mark("a"),
			cty.ListValEmpty(cty.String).Mark("a"),
			``,
		},
		{
			cty.NullVal(cty.Map(cty.String)),
			cty.NilVal,
			`argument must not be null`,
		},
		{
			cty.MapVal(map[string]cty.Value{"hello": cty.StringVal("world")}),
			cty.ListVal([]cty.Value{cty.StringVal("hello")}),
			``,
		},
		{ // The map itself is not marked, just an inner element.
			cty.MapVal(map[string]cty.Value{"hello": cty.StringVal("world").Mark("a")}),
			cty.ListVal([]cty.Value{cty.StringVal("hello")}),
			``,
		},
		{ // The entire map is marked, so the resulting list is also marked.
			cty.MapVal(map[string]cty.Value{"hello": cty.StringVal("world")}).Mark("a"),
			cty.ListVal([]cty.Value{cty.StringVal("hello")}).Mark("a"),
			``,
		},
		{ // Marked both inside and outside.
			cty.MapVal(map[string]cty.Value{"hello": cty.StringVal("world").Mark("a")}).Mark("a"),
			cty.ListVal([]cty.Value{cty.StringVal("hello")}).Mark("a"),
			``,
		},
		{
			cty.ObjectVal(map[string]cty.Value{"hello": cty.StringVal("world")}),
			cty.TupleVal([]cty.Value{cty.StringVal("hello")}),
			``,
		},
		{
			cty.EmptyObjectVal,
			cty.EmptyTupleVal,
			``,
		},
		{
			cty.EmptyObjectVal.Mark("a"),
			cty.EmptyTupleVal.Mark("a"),
			``,
		},
		{
			cty.NullVal(cty.EmptyObject),
			cty.NilVal,
			`argument must not be null`,
		},
		{
			cty.UnknownVal(cty.EmptyObject),
			cty.EmptyTupleVal,
			``,
		},
		{
			cty.UnknownVal(cty.Object(map[string]cty.Type{"a": cty.String})),
			cty.TupleVal([]cty.Value{cty.StringVal("a")}),
			``,
		},
		{ // The object itself is not marked, just an inner attribute value.
			cty.ObjectVal(map[string]cty.Value{"hello": cty.StringVal("world").Mark("a")}),
			cty.TupleVal([]cty.Value{cty.StringVal("hello")}),
			``,
		},
		{ // The entire object is marked, so the resulting tuple is also marked.
			cty.ObjectVal(map[string]cty.Value{"hello": cty.StringVal("world")}).Mark("a"),
			cty.TupleVal([]cty.Value{cty.StringVal("hello")}).Mark("a"),
			``,
		},
		{ // Marked both inside and outside.
			cty.ObjectVal(map[string]cty.Value{"hello": cty.StringVal("world").Mark("a")}).Mark("a"),
			cty.TupleVal([]cty.Value{cty.StringVal("hello")}).Mark("a"),
			``,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Keys(%#v)", test.Collection), func(t *testing.T) {
			got, err := Keys(test.Collection)
			if test.Err != "" {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				if got, want := err.Error(), test.Err; got != want {
					t.Fatalf("wrong error\ngot:  %s\nwant: %s", got, want)
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

func TestFlatten(t *testing.T) {
	tests := []struct {
		List cty.Value
		Want cty.Value
		Err  string
	}{
		{ // Empty case is easy
			cty.ListValEmpty(cty.String),
			cty.EmptyTupleVal,
			"",
		},
		{ // Lists can contain unknown values
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.UnknownVal(cty.String),
					cty.StringVal("a"),
				}),
				cty.ListVal([]cty.Value{
					cty.UnknownVal(cty.String),
					cty.StringVal("b"),
					cty.UnknownVal(cty.String),
				}),
			}),
			cty.TupleVal([]cty.Value{
				cty.UnknownVal(cty.String),
				cty.StringVal("a"),
				cty.UnknownVal(cty.String),
				cty.StringVal("b"),
				cty.UnknownVal(cty.String),
			}),
			"",
		},
		{ // If the list itself is unknown this is the best we can do
			cty.UnknownVal(cty.List(cty.List(cty.String))),
			cty.UnknownVal(cty.DynamicPseudoType),
			"",
		},
		{ // Type error
			cty.MapValEmpty(cty.String),
			cty.DynamicVal,
			"can only flatten lists, sets and tuples",
		},
		{ // Top-level list marks should carry over
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("a"),
				}),
				cty.ListVal([]cty.Value{
					cty.StringVal("b"),
					cty.StringVal("c"),
				}),
				cty.ListValEmpty(cty.String),
			}).Mark("mark"),
			cty.TupleVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
				cty.StringVal("c"),
			}).Mark("mark"),
			"",
		},
		{ // Inner list marks should apply to the result collection
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("a"),
				}).Mark("first"),
				cty.ListVal([]cty.Value{
					cty.StringVal("b"),
					cty.StringVal("c"),
				}).Mark("second"),
				cty.ListValEmpty(cty.String).Mark("third"),
			}),
			cty.TupleVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
				cty.StringVal("c"),
			}).WithMarks(cty.NewValueMarks("first", "second", "third")),
			"",
		},
		{ // Non-list element marks should be retained on the element only
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("a").Mark("a"),
				}),
				cty.ListVal([]cty.Value{
					cty.StringVal("b").Mark("b"),
					cty.StringVal("c").Mark("b"),
				}),
			}),
			cty.TupleVal([]cty.Value{
				cty.StringVal("a").Mark("a"),
				cty.StringVal("b").Mark("b"),
				cty.StringVal("c").Mark("b"),
			}),
			"",
		},
		{ // Nested unknown lists/sets/tuples should still propagate marks
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{cty.StringVal("a")}).Mark("first"),
				cty.UnknownVal(cty.List(cty.String)).Mark("second"),
				cty.ListVal([]cty.Value{cty.StringVal("c")}).Mark("third"),
			}),
			cty.UnknownVal(cty.DynamicPseudoType).WithMarks(cty.NewValueMarks("first", "second", "third")),
			"",
		},
		{ // Empty marked list retains marks
			cty.ListValEmpty(cty.String).Mark("a"),
			cty.EmptyTupleVal.Mark("a"),
			"",
		},
		{
			cty.ListValEmpty(cty.Number),
			cty.EmptyTupleVal,
			"",
		},
		{
			cty.ListVal([]cty.Value{
				cty.DynamicVal,
			}),
			cty.DynamicVal,
			"",
		},
		{
			cty.TupleVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.ListVal([]cty.Value{
						cty.DynamicVal,
					}),
				}),
				cty.ListVal([]cty.Value{
					cty.ListVal([]cty.Value{
						cty.DynamicVal,
					}).Mark("marked"),
				}),
			}),
			cty.DynamicVal.Mark("marked"),
			"",
		},
		{
			cty.TupleVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"blop": cty.ListVal([]cty.Value{
							cty.DynamicVal,
						}),
					}),
				}),
				cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bloop": cty.DynamicVal,
					}),
				}),
			}),
			cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"blop": cty.ListVal([]cty.Value{
						cty.DynamicVal,
					}),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"bloop": cty.DynamicVal,
				}),
			}),
			"",
		},
		{
			cty.ListVal([]cty.Value{
				cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bloop": cty.DynamicVal,
					}),
				}),
				cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bloop": cty.DynamicVal,
					}),
				}),
			}),
			cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"bloop": cty.DynamicVal,
				}),
				cty.ObjectVal(map[string]cty.Value{
					"bloop": cty.DynamicVal,
				}),
			}),
			"",
		},
		{
			cty.TupleVal([]cty.Value{
				cty.StringVal("a"),
				cty.ListVal([]cty.Value{
					cty.StringVal("b"),
				}),
				cty.TupleVal([]cty.Value{
					cty.ListVal([]cty.Value{
						cty.StringVal("c"),
					}),
					cty.ListVal([]cty.Value{
						cty.StringVal("d"),
						cty.StringVal("e"),
					}),
				}),
			}),
			cty.TupleVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
				cty.StringVal("c"),
				cty.StringVal("d"),
				cty.StringVal("e"),
			}),
			"",
		},
		{
			cty.TupleVal([]cty.Value{
				cty.TupleVal([]cty.Value{
					cty.StringVal("a"),
					cty.StringVal("b"),
				}),
				cty.NullVal(cty.DynamicPseudoType),
				cty.TupleVal([]cty.Value{
					cty.StringVal("c"),
				}),
			}),
			cty.TupleVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
				cty.NullVal(cty.DynamicPseudoType),
				cty.StringVal("c"),
			}),
			"",
		},
		{
			cty.TupleVal([]cty.Value{
				cty.TupleVal([]cty.Value{
					cty.StringVal("a"),
					cty.StringVal("b"),
				}),
				cty.DynamicVal,
				cty.TupleVal([]cty.Value{
					cty.StringVal("c"),
				}),
			}),
			cty.UnknownVal(cty.DynamicPseudoType),
			"",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Flatten(%#v)", test.List), func(t *testing.T) {
			got, err := Flatten(test.List)
			if test.Err != "" {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				if got, want := err.Error(), test.Err; got != want {
					t.Fatalf("wrong error\ngot:  %s\nwant: %s", got, want)
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

func TestSetproduct(t *testing.T) {
	tests := []struct {
		Collections []cty.Value
		Want        cty.Value
		Err         string
	}{
		{
			[]cty.Value{cty.ListValEmpty(cty.String)},
			cty.NilVal,
			`at least two arguments are required`,
		},
		{
			[]cty.Value{
				cty.ListValEmpty(cty.EmptyObject),
				cty.ListVal([]cty.Value{
					cty.StringVal("quick"),
					cty.StringVal("fox"),
				}),
			},
			cty.ListValEmpty(cty.Tuple([]cty.Type{cty.EmptyObject, cty.String})),
			``,
		},
		{
			[]cty.Value{
				cty.SetValEmpty(cty.EmptyObject),
				cty.SetVal([]cty.Value{
					cty.StringVal("quick"),
					cty.StringVal("fox"),
				}),
			},
			cty.SetValEmpty(cty.Tuple([]cty.Type{cty.EmptyObject, cty.String})),
			``,
		},
		{
			[]cty.Value{
				cty.ListValEmpty(cty.EmptyObject),
				cty.ListValEmpty(cty.EmptyObject),
			},
			cty.ListValEmpty(cty.Tuple([]cty.Type{cty.EmptyObject, cty.EmptyObject})),
			``,
		},
		{
			[]cty.Value{
				cty.SetValEmpty(cty.EmptyObject),
				cty.SetValEmpty(cty.EmptyObject),
			},
			cty.SetValEmpty(cty.Tuple([]cty.Type{cty.EmptyObject, cty.EmptyObject})),
			``,
		},
		{
			[]cty.Value{
				cty.ListVal([]cty.Value{cty.ListValEmpty(cty.String)}),
				cty.ListVal([]cty.Value{cty.ListValEmpty(cty.String)}),
			},
			cty.ListVal([]cty.Value{cty.TupleVal([]cty.Value{cty.ListValEmpty(cty.String), cty.ListValEmpty(cty.String)})}),
			``,
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.ListValEmpty(cty.String)}),
				cty.SetVal([]cty.Value{cty.ListValEmpty(cty.String)}),
			},
			cty.SetVal([]cty.Value{cty.TupleVal([]cty.Value{cty.ListValEmpty(cty.String), cty.ListValEmpty(cty.String)})}),
			``,
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.ListValEmpty(cty.String).Mark("a")}),
				cty.SetVal([]cty.Value{cty.ListValEmpty(cty.String)}),
			},
			cty.SetVal([]cty.Value{cty.TupleVal([]cty.Value{cty.ListValEmpty(cty.String).Mark("a"), cty.ListValEmpty(cty.String)})}),
			``,
		},
		{
			[]cty.Value{
				cty.TupleVal([]cty.Value{
					cty.StringVal("the"),
					cty.StringVal("brown"),
				}),
				cty.TupleVal([]cty.Value{
					cty.StringVal("fox"),
					cty.NumberIntVal(3),
				}),
			},
			cty.ListVal([]cty.Value{
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("fox")}),
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("3")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown"), cty.StringVal("fox")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown"), cty.StringVal("3")}),
			}),
			``,
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{
					cty.StringVal("the"),
					cty.StringVal("brown"),
				}),
				cty.SetVal([]cty.Value{
					cty.StringVal("quick"),
					cty.StringVal("fox"),
				}),
			},
			cty.SetVal([]cty.Value{
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("fox")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown"), cty.StringVal("fox")}),
			}),
			``,
		},
		{ // The collection itself is not marked, just some elements
			[]cty.Value{
				cty.SetVal([]cty.Value{
					cty.StringVal("the"),
					cty.StringVal("brown").Mark("a"),
				}),
				cty.SetVal([]cty.Value{
					cty.StringVal("quick"),
					cty.StringVal("fox").Mark("b"),
				}),
			},
			// Sets don't allow individually-marked elements, so the marks
			// end up aggregating on the set itself anyway in this case.
			cty.SetVal([]cty.Value{
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("fox")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown"), cty.StringVal("fox")}),
			}).Mark("a").Mark("b"),
			``,
		},
		{ // The collections are marked
			[]cty.Value{
				cty.SetVal([]cty.Value{
					cty.StringVal("the"),
					cty.StringVal("brown"),
				}).Mark("a"),
				cty.SetVal([]cty.Value{
					cty.StringVal("quick"),
					cty.StringVal("fox"),
				}).Mark("b"),
			},
			cty.SetVal([]cty.Value{
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("fox")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown"), cty.StringVal("fox")}),
			}).Mark("a").Mark("b"),
			``,
		},
		{ // One collection is marked
			[]cty.Value{
				cty.SetVal([]cty.Value{
					cty.StringVal("the"),
					cty.StringVal("brown"),
				}).Mark("a"),
				cty.SetVal([]cty.Value{
					cty.StringVal("quick"),
					cty.StringVal("fox"),
				}),
			},
			cty.SetVal([]cty.Value{
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("fox")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown"), cty.StringVal("fox")}),
			}).Mark("a"),
			``,
		},
		{ // Inner and outer marks
			[]cty.Value{
				cty.SetVal([]cty.Value{
					cty.StringVal("the"),
					cty.StringVal("brown").Mark("a"),
				}).Mark("b"),
				cty.SetVal([]cty.Value{
					cty.StringVal("quick"),
					cty.StringVal("fox").Mark("c"),
				}),
			},
			cty.SetVal([]cty.Value{
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("fox")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown"), cty.StringVal("fox")}),
			}).WithMarks(cty.NewValueMarks("b", "c", "a")),
			``,
		},

		// SetproductFunc supports lists too, in which case it preserves the
		// input order and returns a list as the result. In this case we can
		// preserve the marks more precisely.
		{ // The collection itself is not marked, just some elements
			[]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("the"),
					cty.StringVal("brown").Mark("a"),
				}),
				cty.ListVal([]cty.Value{
					cty.StringVal("quick"),
					cty.StringVal("fox").Mark("b"),
				}),
			},
			cty.ListVal([]cty.Value{
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("fox").Mark("b")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown").Mark("a"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown").Mark("a"), cty.StringVal("fox").Mark("b")}),
			}),
			``,
		},
		{ // The collections are marked
			[]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("the"),
					cty.StringVal("brown"),
				}).Mark("a"),
				cty.ListVal([]cty.Value{
					cty.StringVal("quick"),
					cty.StringVal("fox"),
				}).Mark("b"),
			},
			cty.ListVal([]cty.Value{
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("fox")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown"), cty.StringVal("fox")}),
			}).Mark("a").Mark("b"),
			``,
		},
		{ // One collection is marked
			[]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("the"),
					cty.StringVal("brown"),
				}).Mark("a"),
				cty.ListVal([]cty.Value{
					cty.StringVal("quick"),
					cty.StringVal("fox"),
				}),
			},
			cty.ListVal([]cty.Value{
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("fox")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown"), cty.StringVal("fox")}),
			}).Mark("a"),
			``,
		},
		{ // Inner and outer marks
			[]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("the"),
					cty.StringVal("brown").Mark("a"),
				}).Mark("b"),
				cty.ListVal([]cty.Value{
					cty.StringVal("quick"),
					cty.StringVal("fox").Mark("c"),
				}),
			},
			cty.ListVal([]cty.Value{
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("the"), cty.StringVal("fox").Mark("c")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown").Mark("a"), cty.StringVal("quick")}),
				cty.TupleVal([]cty.Value{cty.StringVal("brown").Mark("a"), cty.StringVal("fox").Mark("c")}),
			}).Mark("b"),
			``,
		},
		{
			// Empty lists with marks should propagate the marks
			[]cty.Value{
				cty.ListValEmpty(cty.String).Mark("a"),
				cty.ListValEmpty(cty.Bool).Mark("b"),
			},
			cty.ListValEmpty(cty.Tuple([]cty.Type{cty.String, cty.Bool})).WithMarks(cty.NewValueMarks("a", "b")),
			``,
		},
		{
			// Empty sets with marks should propagate the marks
			[]cty.Value{
				cty.SetValEmpty(cty.String).Mark("a"),
				cty.SetValEmpty(cty.Bool).Mark("b"),
			},
			cty.SetValEmpty(cty.Tuple([]cty.Type{cty.String, cty.Bool})).WithMarks(cty.NewValueMarks("a", "b")),
			``,
		},
		{
			// Arguments which are sets with partially unknown values results
			// in unknown length (since the unknown values may already be
			// present in the set). This gives an unknown result preserving all
			// marks
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.StringVal("x"), cty.UnknownVal(cty.String)}).Mark("a"),
				cty.SetVal([]cty.Value{cty.True, cty.False}).Mark("b"),
			},
			cty.UnknownVal(cty.Set(cty.Tuple([]cty.Type{cty.String, cty.Bool}))).WithMarks(cty.NewValueMarks("a", "b")),
			``,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Setproduct(%#v)", test.Collections), func(t *testing.T) {
			got, err := SetProduct(test.Collections...)
			if test.Err != "" {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				if got, want := err.Error(), test.Err; got != want {
					t.Fatalf("wrong error\ngot:  %s\nwant: %s", got, want)
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

func TestReverseList(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
		Err   string
	}{
		{
			cty.NilVal,
			cty.NilVal,
			`argument must not be null`,
		},
		{
			cty.ListValEmpty(cty.String),
			cty.ListValEmpty(cty.String),
			``,
		},
		{
			cty.ListValEmpty(cty.String).Mark("foo"),
			cty.ListValEmpty(cty.String).Mark("foo"),
			``,
		},
		{
			cty.UnknownVal(cty.List(cty.String)),
			cty.UnknownVal(cty.List(cty.String)),
			``,
		},
		{ // marks on list elements
			cty.ListVal([]cty.Value{
				cty.StringVal("beep").Mark("boop"),
				cty.StringVal("bop"),
				cty.StringVal("bloop"),
			}),
			cty.ListVal([]cty.Value{
				cty.StringVal("bloop"),
				cty.StringVal("bop"),
				cty.StringVal("beep").Mark("boop"),
			}),
			``,
		},
		{ // marks on the entire input are preserved
			cty.ListVal([]cty.Value{
				cty.StringVal("beep").Mark("boop"),
				cty.StringVal("bop"),
				cty.StringVal("bloop"),
			}).Mark("outer"),
			cty.ListVal([]cty.Value{
				cty.StringVal("bloop"),
				cty.StringVal("bop"),
				cty.StringVal("beep").Mark("boop"),
			}).Mark("outer"),
			``,
		},
		{ // marks on tuple elements
			cty.TupleVal([]cty.Value{
				cty.StringVal("beep").Mark("boop"),
				cty.StringVal("bop"),
				cty.StringVal("bloop"),
			}),
			cty.TupleVal([]cty.Value{
				cty.StringVal("bloop"),
				cty.StringVal("bop"),
				cty.StringVal("beep").Mark("boop"),
			}),
			``,
		},
		{ // Set elements don't support individual marks; any marks on elements get propegated to the entire set.
			cty.SetVal([]cty.Value{
				cty.StringVal("beep").Mark("boop"),
				cty.StringVal("bop"),
				cty.StringVal("bloop"),
			}),
			// sets end up sorted alphabetically when converted to lists
			cty.ListVal([]cty.Value{
				cty.StringVal("bop"),
				cty.StringVal("bloop"),
				cty.StringVal("beep"),
			}).Mark("boop"),
			``,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("ReverseList(%#v)", test.Input), func(t *testing.T) {
			got, err := ReverseList(test.Input)
			if test.Err != "" {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				if got, want := err.Error(), test.Err; got != want {
					t.Fatalf("wrong error\ngot:  %s\nwant: %s", got, want)
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

func TestSlice(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Start cty.Value
		End   cty.Value
		Want  cty.Value
		Err   string
	}{
		{
			Input: cty.ListVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
				cty.StringVal("c"),
			}),
			Start: cty.NumberIntVal(0),
			End:   cty.NumberIntVal(2),
			Want: cty.ListVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
			}),
			Err: ``,
		},
		{ // The entire input list is marked, so the return should be marked
			Input: cty.ListVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
				cty.StringVal("c"),
			}).Mark("bloop"),
			Start: cty.NumberIntVal(0),
			End:   cty.NumberIntVal(2),
			Want: cty.ListVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
			}).Mark("bloop"),
			Err: ``,
		},
		{ // individual element marks should be preserved
			Input: cty.ListVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b").Mark("bloop"),
				cty.StringVal("c"),
			}),
			Start: cty.NumberIntVal(0),
			End:   cty.NumberIntVal(2),
			Want: cty.ListVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b").Mark("bloop"),
			}),
			Err: ``,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Slice(%#v)", test.Input), func(t *testing.T) {
			got, err := Slice(test.Input, test.Start, test.End)
			if test.Err != "" {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				if got, want := err.Error(), test.Err; got != want {
					t.Fatalf("wrong error\ngot:  %s\nwant: %s", got, want)
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
