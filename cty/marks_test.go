package cty

import (
	"fmt"
	"testing"
)

func TestContainsMarked(t *testing.T) {
	testCases := []struct {
		val  Value
		want bool
	}{
		{
			StringVal("a"),
			false,
		},
		{
			NumberIntVal(1).Mark("a"),
			true,
		},
		{
			ListVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			false,
		},
		{
			ListVal([]Value{NumberIntVal(1), NumberIntVal(2).Mark("a")}),
			true,
		},
		{
			ListVal([]Value{NumberIntVal(1), NumberIntVal(2)}).Mark("a"),
			true,
		},
		{
			ListValEmpty(String).Mark("c"),
			true,
		},
		{
			MapVal(map[string]Value{"a": StringVal("b").Mark("c"), "x": StringVal("y").Mark("z")}),
			true,
		},
		{
			TupleVal([]Value{NumberIntVal(1).Mark("a"), StringVal("y").Mark("z")}),
			true,
		},
		{
			SetVal([]Value{NumberIntVal(1).Mark("a"), NumberIntVal(2).Mark("z")}),
			true,
		},
		{
			ObjectVal(map[string]Value{
				"x": ListVal([]Value{
					NumberIntVal(1).Mark("a"),
					NumberIntVal(2),
				}),
				"y": StringVal("y"),
				"z": BoolVal(true),
			}),
			true,
		},
	}

	for _, tc := range testCases {
		if got, want := tc.val.ContainsMarked(), tc.want; got != want {
			t.Errorf("wrong result (got %v, want %v) for %#v", got, want, tc.val)
		}
	}
}

func TestIsMarked(t *testing.T) {
	testCases := []struct {
		val  Value
		want bool
	}{
		{
			StringVal("a"),
			false,
		},
		{
			NumberIntVal(1).Mark("a"),
			true,
		},
		{
			ListVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			false,
		},
		{
			ListVal([]Value{NumberIntVal(1), NumberIntVal(2).Mark("a")}),
			false,
		},
		{
			ListVal([]Value{NumberIntVal(1), NumberIntVal(2)}).Mark("a"),
			true,
		},
	}

	for _, tc := range testCases {
		if got, want := tc.val.IsMarked(), tc.want; got != want {
			t.Errorf("wrong result (got %v, want %v) for %#v", got, want, tc.val)
		}
	}
}

func TestValueMarks(t *testing.T) {
	v := True
	v1 := v.Mark(1)
	v2 := v.Mark(2)

	if got, want := v.Marks(), NewValueMarks(); !want.Equal(got) {
		t.Errorf("wrong v marks\ngot:  %#v\nwant: %#v", got, want)
	}
	if got, want := v1.Marks(), NewValueMarks(1); !want.Equal(got) {
		t.Errorf("wrong v1 marks\ngot:  %#v\nwant: %#v", got, want)
	}
	if got, want := v2.Marks(), NewValueMarks(2); !want.Equal(got) {
		t.Errorf("wrong v2 marks\ngot:  %#v\nwant: %#v", got, want)
	}

	v12 := False.WithSameMarks(v, v1, v2)
	if got, want := v12.Marks(), NewValueMarks(1, 2); !want.Equal(got) {
		t.Errorf("wrong v12 marks\ngot:  %#v\nwant: %#v", got, want)
	}

	v12Again := v12.Mark(1)
	if got, want := v12Again.Marks(), NewValueMarks(1, 2); !want.Equal(got) {
		t.Errorf("wrong v12Again marks\ngot:  %#v\nwant: %#v", got, want)
	}

	v1234 := v12.WithMarks(NewValueMarks(2, 3, 4))
	if got, want := v1234.Marks(), NewValueMarks(1, 2, 3, 4); !want.Equal(got) {
		t.Errorf("wrong v1234 marks\ngot:  %#v\nwant: %#v", got, want)
	}
	if !v1234.HasMark(2) {
		t.Errorf("v1234 should have mark 2")
	}
	if v1234.HasMark(5) {
		t.Errorf("v1234 should not have mark 5")
	}

	v, marks1234 := v1234.Unmark()
	if got, want := v.Marks(), NewValueMarks(); !want.Equal(got) {
		t.Errorf("wrong v marks after unmarking\ngot:  %#v\nwant: %#v", got, want)
	}
	if got, want := marks1234, NewValueMarks(1, 2, 3, 4); !want.Equal(got) {
		t.Errorf("wrong marks1234\ngot:  %#v\nwant: %#v", got, want)
	}
	if got, want := v, False; !want.RawEquals(got) {
		t.Errorf("wrong v after unmarking\ngot:  %#v\nwant: %#v", got, want)
	}

	// One more test for a more interesting/realistic situation involving
	// a number of different operations.
	a := NumberIntVal(2).Mark("a")
	b := NumberIntVal(5).Mark("b")
	c := NumberIntVal(1).Mark("c")
	d := NumberIntVal(12).Mark("d")
	result := a.Multiply(b).Subtract(c).GreaterThanOrEqualTo(d)
	if got, want := result, False.WithMarks(NewValueMarks("a", "b", "c", "d")); !want.RawEquals(got) {
		t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, want)
	}

	// Unmark the result and capture the paths
	unmarkedResult, pvm := result.UnmarkDeepWithPaths()
	// Remark the result with those paths
	remarked := unmarkedResult.MarkWithPaths(pvm)
	if got, want := remarked, False.WithMarks(NewValueMarks("a", "b", "c", "d")); !want.RawEquals(got) {
		t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, want)
	}

	// If we call MarkWithPaths without any matching paths, we should get the unmarked result
	markedWithNoPaths := unmarkedResult.MarkWithPaths([]PathValueMarks{{Path{IndexStep{Key: NumberIntVal(0)}}, NewValueMarks("z")}})
	if got, want := markedWithNoPaths, False; !want.RawEquals(got) {
		t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestPathValueMarksEqual(t *testing.T) {
	tests := []struct {
		original PathValueMarks
		compare  PathValueMarks
		want     bool
	}{
		{
			PathValueMarks{Path{IndexStep{Key: NumberIntVal(0)}}, NewValueMarks("a")},
			PathValueMarks{Path{IndexStep{Key: NumberIntVal(0)}}, NewValueMarks("a")},
			true,
		},
		{
			PathValueMarks{Path{IndexStep{Key: StringVal("p")}}, NewValueMarks(123)},
			PathValueMarks{Path{IndexStep{Key: StringVal("p")}}, NewValueMarks(123)},
			true,
		},
		{
			PathValueMarks{Path{IndexStep{Key: NumberIntVal(0)}}, NewValueMarks("a")},
			PathValueMarks{Path{IndexStep{Key: NumberIntVal(1)}}, NewValueMarks("a")},
			false,
		},
		{
			PathValueMarks{Path{IndexStep{Key: NumberIntVal(0)}}, NewValueMarks("a")},
			PathValueMarks{Path{IndexStep{Key: NumberIntVal(0)}}, NewValueMarks("b")},
			false,
		},
		{
			PathValueMarks{Path{IndexStep{Key: NumberIntVal(0)}}, NewValueMarks("a")},
			PathValueMarks{Path{IndexStep{Key: NumberIntVal(1)}}, NewValueMarks("b")},
			false,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("Comparing %#v to %#v", test.original, test.compare), func(t *testing.T) {
			got := test.original.Equal(test.compare)
			if got != test.want {
				t.Errorf("wrong result\ngot: %v\nwant: %v", got, test.want)
			}
		})
	}
}

func TestMarks(t *testing.T) {
	wantMarks := func(marks ValueMarks, expected ...string) {
		if len(marks) != len(expected) {
			t.Fatalf("wrong marks: %#v", marks)
		}
		for _, mark := range expected {
			if _, ok := marks[mark]; !ok {
				t.Fatalf("missing mark %q: %#v", mark, marks)
			}
		}
	}

	// Single mark
	val := StringVal("foo").Mark("a")
	wantMarks(val.Marks(), "a")
	val, marks := val.Unmark()
	if val.IsMarked() {
		t.Fatalf("still marked after unmark: %#v", marks)
	}
	wantMarks(marks, "a")

	// Multiple marks
	val = val.WithMarks(NewValueMarks("a", "b", "c"))
	wantMarks(val.Marks(), "a", "b", "c")
	val, marks = val.Unmark()
	if val.IsMarked() {
		t.Fatalf("still marked after unmark: %#v", marks)
	}
	wantMarks(marks, "a", "b", "c")

	// Multiple marks, applied separately
	val = val.Mark("a").Mark("b")
	wantMarks(val.Marks(), "a", "b")
	val, marks = val.Unmark()
	if val.IsMarked() {
		t.Fatalf("still marked after unmark: %#v", marks)
	}
	wantMarks(marks, "a", "b")
}

func TestUnmarkDeep(t *testing.T) {
	testCases := map[string]struct {
		val   Value
		want  Value
		marks ValueMarks
	}{
		"unmarked string": {
			StringVal("a"),
			StringVal("a"),
			NewValueMarks(),
		},
		"marked number": {
			NumberIntVal(1).Mark("a"),
			NumberIntVal(1),
			NewValueMarks("a"),
		},
		"unmarked list": {
			ListVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			ListVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			NewValueMarks(),
		},
		"list with some elements marked": {
			ListVal([]Value{NumberIntVal(1).Mark("a"), NumberIntVal(2)}),
			ListVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			NewValueMarks("a"),
		},
		"marked list with all elements marked": {
			ListVal([]Value{NumberIntVal(1).Mark("a"), NumberIntVal(2).Mark("b")}).Mark("c"),
			ListVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			NewValueMarks("a", "b", "c"),
		},
		"marked empty list": {
			ListValEmpty(String).Mark("c"),
			ListValEmpty(String),
			NewValueMarks("c"),
		},
		"map with elements marked": {
			MapVal(map[string]Value{"a": StringVal("b").Mark("c"), "x": StringVal("y").Mark("z")}),
			MapVal(map[string]Value{"a": StringVal("b"), "x": StringVal("y")}),
			NewValueMarks("c", "z"),
		},
		"tuple with elements marked": {
			TupleVal([]Value{NumberIntVal(1).Mark("a"), StringVal("y").Mark("z")}),
			TupleVal([]Value{NumberIntVal(1), StringVal("y")}),
			NewValueMarks("a", "z"),
		},
		"set with elements marked": {
			SetVal([]Value{NumberIntVal(1).Mark("a"), NumberIntVal(2).Mark("z")}),
			SetVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			NewValueMarks("a", "z"),
		},
		"complex marked object with lots of marks": {
			ObjectVal(map[string]Value{
				"x": ListVal([]Value{
					NumberIntVal(3).Mark("a"),
					NumberIntVal(5).Mark("b"),
				}).WithMarks(NewValueMarks("c", "d")),
				"y": StringVal("y").Mark("e"),
				"z": BoolVal(true).Mark("f"),
			}).Mark("g"),
			ObjectVal(map[string]Value{
				"x": ListVal([]Value{
					NumberIntVal(3),
					NumberIntVal(5),
				}),
				"y": StringVal("y"),
				"z": BoolVal(true),
			}),
			NewValueMarks("a", "b", "c", "d", "e", "f", "g"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, marks := tc.val.UnmarkDeep()
			if !got.RawEquals(tc.want) {
				t.Errorf("wrong value\n got: %#v\nwant: %#v", got, tc.want)
			}
			if !marks.Equal(tc.marks) {
				t.Errorf("wrong marks\n got: %#v\nwant: %#v", got, tc.want)
			}
		})
	}
}

func TestPathValueMarks(t *testing.T) {
	testCases := map[string]struct {
		marked   Value
		unmarked Value
		pvms     []PathValueMarks
	}{
		"unmarked string": {
			StringVal("a"),
			StringVal("a"),
			nil,
		},
		"marked number": {
			NumberIntVal(1).Mark("a"),
			NumberIntVal(1),
			[]PathValueMarks{
				{Path{}, NewValueMarks("a")},
			},
		},
		"list with some elements marked": {
			ListVal([]Value{NumberIntVal(1).Mark("a"), NumberIntVal(2)}),
			ListVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			[]PathValueMarks{
				{IndexIntPath(0), NewValueMarks("a")},
			},
		},
		"marked list with all elements marked": {
			ListVal([]Value{NumberIntVal(1).Mark("a"), NumberIntVal(2).Mark("b")}).Mark("c"),
			ListVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			[]PathValueMarks{
				{Path{}, NewValueMarks("c")},
				{IndexIntPath(0), NewValueMarks("a")},
				{IndexIntPath(1), NewValueMarks("b")},
			},
		},
		"marked empty list": {
			ListValEmpty(String).Mark("c"),
			ListValEmpty(String),
			[]PathValueMarks{
				{Path{}, NewValueMarks("c")},
			},
		},
		"map with elements marked": {
			MapVal(map[string]Value{"a": StringVal("b").Mark("c"), "x": StringVal("y").Mark("z")}),
			MapVal(map[string]Value{"a": StringVal("b"), "x": StringVal("y")}),
			[]PathValueMarks{
				{IndexStringPath("a"), NewValueMarks("c")},
				{IndexStringPath("x"), NewValueMarks("z")},
			},
		},
		"tuple with elements marked": {
			TupleVal([]Value{NumberIntVal(1).Mark("a"), StringVal("y").Mark("z"), ObjectVal(map[string]Value{"x": True}).Mark("o")}),
			TupleVal([]Value{NumberIntVal(1), StringVal("y"), ObjectVal(map[string]Value{"x": True})}),
			[]PathValueMarks{
				{IndexIntPath(0), NewValueMarks("a")},
				{IndexIntPath(1), NewValueMarks("z")},
				{IndexIntPath(2), NewValueMarks("o")},
			},
		},
		"set with elements marked": {
			SetVal([]Value{NumberIntVal(1).Mark("a"), NumberIntVal(2).Mark("z")}),
			SetVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			[]PathValueMarks{
				{Path{}, NewValueMarks("a", "z")},
			},
		},
		"complex marked object with lots of marks": {
			ObjectVal(map[string]Value{
				"x": ListVal([]Value{
					NumberIntVal(3).Mark("a"),
					NumberIntVal(5).Mark("b"),
				}).WithMarks(NewValueMarks("c", "d")),
				"y": StringVal("y").Mark("e"),
				"z": BoolVal(true).Mark("f"),
			}).Mark("g"),
			ObjectVal(map[string]Value{
				"x": ListVal([]Value{
					NumberIntVal(3),
					NumberIntVal(5),
				}),
				"y": StringVal("y"),
				"z": BoolVal(true),
			}),
			[]PathValueMarks{
				{Path{}, NewValueMarks("g")},
				{GetAttrPath("x"), NewValueMarks("c", "d")},
				{GetAttrPath("x").IndexInt(0), NewValueMarks("a")},
				{GetAttrPath("x").IndexInt(1), NewValueMarks("b")},
				{GetAttrPath("y"), NewValueMarks("e")},
				{GetAttrPath("z"), NewValueMarks("f")},
			},
		},
		"path array reuse regression test": {
			ObjectVal(map[string]Value{
				"environment": ListVal([]Value{
					ObjectVal(map[string]Value{
						"variables": MapVal(map[string]Value{
							"bar": StringVal("secret").Mark("sensitive"),
							"foo": StringVal("secret").Mark("sensitive"),
						}),
					}),
				}),
			}),
			ObjectVal(map[string]Value{
				"environment": ListVal([]Value{
					ObjectVal(map[string]Value{
						"variables": MapVal(map[string]Value{
							"bar": StringVal("secret"),
							"foo": StringVal("secret"),
						}),
					}),
				}),
			}),
			[]PathValueMarks{
				{GetAttrPath("environment").IndexInt(0).GetAttr("variables").IndexString("bar"), NewValueMarks("sensitive")},
				{GetAttrPath("environment").IndexInt(0).GetAttr("variables").IndexString("foo"), NewValueMarks("sensitive")},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(fmt.Sprintf("unmark: %s", name), func(t *testing.T) {
			got, pvms := tc.marked.UnmarkDeepWithPaths()
			if !got.RawEquals(tc.unmarked) {
				t.Errorf("wrong value\n got: %#v\nwant: %#v", got, tc.unmarked)
			}

			if len(pvms) != len(tc.pvms) {
				t.Errorf("wrong length\n got: %d\nwant: %d", len(pvms), len(tc.pvms))
			}

		findPvm:
			for _, wantPvm := range tc.pvms {
				for _, gotPvm := range pvms {
					if gotPvm.Path.Equals(wantPvm.Path) && gotPvm.Marks.Equal(wantPvm.Marks) {
						continue findPvm
					}
				}
				t.Errorf("missing %#v\nnot found in: %#v", wantPvm, pvms)
			}
		})

		t.Run(fmt.Sprintf("mark: %s", name), func(t *testing.T) {
			got := tc.unmarked.MarkWithPaths(tc.pvms)
			if !got.RawEquals(tc.marked) {
				t.Errorf("wrong value\n got: %#v\nwant: %#v", got, tc.marked)
			}
		})
	}
}
