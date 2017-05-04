package stdlib

import (
	"testing"

	"github.com/apparentlymart/go-cty/cty"
)

func TestUpper(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.StringVal("hello"),
			cty.StringVal("HELLO"),
		},
		{
			cty.StringVal("HELLO"),
			cty.StringVal("HELLO"),
		},
		{
			cty.StringVal(""),
			cty.StringVal(""),
		},
		{
			cty.StringVal("1"),
			cty.StringVal("1"),
		},
		{
			cty.StringVal("жж"),
			cty.StringVal("ЖЖ"),
		},
		{
			cty.StringVal("noël"),
			cty.StringVal("NOËL"),
		},
		{
			// Go's case conversions don't handle this ligature, which is
			// unfortunate but is now a compatibility constraint since it
			// would be potentially-breaking to behave differently here in
			// future.
			cty.StringVal("baﬄe"),
			cty.StringVal("BAﬄE"),
		},
		{
			cty.StringVal("😸😾"),
			cty.StringVal("😸😾"),
		},
		{
			cty.UnknownVal(cty.String),
			cty.UnknownVal(cty.String),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.String),
		},
	}

	upper := Upper.Proxy()

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := upper(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestLower(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.StringVal("HELLO"),
			cty.StringVal("hello"),
		},
		{
			cty.StringVal("hello"),
			cty.StringVal("hello"),
		},
		{
			cty.StringVal(""),
			cty.StringVal(""),
		},
		{
			cty.StringVal("1"),
			cty.StringVal("1"),
		},
		{
			cty.StringVal("ЖЖ"),
			cty.StringVal("жж"),
		},
		{
			cty.UnknownVal(cty.String),
			cty.UnknownVal(cty.String),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.String),
		},
	}

	lower := Lower.Proxy()

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := lower(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
