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

func TestWithoutOptionalAttributesDeep(t *testing.T) {
	tests := []struct {
		ty       Type
		expected Type
	}{
		{
			DynamicPseudoType,
			DynamicPseudoType,
		},
		{
			List(DynamicPseudoType),
			List(DynamicPseudoType),
		},
		{
			Tuple([]Type{String, DynamicPseudoType}),
			Tuple([]Type{String, DynamicPseudoType}),
		},
		{
			Object(map[string]Type{
				"a":       String,
				"unknown": DynamicPseudoType,
			}),
			Object(map[string]Type{
				"a":       String,
				"unknown": DynamicPseudoType,
			}),
		},
		{
			ObjectWithOptionalAttrs(map[string]Type{
				"a":       String,
				"unknown": DynamicPseudoType,
			}, []string{"a"}),
			Object(map[string]Type{
				"a":       String,
				"unknown": DynamicPseudoType,
			}),
		},
		{
			Map(ObjectWithOptionalAttrs(map[string]Type{
				"a":       String,
				"unknown": DynamicPseudoType,
			}, []string{"a"})),
			Map(Object(map[string]Type{
				"a":       String,
				"unknown": DynamicPseudoType,
			})),
		},
		{
			Set(ObjectWithOptionalAttrs(map[string]Type{
				"a":       String,
				"unknown": DynamicPseudoType,
			}, []string{"a"})),
			Set(Object(map[string]Type{
				"a":       String,
				"unknown": DynamicPseudoType,
			})),
		},
		{
			List(ObjectWithOptionalAttrs(map[string]Type{
				"a":       String,
				"unknown": DynamicPseudoType,
			}, []string{"a"})),
			List(Object(map[string]Type{
				"a":       String,
				"unknown": DynamicPseudoType,
			})),
		},
		{
			Tuple([]Type{
				ObjectWithOptionalAttrs(map[string]Type{
					"a":       String,
					"unknown": DynamicPseudoType,
				}, []string{"a"}),
				ObjectWithOptionalAttrs(map[string]Type{
					"b": Number,
				}, []string{"b"}),
			}),
			Tuple([]Type{
				Object(map[string]Type{
					"a":       String,
					"unknown": DynamicPseudoType,
				}),
				Object(map[string]Type{
					"b": Number,
				}),
			}),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.HasDynamicTypes()", test.ty), func(t *testing.T) {
			got := test.ty.WithoutOptionalAttributesDeep()
			if !test.expected.Equals(got) {
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

func TestTypeGoString(t *testing.T) {
	tests := []struct {
		Type Type
		Want string
	}{
		{
			DynamicPseudoType,
			`cty.DynamicPseudoType`,
		},
		{
			String,
			`cty.String`,
		},
		{
			Tuple([]Type{String, Bool}),
			`cty.Tuple([]cty.Type{cty.String, cty.Bool})`,
		},

		{
			Number,
			`cty.Number`,
		},
		{
			Bool,
			`cty.Bool`,
		},
		{
			List(String),
			`cty.List(cty.String)`,
		},
		{
			List(List(String)),
			`cty.List(cty.List(cty.String))`,
		},
		{
			List(Bool),
			`cty.List(cty.Bool)`,
		},
		{
			Set(String),
			`cty.Set(cty.String)`,
		},
		{
			Set(Map(String)),
			`cty.Set(cty.Map(cty.String))`,
		},
		{
			Set(Bool),
			`cty.Set(cty.Bool)`,
		},
		{
			Tuple([]Type{Bool}),
			`cty.Tuple([]cty.Type{cty.Bool})`,
		},

		{
			Map(String),
			`cty.Map(cty.String)`,
		},
		{
			Map(Set(String)),
			`cty.Map(cty.Set(cty.String))`,
		},
		{
			Map(Bool),
			`cty.Map(cty.Bool)`,
		},
		{
			Object(map[string]Type{"foo": Bool}),
			`cty.Object(map[string]cty.Type{"foo":cty.Bool})`,
		},
		{
			ObjectWithOptionalAttrs(map[string]Type{"foo": Bool, "bar": String}, []string{"bar"}),
			`cty.ObjectWithOptionalAttrs(map[string]cty.Type{"bar":cty.String, "foo":cty.Bool}, []string{"bar"})`,
		},
	}

	for _, test := range tests {
		t.Run(test.Type.GoString(), func(t *testing.T) {
			got := test.Type.GoString()
			want := test.Want
			if got != want {
				t.Errorf("wrong result\ngot:  %s\nwant: %s", got, want)
			}
		})
	}
}
