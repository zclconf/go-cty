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

}
