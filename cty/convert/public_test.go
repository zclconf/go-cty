package convert

import (
	"fmt"
	"testing"

	"github.com/apparentlymart/go-cty/cty"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		Value     cty.Value
		Type      cty.Type
		Want      cty.Value
		WantError bool
	}{
		{
			Value: cty.StringVal("hello"),
			Type:  cty.String,
			Want:  cty.StringVal("hello"),
		},
		{
			Value: cty.StringVal("1"),
			Type:  cty.Number,
			Want:  cty.NumberIntVal(1),
		},
		{
			Value: cty.StringVal("1.5"),
			Type:  cty.Number,
			Want:  cty.NumberFloatVal(1.5),
		},
		{
			Value:     cty.StringVal("hello"),
			Type:      cty.Number,
			WantError: true,
		},
		{
			Value: cty.StringVal("true"),
			Type:  cty.Bool,
			Want:  cty.True,
		},
		{
			Value: cty.StringVal("1"),
			Type:  cty.Bool,
			Want:  cty.True,
		},
		{
			Value: cty.StringVal("false"),
			Type:  cty.Bool,
			Want:  cty.False,
		},
		{
			Value: cty.StringVal("0"),
			Type:  cty.Bool,
			Want:  cty.False,
		},
		{
			Value:     cty.StringVal("hello"),
			Type:      cty.Bool,
			WantError: true,
		},
		{
			Value: cty.NumberIntVal(4),
			Type:  cty.String,
			Want:  cty.StringVal("4"),
		},
		{
			Value: cty.NumberFloatVal(3.14159265359),
			Type:  cty.String,
			Want:  cty.StringVal("3.14159265359"),
		},
		{
			Value: cty.True,
			Type:  cty.String,
			Want:  cty.StringVal("true"),
		},
		{
			Value: cty.False,
			Type:  cty.String,
			Want:  cty.StringVal("false"),
		},
		{
			Value: cty.UnknownVal(cty.String),
			Type:  cty.Number,
			Want:  cty.UnknownVal(cty.Number),
		},
		{
			Value: cty.UnknownVal(cty.Number),
			Type:  cty.String,
			Want:  cty.UnknownVal(cty.String),
		},
		{
			Value: cty.DynamicVal,
			Type:  cty.String,
			Want:  cty.UnknownVal(cty.String),
		},
		{
			Value: cty.StringVal("hello"),
			Type:  cty.DynamicPseudoType,
			Want:  cty.StringVal("hello"),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v to %#v", test.Value, test.Type), func(t *testing.T) {
			got, err := Convert(test.Value, test.Type)

			switch {
			case test.WantError:
				if err == nil {
					t.Errorf("conversion succeeded with %#v; want error", got)
				}
			default:
				if err != nil {
					t.Fatalf("conversion failed: %s", err)
				}

				if !got.RawEquals(test.Want) {
					t.Errorf(
						"wrong result\nvalue: %#v\ntype:  %#v\ngot:   %#v\nwant:  %#v",
						test.Value, test.Type,
						got, test.Want,
					)
				}
			}
		})
	}
}
