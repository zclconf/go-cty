package cty

import (
	"testing"
)

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
}

func TestUnmarkDeep(t *testing.T) {
	v := NumberIntVal(1).Mark("a")
	v1 := NumberIntVal(2)
	l := ListVal([]Value{v, v1})
	if l.IsMarked() {
		t.Error("Value containing marks should not be marked itself")
	}
	if !l.ContainsMarked() {
		t.Error("Value containing marks should be caught by ContainsMarked")
	}

	l1, marks := l.UnmarkDeep()
	if got, want := l1, ListVal([]Value{NumberIntVal(1), v1}); !want.RawEquals(got) {
		t.Errorf("wrong result\ngot: #%v\nwant: %#v", got, want)
	}
	if got, want := marks, NewValueMarks("a"); !want.Equal(got) {
		t.Errorf("wrong result\ngot: #%v\nwant: %#v", got, want)
	}

	l2, paths := l.UnmarkDeepWithPaths()
	if got, want := l2, ListVal([]Value{NumberIntVal(1), v1}); !want.RawEquals(got) {
		t.Errorf("wrong result\ngot: #%v\nwant: %#v", got, want)
	}
	expectedPathValueMarks := []PathValueMarks{{Path{IndexStep{Key: NumberIntVal(0)}}, NewValueMarks("a")}, {}, {}}
	for i, p := range paths {
		if got, want := p, expectedPathValueMarks[i]; !want.Equal(got) {
			t.Errorf("wrong result\ngot: #%v\nwant: %#v", got, want)
		}
	}
}
