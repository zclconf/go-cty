package cty

import (
	"fmt"
	"testing"
)

func TestHasDynamicTypes(t *testing.T) {
	tests := []struct {
		ty       Type
		expected bool
	}{
		{
			DynamicPseudoType,
			true,
		},
		{
			List(DynamicPseudoType),
			true,
		},
		{
			Tuple([]Type{String, DynamicPseudoType}),
			true,
		},
		{
			Object(map[string]Type{
				"a":       String,
				"unknown": DynamicPseudoType,
			}),
			true,
		},
		{
			List(Object(map[string]Type{
				"a":       String,
				"unknown": DynamicPseudoType,
			})),
			true,
		},
		{
			Tuple([]Type{Object(map[string]Type{
				"a":       String,
				"unknown": DynamicPseudoType,
			})}),
			true,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.HasDynamicTypes()", test.ty), func(t *testing.T) {
			got := test.ty.HasDynamicTypes()
			if got != test.expected {
				t.Errorf("Equals returned %#v; want %#v", got, test.expected)
			}
		})
	}
}

func TestNilTypeEquals(t *testing.T) {
	var typ Type
	if !typ.Equals(NilType) {
		t.Fatal("expected NilTypes to equal")
	}
}
