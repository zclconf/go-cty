package stdlib

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestBase64Gzip(t *testing.T) {
	tests := []struct {
		String cty.Value
		Want   cty.Value
		Err    bool
	}{
		{
			cty.StringVal("test"),
			cty.StringVal("H4sIAAAAAAAA/ypJLS4BAAAA//8BAAD//wx+f9gEAAAA"),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("base64gzip(%#v)", test.String), func(t *testing.T) {
			got, err := Base64Gzip(test.String)

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestURLEncode(t *testing.T) {
	tests := []struct {
		String cty.Value
		Want   cty.Value
		Err    bool
	}{
		{
			cty.StringVal("abc123-_"),
			cty.StringVal("abc123-_"),
			false,
		},
		{
			cty.StringVal("foo:bar@localhost?foo=bar&bar=baz"),
			cty.StringVal("foo%3Abar%40localhost%3Ffoo%3Dbar%26bar%3Dbaz"),
			false,
		},
		{
			cty.StringVal("mailto:email?subject=this+is+my+subject"),
			cty.StringVal("mailto%3Aemail%3Fsubject%3Dthis%2Bis%2Bmy%2Bsubject"),
			false,
		},
		{
			cty.StringVal("foo/bar"),
			cty.StringVal("foo%2Fbar"),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("urlencode(%#v)", test.String), func(t *testing.T) {
			got, err := URLEncode(test.String)

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
