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

func TestReverse(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.StringVal("hello"),
			cty.StringVal("olleh"),
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
			cty.StringVal("Живой Журнал"),
			cty.StringVal("ланруЖ йовиЖ"),
		},
		{
			// note that the dieresis here is intentionally a combining
			// ligature.
			cty.StringVal("noël"),
			cty.StringVal("lëon"),
		},
		{
			// The Es in this string has three combining acute accents.
			// This tests something that NFC-normalization cannot collapse
			// into a single precombined codepoint, since otherwise we might
			// be cheating and relying on the single-codepoint forms.
			cty.StringVal("wé́́é́́é́́!"),
			cty.StringVal("!é́́é́́é́́w"),
		},
		{
			// Go's normalization forms don't handle this ligature, so we
			// will produce the wrong result but this is now a compatibility
			// constraint and so we'll test it.
			cty.StringVal("baﬄe"),
			cty.StringVal("eﬄab"),
		},
		{
			cty.StringVal("😸😾"),
			cty.StringVal("😾😸"),
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

	reverse := Reverse.Proxy()

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := reverse(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
