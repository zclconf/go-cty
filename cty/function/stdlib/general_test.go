package stdlib

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestEqual(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(1),
			cty.NumberIntVal(2),
			cty.False,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(2),
			cty.True,
		},
		{
			cty.NullVal(cty.Number),
			cty.NullVal(cty.Number),
			cty.True,
		},
		{
			cty.NumberIntVal(2),
			cty.NullVal(cty.Number),
			cty.False,
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool).RefineNotNull(),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool).RefineNotNull(),
		},
		{
			cty.NumberIntVal(1),
			cty.DynamicVal,
			cty.UnknownVal(cty.Bool).RefineNotNull(),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Bool).RefineNotNull(),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Equal(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := Equal(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestCoalesce(t *testing.T) {
	tests := []struct {
		Values []cty.Value
		Want   cty.Value
	}{
		{
			[]cty.Value{cty.True},
			cty.True,
		},
		{
			[]cty.Value{cty.NullVal(cty.Bool), cty.True},
			cty.True,
		},
		{
			[]cty.Value{cty.NullVal(cty.Bool), cty.False},
			cty.False,
		},
		{
			[]cty.Value{cty.NullVal(cty.Bool), cty.False, cty.StringVal("hello")},
			cty.StringVal("false"),
		},
		{
			[]cty.Value{cty.True, cty.UnknownVal(cty.Bool)},
			cty.True,
		},
		{
			[]cty.Value{cty.UnknownVal(cty.Bool), cty.True},
			cty.UnknownVal(cty.Bool).RefineNotNull(),
		},
		{
			[]cty.Value{cty.UnknownVal(cty.Bool), cty.StringVal("hello")},
			cty.UnknownVal(cty.String).RefineNotNull(),
		},
		{
			[]cty.Value{cty.DynamicVal, cty.True},
			cty.UnknownVal(cty.Bool).RefineNotNull(),
		},
		{
			[]cty.Value{cty.DynamicVal},
			cty.DynamicVal,
		},
		{
			[]cty.Value{cty.UnknownVal(cty.Bool).Mark("a").RefineNotNull()},
			cty.UnknownVal(cty.Bool).Mark("a").RefineNotNull(),
		},
		{
			[]cty.Value{cty.UnknownVal(cty.Bool).Mark("a"), cty.UnknownVal(cty.Bool).Mark("b")},
			cty.UnknownVal(cty.Bool).Mark("a").Mark("b").RefineNotNull(),
		},
		{
			[]cty.Value{cty.UnknownVal(cty.Bool).Mark("a").RefineNotNull(), cty.UnknownVal(cty.Bool).Mark("b")},
			cty.UnknownVal(cty.Bool).Mark("a").RefineNotNull(),
		},
		{
			[]cty.Value{cty.UnknownVal(cty.Bool).Mark("a"), cty.UnknownVal(cty.Bool).Mark("b").RefineNotNull(), cty.UnknownVal(cty.Bool).Mark("c")},
			cty.UnknownVal(cty.Bool).Mark("a").Mark("b").RefineNotNull(),
		},
		{
			[]cty.Value{cty.NullVal(cty.Bool).Mark("n"), cty.UnknownVal(cty.Bool).Mark("a"), cty.UnknownVal(cty.Bool).Mark("b").RefineNotNull(), cty.UnknownVal(cty.Bool).Mark("c")},
			cty.UnknownVal(cty.Bool).Mark("a").Mark("b").Mark("n").RefineNotNull(),
		},
		{
			[]cty.Value{
				cty.UnknownVal(cty.String).Refine().
					StringPrefixFull("beep").
					NewValue(),
				cty.UnknownVal(cty.String).Refine().
					StringPrefixFull("boop").
					NewValue(),
			},
			cty.UnknownVal(cty.String).Refine().
				NotNull().
				StringPrefixFull("b"). // the common prefix of "beep" and "boop"
				NewValue(),
		},
		{
			[]cty.Value{
				cty.UnknownVal(cty.String).Refine().
					StringPrefixFull("beep").
					NewValue(),
				cty.UnknownVal(cty.String).Refine().
					NotNull().
					StringPrefixFull("boop").
					NewValue(),
				cty.UnknownVal(cty.String).Refine().
					StringPrefixFull("foo").
					NewValue(),
			},
			cty.UnknownVal(cty.String).Refine().
				NotNull().
				StringPrefixFull("b"). // the common prefix of "beep" and "boop"; "foo" is excluded because second argument is known not null
				NewValue(),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Coalesce(%#v...)", test.Values), func(t *testing.T) {
			got, err := Coalesce(test.Values...)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
