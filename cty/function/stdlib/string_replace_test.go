package stdlib

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestReplace(t *testing.T) {
	tests := []struct {
		Input              cty.Value
		Substr, Replace, N cty.Value
		Want               cty.Value
	}{
		{
			cty.StringVal("hello"),
			cty.StringVal("l"),
			cty.StringVal(""),
			cty.NumberIntVal(1),
			cty.StringVal("helo"),
		},
		{
			cty.StringVal("hello"),
			cty.StringVal("l"),
			cty.StringVal(""),
			cty.NumberIntVal(-1),
			cty.StringVal("heo"),
		},
		{
			cty.StringVal("😸😸😸😾😾😾"),
			cty.StringVal("😾"),
			cty.StringVal("😸"),
			cty.NumberIntVal(1),
			cty.StringVal("😸😸😸😸😾😾"),
		},
		{
			cty.StringVal("😸😸😸😾😾😾"),
			cty.StringVal("😾"),
			cty.StringVal("😸"),
			cty.NumberIntVal(-1),
			cty.StringVal("😸😸😸😸😸😸"),
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := Replace(test.Input, test.Substr, test.Replace, test.N)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
