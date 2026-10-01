package cty

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestSetVal(t *testing.T) {
	plain := SetVal([]Value{True})
	marked := SetVal([]Value{True}).Mark(1)
	deepMarked := SetVal([]Value{True.Mark(2), True.Mark(3)})

	if plain.RawEquals(marked) {
		t.Errorf("plain should be unequal to marked\nplain:  %#v\nmarked: %#v", plain, marked)
	}
	if marked.RawEquals(deepMarked) {
		t.Errorf("marked should be unequal to deepMarked\nmarked:      %#v\ndeepmarked: %#v", marked, deepMarked)
	}
	if got, want := marked.Marks(), NewValueMarks(1); !got.Equal(want) {
		t.Errorf("wrong marks for marked\ngot:  %#v\nwant: %#v", got, want)
	}
	if got, want := deepMarked.Marks(), NewValueMarks(2, 3); !got.Equal(want) {
		// Both 2 and 3 marks are preserved even though both of them are
		// marking the same value True, and thus the resulting set contains
		// only one element.
		t.Errorf("wrong marks for deepMarked\ngot:  %#v\nwant: %#v", got, want)
	}

	if got, want := deepMarked.unmarkForce(), SetVal([]Value{True}); !got.RawEquals(want) {
		t.Errorf("wrong unmarked value for deepMarked\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestSetVal_nestedStructures(t *testing.T) {
	testCases := []struct {
		Name  string
		Elems []Value
	}{
		{
			"set",
			[]Value{
				SetVal([]Value{
					NumberIntVal(5),
				}),
			},
		},
		{
			"doubly nested set",
			[]Value{
				SetVal([]Value{
					SetVal([]Value{
						NumberIntVal(5),
					}),
				}),
			},
		},
		{
			"list",
			[]Value{
				ListVal([]Value{
					NumberIntVal(5),
				}),
			},
		},
		{
			"doubly nested list",
			[]Value{
				ListVal([]Value{
					ListVal([]Value{
						NumberIntVal(5),
					}),
				}),
			},
		},
		{
			"map",
			[]Value{
				MapVal(map[string]Value{
					"key": NumberIntVal(5),
				}),
			},
		},
		{
			"doubly nested map",
			[]Value{
				MapVal(map[string]Value{
					"key": MapVal(map[string]Value{
						"child": StringVal("hello world"),
					}),
				}),
			},
		},
		{
			"tuple",
			[]Value{
				TupleVal([]Value{
					NumberIntVal(5),
				}),
			},
		},
		{
			"doubly nested tuple",
			[]Value{
				TupleVal([]Value{
					TupleVal([]Value{
						NumberIntVal(5),
					}),
				}),
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.Name), func(t *testing.T) {
			SetVal(tc.Elems)
		})
	}
}

func TestCanListVal(t *testing.T) {
	testCases := []struct {
		Elems []Value
		Want  bool
	}{
		// Valid lists
		{
			[]Value{StringVal("Hello"), StringVal("World")},
			true,
		},
		{
			[]Value{NumberIntVal(13), NumberIntVal(31)},
			true,
		},
		{
			[]Value{BoolVal(true), BoolVal(false)},
			true,
		},
		{
			[]Value{
				ListVal([]Value{
					StringVal("Hello"), StringVal("World"),
				}),
				ListVal([]Value{
					StringVal("beep"), StringVal("boop"), StringVal("bloop"),
				}),
			},
			true,
		},
		{
			[]Value{
				MapVal(map[string]Value{
					"a": StringVal("Hello"),
				}),
				MapVal(map[string]Value{
					"c": StringVal("World"),
				}),
			},
			true,
		},
		{
			[]Value{
				SetVal([]Value{
					StringVal("Hello"), StringVal("World"),
				}),
				SetVal([]Value{
					StringVal("beep"), StringVal("boop"), StringVal("bloop"),
				}),
			},
			true,
		},
		// invalid list elements
		{
			[]Value{StringVal("hello"), NumberIntVal(13)},
			false,
		},
		{
			[]Value{
				ListVal([]Value{
					StringVal("Hello"), StringVal("World"),
				}),
				MapVal(map[string]Value{
					"a": StringVal("bloop"),
				}),
			},
			false,
		},
		{ // List of string and List of lists
			[]Value{
				ListVal([]Value{
					StringVal("Hello"), StringVal("World"),
				}),
				ListVal([]Value{
					ListVal([]Value{
						StringVal("a"), StringVal("b"),
					}),
					ListVal([]Value{
						StringVal("c"), StringVal("d"),
					}),
				}),
			},
			false,
		},
		{ // Inconsistent map elements
			[]Value{
				MapVal(map[string]Value{
					"a": StringVal("Hello"),
				}),
				MapVal(map[string]Value{
					"a": BoolVal(true),
				}),
			},
			false,
		},
	}

	for _, tc := range testCases {
		got := CanListVal(tc.Elems)
		if got != tc.Want {
			t.Errorf("wrong result for elements %#v:\ngot %v, want %v", tc.Elems, got, tc.Want)
		}
	}
}

func TestCanSetVal(t *testing.T) {
	testCases := []struct {
		Elems []Value
		Want  bool
	}{
		// Valid set elements
		{
			[]Value{StringVal("Hello"), StringVal("World")},
			true,
		},
		{
			[]Value{StringVal("Hello").Mark(1), StringVal("World").Mark(2)},
			true,
		},
		{
			[]Value{NumberIntVal(13), NumberIntVal(31)},
			true,
		},
		{
			[]Value{BoolVal(true), BoolVal(false)},
			true,
		},
		{
			[]Value{
				ListVal([]Value{
					StringVal("Hello"), StringVal("World"),
				}),
				ListVal([]Value{
					StringVal("beep"), StringVal("boop"), StringVal("bloop"),
				}),
			},
			true,
		},
		{
			[]Value{
				MapVal(map[string]Value{
					"a": StringVal("Hello"),
				}),
				MapVal(map[string]Value{
					"c": StringVal("World"),
				}),
			},
			true,
		},
		{
			[]Value{
				SetVal([]Value{
					StringVal("Hello"), StringVal("World"),
				}),
				SetVal([]Value{
					StringVal("beep"), StringVal("boop"), StringVal("bloop"),
				}),
			},
			true,
		},
		// invalid set elements
		{
			[]Value{StringVal("hello"), NumberIntVal(13)},
			false,
		},
		{
			[]Value{
				ListVal([]Value{
					StringVal("Hello"), StringVal("World"),
				}),
				MapVal(map[string]Value{
					"a": StringVal("bloop"),
				}),
			},
			false,
		},
		{ // List of string and List of lists
			[]Value{
				ListVal([]Value{
					StringVal("Hello"), StringVal("World"),
				}),
				ListVal([]Value{
					ListVal([]Value{
						StringVal("a"), StringVal("b"),
					}),
					ListVal([]Value{
						StringVal("c"), StringVal("d"),
					}),
				}),
			},
			false,
		},
		{ // Inconsistent map elements
			[]Value{
				MapVal(map[string]Value{
					"a": StringVal("Hello"),
				}),
				MapVal(map[string]Value{
					"a": BoolVal(true),
				}),
			},
			false,
		},
	}

	for _, tc := range testCases {
		got := CanSetVal(tc.Elems)
		if got != tc.Want {
			t.Errorf("wrong result for elements %#v:\ngot %v, want %v", tc.Elems, got, tc.Want)
		}
	}
}

func TestCanMapVal(t *testing.T) {
	testCases := []struct {
		Elems map[string]Value
		Want  bool
	}{
		// Valid lists
		{
			map[string]Value{"a": StringVal("Hello"), "b": StringVal("World")},
			true,
		},
		{
			map[string]Value{"one": NumberIntVal(13), "two": NumberIntVal(31)},
			true,
		},
		{
			map[string]Value{"one": BoolVal(true), "two": BoolVal(false)},
			true,
		},
		{
			map[string]Value{
				"lista": ListVal([]Value{
					StringVal("Hello"), StringVal("World"),
				}),
				"listb": ListVal([]Value{
					StringVal("beep"), StringVal("boop"), StringVal("bloop"),
				}),
			},
			true,
		},
		{
			map[string]Value{
				"map_a": MapVal(map[string]Value{
					"a": StringVal("Hello"),
				}),
				"map_b": MapVal(map[string]Value{
					"c": StringVal("World"),
				}),
			},
			true,
		},
		{
			map[string]Value{
				"set_a": SetVal([]Value{
					StringVal("Hello"), StringVal("World"),
				}),
				"set_b": SetVal([]Value{
					StringVal("beep"), StringVal("boop"), StringVal("bloop"),
				}),
			},
			true,
		},
		// invalid map elements
		{
			map[string]Value{"one": StringVal("hello"), "two": NumberIntVal(13)},
			false,
		},
		{
			map[string]Value{
				"one": ListVal([]Value{
					StringVal("Hello"), StringVal("World"),
				}),
				"two": MapVal(map[string]Value{
					"a": StringVal("bloop"),
				}),
			},
			false,
		},
		{
			map[string]Value{
				"one": ListVal([]Value{
					StringVal("Hello"), StringVal("World"),
				}),
				"two": ListVal([]Value{
					ListVal([]Value{
						StringVal("a"), StringVal("b"),
					}),
					ListVal([]Value{
						StringVal("c"), StringVal("d"),
					}),
				}),
			},
			false,
		},
		{ // Inconsistent map elements
			map[string]Value{
				"one": MapVal(map[string]Value{
					"a": StringVal("Hello"),
				}),
				"two": MapVal(map[string]Value{
					"a": BoolVal(true),
				}),
			},
			false,
		},
	}

	for _, tc := range testCases {
		got := CanMapVal(tc.Elems)
		if got != tc.Want {
			t.Errorf("wrong result for elements %#v:\ngot %v, want %v", tc.Elems, got, tc.Want)
		}
	}
}

func TestUnknownChoice(t *testing.T) {
	tests := []struct {
		Inputs []Value
		Want   Value
	}{
		{
			nil,
			DynamicVal,
		},

		{
			[]Value{
				True,
			},
			True,
		},
		{
			[]Value{
				True,
				True,
			},
			True,
		},
		{
			[]Value{
				True,
				True,
				True,
			},
			True,
		},
		{
			[]Value{
				True,
				False,
			},
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			[]Value{
				True,
				NullVal(Bool),
			},
			UnknownVal(Bool),
		},
		{
			[]Value{
				NullVal(Bool),
				NullVal(Bool),
			},
			NullVal(Bool),
		},
		{
			[]Value{
				True,
				Zero,
			},
			DynamicVal,
		},
		{
			[]Value{
				True.Mark("a"),
			},
			True.Mark("a"),
		},
		{
			[]Value{
				True.Mark("a"),
				True.Mark("a"),
			},
			True.Mark("a"),
		},
		{
			[]Value{
				True.Mark("a"),
				True.Mark("b"),
			},
			UnknownVal(Bool).RefineNotNull().Mark("a").Mark("b"),
		},
		{
			[]Value{
				True.Mark("a"),
				True.Mark("b"),
				True.Mark("c"),
			},
			UnknownVal(Bool).RefineNotNull().Mark("a").Mark("b").Mark("c"),
		},
		{
			[]Value{
				True.Mark("a"),
				False.Mark("b"),
			},
			UnknownVal(Bool).RefineNotNull().Mark("a").Mark("b"),
		},
		{
			[]Value{
				True.Mark("a"),
				NullVal(Bool).Mark("b"),
			},
			UnknownVal(Bool).Mark("a").Mark("b"),
		},
		{
			[]Value{
				NullVal(Bool).Mark("a"),
				NullVal(Bool).Mark("b"),
			},
			NullVal(Bool).Mark("a").Mark("b"),
		},
		{
			[]Value{
				True.Mark("a"),
				Zero.Mark("b"),
			},
			DynamicVal.Mark("a").Mark("b"),
		},
		{
			[]Value{
				StringVal("a"),
				StringVal("b"),
			},
			UnknownVal(String).RefineNotNull(),
		},
		{
			[]Value{
				StringVal("boop:1"),
				StringVal("boop:2"),
			},
			UnknownVal(String).Refine().
				NotNull().
				StringPrefixFull("boop:").
				NewValue(),
		},
		{
			[]Value{
				StringVal("boop:1").Mark("a"),
				StringVal("boop:2").Mark("b"),
			},
			UnknownVal(String).Refine().
				NotNull().
				StringPrefixFull("boop:").
				NewValue().
				Mark("a").Mark("b"),
		},
		{
			[]Value{
				StringVal("boop").Mark("a"),
				StringVal("boosh").Mark("b"),
			},
			UnknownVal(String).Refine().
				NotNull().
				StringPrefixFull("boo").
				NewValue().
				Mark("a").Mark("b"),
		},
		{
			[]Value{
				ListValEmpty(String).Mark("a"),
				ListValEmpty(String).Mark("b"),
			},
			ListValEmpty(String).Mark("a").Mark("b"),
		},
		{
			[]Value{
				ListVal([]Value{True}),
				ListVal([]Value{True}),
			},
			ListVal([]Value{True}),
		},
		{
			[]Value{
				ListVal([]Value{True}),
				ListVal([]Value{False}),
			},
			ListVal([]Value{UnknownVal(Bool)}),
		},
		{
			[]Value{
				ListVal([]Value{True}).Mark("a"),
				ListVal([]Value{True}).Mark("a"),
			},
			ListVal([]Value{True}).Mark("a"),
		},
		{
			[]Value{
				ListVal([]Value{True}).Mark("a"),
				ListVal([]Value{True}).Mark("b"),
			},
			ListVal([]Value{UnknownVal(Bool)}).Mark("a").Mark("b"),
		},
		{
			[]Value{
				ListVal([]Value{True}).Mark("a"),
				ListVal([]Value{False}).Mark("b"),
			},
			ListVal([]Value{UnknownVal(Bool)}).Mark("a").Mark("b"),
		},
		{
			[]Value{
				ListValEmpty(Bool),
				ListVal([]Value{True}),
				ListVal([]Value{True, True}),
			},
			UnknownVal(List(Bool)).Refine().
				NotNull().
				CollectionLengthLowerBound(0).
				CollectionLengthUpperBound(2).
				NewValue(),
		},
		{
			[]Value{
				ListVal([]Value{True}),
				ListVal([]Value{True, True}),
			},
			UnknownVal(List(Bool)).Refine().
				NotNull().
				CollectionLengthLowerBound(1).
				CollectionLengthUpperBound(2).
				NewValue(),
		},
		{
			[]Value{
				ListVal([]Value{True}).Mark("a"),
				ListVal([]Value{True, True}).Mark("b"),
			},
			UnknownVal(List(Bool)).Refine().
				NotNull().
				CollectionLengthLowerBound(1).
				CollectionLengthUpperBound(2).
				NewValue().
				Mark("a").Mark("b"),
		},
		{
			[]Value{
				SetVal([]Value{True}),
				SetVal([]Value{False}),
			},
			SetVal([]Value{UnknownVal(Bool)}),
		},
		{
			[]Value{
				SetVal([]Value{True}),
				ListVal([]Value{False}),
			},
			DynamicVal,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.Inputs), func(t *testing.T) {
			got := UnknownChoice(slices.Values(test.Inputs))
			if !test.Want.RawEquals(got) {
				var buf strings.Builder
				for _, input := range test.Inputs {
					fmt.Fprintf(&buf, " - %#v\n", input)
				}
				t.Errorf("wrong result\ninputs:\n%s\ngot:  %#v\nwant: %#v", buf.String(), got, test.Want)
				if got.GoString() == test.Want.GoString() {
					// We should not generally get in here, but since this test
					// and the code it's testing involve a bunch of
					// manually-assembled values rather than going through the
					// public API it's possible that we'll accidentally
					// construct something that the normal API couldn't normally
					// make and thus have differences the GoString function
					// doesn't take into account, in which case we'll show the
					// internals too for easier debugging.
					t.Logf(
						"value internals\ngot: cty.Value{\n  v:  %#v,\n  ty: %#v,\n}\nwant: cty.Value{\n  v:  %#v,\n  ty: %#v,\n}",
						got.v, got.ty, test.Want.v, test.Want.ty,
					)
				}
			}
		})
	}
}
