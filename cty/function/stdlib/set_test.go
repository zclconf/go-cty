package stdlib

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestSetUnion(t *testing.T) {
	tests := []struct {
		Input []cty.Value
		Want  cty.Value
	}{
		{
			[]cty.Value{
				cty.SetValEmpty(cty.String),
			},
			cty.SetValEmpty(cty.String),
		},
		{
			[]cty.Value{
				cty.SetValEmpty(cty.String),
				cty.SetValEmpty(cty.String),
			},
			cty.SetValEmpty(cty.String),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.True}),
				cty.SetValEmpty(cty.String),
			},
			cty.SetVal([]cty.Value{cty.StringVal("true")}),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.True}),
				cty.SetVal([]cty.Value{cty.True}),
				cty.SetVal([]cty.Value{cty.False}),
			},
			cty.SetVal([]cty.Value{
				cty.True,
				cty.False,
			}),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.StringVal("a")}),
				cty.SetVal([]cty.Value{cty.StringVal("b")}),
				cty.SetVal([]cty.Value{cty.StringVal("b"), cty.StringVal("c")}),
			},
			cty.SetVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
				cty.StringVal("c"),
			}),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.True}),
				cty.SetValEmpty(cty.DynamicPseudoType),
			},
			cty.SetVal([]cty.Value{cty.True}),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.EmptyObjectVal}),
				cty.SetValEmpty(cty.DynamicPseudoType),
			},
			cty.SetVal([]cty.Value{cty.EmptyObjectVal}),
		},
		{
			[]cty.Value{
				cty.SetValEmpty(cty.DynamicPseudoType),
				cty.SetValEmpty(cty.DynamicPseudoType),
			},
			cty.SetValEmpty(cty.DynamicPseudoType),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.StringVal("5")}),
				cty.UnknownVal(cty.Set(cty.Number)),
			},
			cty.UnknownVal(cty.Set(cty.String)),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.StringVal("5")}),
				cty.SetVal([]cty.Value{cty.UnknownVal(cty.String)}),
			},
			cty.SetVal([]cty.Value{cty.StringVal("5"), cty.UnknownVal(cty.String)}),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("SetUnion(%#v...)", test.Input), func(t *testing.T) {
			got, err := SetUnion(test.Input...)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestSetIntersection(t *testing.T) {
	tests := []struct {
		Input []cty.Value
		Want  cty.Value
	}{
		{
			[]cty.Value{
				cty.SetValEmpty(cty.String),
			},
			cty.SetValEmpty(cty.String),
		},
		{
			[]cty.Value{
				cty.SetValEmpty(cty.String),
				cty.SetValEmpty(cty.String),
			},
			cty.SetValEmpty(cty.String),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.True}),
				cty.SetValEmpty(cty.String),
			},
			cty.SetValEmpty(cty.String),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.True}),
				cty.SetVal([]cty.Value{cty.True, cty.False}),
				cty.SetVal([]cty.Value{cty.True, cty.False}),
			},
			cty.SetVal([]cty.Value{
				cty.True,
			}),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b")}),
				cty.SetVal([]cty.Value{cty.StringVal("b")}),
				cty.SetVal([]cty.Value{cty.StringVal("b"), cty.StringVal("c")}),
			},
			cty.SetVal([]cty.Value{
				cty.StringVal("b"),
			}),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.True}),
				cty.SetValEmpty(cty.DynamicPseudoType),
			},
			cty.SetValEmpty(cty.Bool),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.EmptyObjectVal}),
				cty.SetValEmpty(cty.DynamicPseudoType),
			},
			cty.SetValEmpty(cty.EmptyObject),
		},
		{
			[]cty.Value{
				cty.SetValEmpty(cty.DynamicPseudoType),
				cty.SetValEmpty(cty.DynamicPseudoType),
			},
			cty.SetValEmpty(cty.DynamicPseudoType),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.StringVal("5")}),
				cty.UnknownVal(cty.Set(cty.Number)),
			},
			cty.UnknownVal(cty.Set(cty.String)),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.StringVal("5")}),
				cty.SetVal([]cty.Value{cty.UnknownVal(cty.String)}),
			},
			cty.UnknownVal(cty.Set(cty.String)),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("SetIntersection(%#v...)", test.Input), func(t *testing.T) {
			got, err := SetIntersection(test.Input...)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestSetSubtract(t *testing.T) {
	tests := []struct {
		InputA cty.Value
		InputB cty.Value
		Want   cty.Value
	}{
		{
			cty.SetValEmpty(cty.String),
			cty.SetValEmpty(cty.String),
			cty.SetValEmpty(cty.String),
		},
		{
			cty.SetVal([]cty.Value{cty.True}),
			cty.SetValEmpty(cty.String),
			cty.SetVal([]cty.Value{cty.StringVal("true")}),
		},
		{
			cty.SetVal([]cty.Value{cty.True}),
			cty.SetVal([]cty.Value{cty.False}),
			cty.SetVal([]cty.Value{cty.True}),
		},
		{
			cty.SetVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
				cty.StringVal("c"),
			}),
			cty.SetVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("c"),
			}),
			cty.SetVal([]cty.Value{
				cty.StringVal("b"),
			}),
		},
		{
			cty.SetVal([]cty.Value{cty.StringVal("a")}),
			cty.SetValEmpty(cty.DynamicPseudoType),
			cty.SetVal([]cty.Value{cty.StringVal("a")}),
		},
		{
			cty.SetVal([]cty.Value{cty.EmptyObjectVal}),
			cty.SetValEmpty(cty.DynamicPseudoType),
			cty.SetVal([]cty.Value{cty.EmptyObjectVal}),
		},
		{
			cty.SetValEmpty(cty.DynamicPseudoType),
			cty.SetValEmpty(cty.DynamicPseudoType),
			cty.SetValEmpty(cty.DynamicPseudoType),
		},
		{
			cty.SetVal([]cty.Value{cty.StringVal("5")}),
			cty.UnknownVal(cty.Set(cty.Number)),
			cty.UnknownVal(cty.Set(cty.String)),
		},
		{
			cty.SetVal([]cty.Value{cty.StringVal("5")}),
			cty.SetVal([]cty.Value{cty.UnknownVal(cty.String)}),
			cty.UnknownVal(cty.Set(cty.String)),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("SetSubtract(%#v, %#v)", test.InputA, test.InputB), func(t *testing.T) {
			got, err := SetSubtract(test.InputA, test.InputB)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestSetSymmetricDifference(t *testing.T) {
	tests := []struct {
		InputA cty.Value
		InputB cty.Value
		Want   cty.Value
	}{
		{
			cty.SetValEmpty(cty.String),
			cty.SetValEmpty(cty.String),
			cty.SetValEmpty(cty.String),
		},
		{
			cty.SetVal([]cty.Value{cty.True}),
			cty.SetValEmpty(cty.String),
			cty.SetVal([]cty.Value{cty.StringVal("true")}),
		},
		{
			cty.SetVal([]cty.Value{cty.True}),
			cty.SetVal([]cty.Value{cty.False}),
			cty.SetVal([]cty.Value{cty.True, cty.False}),
		},
		{
			cty.SetVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
				cty.StringVal("c"),
			}),
			cty.SetVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("c"),
			}),
			cty.SetVal([]cty.Value{
				cty.StringVal("b"),
			}),
		},
		{
			cty.SetVal([]cty.Value{cty.StringVal("a")}),
			cty.SetValEmpty(cty.DynamicPseudoType),
			cty.SetVal([]cty.Value{cty.StringVal("a")}),
		},
		{
			cty.SetVal([]cty.Value{cty.EmptyObjectVal}),
			cty.SetValEmpty(cty.DynamicPseudoType),
			cty.SetVal([]cty.Value{cty.EmptyObjectVal}),
		},
		{
			cty.SetValEmpty(cty.DynamicPseudoType),
			cty.SetValEmpty(cty.DynamicPseudoType),
			cty.SetValEmpty(cty.DynamicPseudoType),
		},
		{
			cty.SetVal([]cty.Value{cty.StringVal("5")}),
			cty.UnknownVal(cty.Set(cty.Number)),
			cty.UnknownVal(cty.Set(cty.String)),
		},
		{
			cty.SetVal([]cty.Value{cty.StringVal("5")}),
			cty.SetVal([]cty.Value{cty.UnknownVal(cty.Number)}),
			cty.UnknownVal(cty.Set(cty.String)),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("SetSymmetricDifference(%#v, %#v)", test.InputA, test.InputB), func(t *testing.T) {
			got, err := SetSymmetricDifference(test.InputA, test.InputB)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
