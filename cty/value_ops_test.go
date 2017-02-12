package cty

import (
	"fmt"
	"testing"
)

func TestValueAdd(t *testing.T) {
	tests := []struct {
		LHS      Value
		RHS      Value
		Expected Value
	}{
		{
			NumberIntVal(1),
			NumberIntVal(2),
			NumberIntVal(3),
		},
		{
			NumberIntVal(1),
			NumberIntVal(-2),
			NumberIntVal(-1),
		},
		{
			NumberIntVal(1),
			NumberFloatVal(0.5),
			NumberFloatVal(1.5),
		},
		{
			NumberIntVal(1),
			Unknown(Number),
			Unknown(Number),
		},
		{
			Unknown(Number),
			Unknown(Number),
			Unknown(Number),
		},
		{
			NumberIntVal(1),
			DynamicValue,
			DynamicValue,
		},
		{
			DynamicValue,
			DynamicValue,
			DynamicValue,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.Add(%#v)", test.LHS, test.RHS), func(t *testing.T) {
			got := test.LHS.Add(test.RHS)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("Add returned %#v; want %#v", got, test.Expected)
			}
		})
	}
}
