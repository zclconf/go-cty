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

func TestValueSub(t *testing.T) {
	tests := []struct {
		LHS      Value
		RHS      Value
		Expected Value
	}{
		{
			NumberIntVal(1),
			NumberIntVal(2),
			NumberIntVal(-1),
		},
		{
			NumberIntVal(1),
			NumberIntVal(-2),
			NumberIntVal(3),
		},
		{
			NumberIntVal(1),
			NumberFloatVal(0.5),
			NumberFloatVal(0.5),
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
		t.Run(fmt.Sprintf("%#v.Sub(%#v)", test.LHS, test.RHS), func(t *testing.T) {
			got := test.LHS.Sub(test.RHS)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("Sub returned %#v; want %#v", got, test.Expected)
			}
		})
	}
}

func TestValueNeg(t *testing.T) {
	tests := []struct {
		Receiver Value
		Expected Value
	}{
		{
			NumberIntVal(1),
			NumberIntVal(-1),
		},
		{
			NumberFloatVal(0.5),
			NumberFloatVal(-0.5),
		},
		{
			Unknown(Number),
			Unknown(Number),
		},
		{
			DynamicValue,
			DynamicValue,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.Neg()", test.Receiver), func(t *testing.T) {
			got := test.Receiver.Neg()
			if !got.RawEquals(test.Expected) {
				t.Fatalf("Neg returned %#v; want %#v", got, test.Expected)
			}
		})
	}
}
