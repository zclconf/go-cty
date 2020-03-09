package stdlib

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestReplace(t *testing.T) {
	tests := []struct {
		Input           cty.Value
		Substr, Replace cty.Value
		Want            cty.Value
	}{
		{
			cty.StringVal("hello"),
			cty.StringVal("l"),
			cty.StringVal(""),
			cty.StringVal("heo"),
		},
		{
			cty.StringVal("😸😸😸😾😾😾"),
			cty.StringVal("😾"),
			cty.StringVal("😸"),
			cty.StringVal("😸😸😸😸😸😸"),
		},
		{
			cty.StringVal("😸😸😸😸😸😾"),
			cty.StringVal("😾"),
			cty.StringVal("😸"),
			cty.StringVal("😸😸😸😸😸😸"),
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString()+"_replace", func(t *testing.T) {
			got, err := Replace(test.Input, test.Substr, test.Replace)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Fatalf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
		t.Run(test.Input.GoString()+"_regex_replace", func(t *testing.T) {
			got, err := Replace(test.Input, test.Substr, test.Replace)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Fatalf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestRegexReplace(t *testing.T) {
	tests := []struct {
		Input           cty.Value
		Substr, Replace cty.Value
		Want            cty.Value
	}{
		{
			cty.StringVal("-ab-axxb-"),
			cty.StringVal("a(x*)b"),
			cty.StringVal("T"),
			cty.StringVal("-T-T-"),
		},
		{
			cty.StringVal("-ab-axxb-"),
			cty.StringVal("a(x*)b"),
			cty.StringVal("${1}W"),
			cty.StringVal("-W-xxW-"),
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := RegexReplace(test.Input, test.Substr, test.Replace)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Fatalf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestRegexReplaceInvalidRegex(t *testing.T) {
	_, err := RegexReplace(cty.StringVal(""), cty.StringVal("("), cty.StringVal(""))
	if err == nil {
		t.Fatal("expected an error")
	}
}
