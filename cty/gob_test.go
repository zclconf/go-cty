package cty

import (
	"bytes"
	"testing"

	"encoding/gob"
)

func TestGobabilty(t *testing.T) {
	tests := []Value{
		StringVal("hi"),
		True,
		NumberIntVal(1),
		NumberFloatVal(96.5),
		ListVal([]Value{True}),
		MapVal(map[string]Value{"true": True}),
		SetVal([]Value{True}),
		TupleVal([]Value{True}),
		ObjectVal(map[string]Value{"true": True}),

		// Numbers are particularly tricky because big.Float.GobEncode is
		// implemented as a pointer method and thus big floats lose their
		// "pointerness" during gob round-trip. For that reason, we're testing
		// all of the containers with nested numbers inside to make sure that
		// our fixup step is working correctly for all of them.
		ListVal([]Value{NumberIntVal(1)}),
		MapVal(map[string]Value{
			"num": NumberIntVal(1),
		}),
		SetVal([]Value{NumberIntVal(1)}),
		TupleVal([]Value{NumberIntVal(1)}),
		ObjectVal(map[string]Value{
			"num": NumberIntVal(1),
		}),
	}

	for _, testValue := range tests {
		t.Run(testValue.GoString(), func(t *testing.T) {
			tv := testGob{
				testValue,
			}

			buf := &bytes.Buffer{}
			enc := gob.NewEncoder(buf)

			err := enc.Encode(tv)
			if err != nil {
				t.Fatalf("gob encode error: %s", err)
			}

			var ov testGob

			dec := gob.NewDecoder(buf)
			err = dec.Decode(&ov)
			if err != nil {
				t.Fatalf("gob decode error: %s", err)
			}

			if !ov.Value.RawEquals(tv.Value) {
				t.Errorf("value did not survive gobbing\ninput:  %#v\noutput: %#v", tv, ov)
			}
		})
	}
}

type testGob struct {
	Value Value
}
