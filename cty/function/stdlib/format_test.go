package stdlib

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		Format  cty.Value
		Args    []cty.Value
		Want    cty.Value
		WantErr string
	}{
		{
			cty.StringVal(""),
			nil,
			cty.StringVal(""),
			``,
		},
		{
			cty.StringVal("hello"),
			nil,
			cty.StringVal("hello"),
			``,
		},
		{
			cty.StringVal("100%% successful"),
			nil,
			cty.StringVal("100% successful"),
			``,
		},
		{
			cty.StringVal("100%%"),
			nil,
			cty.StringVal("100%"),
			``,
		},

		// Strings
		{
			cty.StringVal("Hello, %s!"),
			[]cty.Value{cty.StringVal("Ermintrude")},
			cty.StringVal("Hello, Ermintrude!"),
			``,
		},

		// Integer Numbers
		{
			cty.StringVal("%d green bottles standing on the wall"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.StringVal("10 green bottles standing on the wall"),
			``,
		},
		{
			cty.StringVal("%i green bottles standing on the wall"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.StringVal("10 green bottles standing on the wall"),
			``,
		},

		// Invalids
		{
			cty.StringVal("%s is not in the args list"),
			nil,
			cty.NilVal,
			`not enough arguments for "%s" at 0: requires at least 1 total (1 starting at 1)`,
		},
		{
			cty.StringVal("%*s is not in the args list"),
			nil,
			cty.NilVal,
			`not enough arguments for "%*s" at 0: requires at least 2 total (2 starting at 1)`,
		},
		{
			cty.StringVal("%.*s is not in the args list"),
			nil,
			cty.NilVal,
			`not enough arguments for "%.*s" at 0: requires at least 2 total (2 starting at 1)`,
		},
		{
			cty.StringVal("%*.*s is not in the args list"),
			nil,
			cty.NilVal,
			`not enough arguments for "%*.*s" at 0: requires at least 3 total (3 starting at 1)`,
		},
		{
			cty.StringVal("%[3]s is not in the args list"),
			nil,
			cty.NilVal,
			`not enough arguments for "%[3]s" at 0: requires at least 3 total (1 starting at 3)`,
		},
		{
			cty.StringVal("%z is not a valid sequence"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.NilVal,
			`unsupported format verb 'z' in "%z" at offset 0`,
		},
		{
			cty.StringVal("%#z is not a valid sequence"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.NilVal,
			`unsupported format verb 'z' in "%#z" at offset 0`,
		},
		{
			cty.StringVal("%012z is not a valid sequence"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.NilVal,
			`unsupported format verb 'z' in "%012z" at offset 0`,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d-%#v", i, test.Format), func(t *testing.T) {
			got, err := Format(test.Format, test.Args...)

			if test.WantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
			} else {
				if err == nil {
					t.Fatalf("no error; want %q", test.WantErr)
				}
				errStr := err.Error()
				if errStr != test.WantErr {
					t.Fatalf("wrong error\ngot:  %s\nwant: %s", errStr, test.WantErr)
				}
				return
			}

			if test.Want != cty.NilVal {
				if !got.RawEquals(test.Want) {
					t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
				}
			} else {
				t.Errorf("unexpected success %#v; want error", got)
			}
		})
	}
}
