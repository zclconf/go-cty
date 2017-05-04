package stdlib

import (
	"strings"

	"golang.org/x/text/unicode/norm"

	"github.com/apparentlymart/go-cty/cty"
	"github.com/apparentlymart/go-cty/cty/function"
)

// Upper is a Function that converts a given string to uppercase.
var Upper = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "str",
			Type:             cty.String,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		in := args[0].AsString()
		out := strings.ToUpper(in)
		return cty.StringVal(out), nil
	},
})

// Lower is a Function that converts a given string to lowercase.
var Lower = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "str",
			Type:             cty.String,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		in := args[0].AsString()
		out := strings.ToLower(in)
		return cty.StringVal(out), nil
	},
})

// Reverse is a Function that reverses the order of the characters in the
// given string.
//
// As usual, "character" for the sake of this function is a grapheme cluster,
// so combining diacritics (for example) will be considered together as a
// single character.
var Reverse = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "str",
			Type:             cty.String,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		in := []byte(args[0].AsString())
		out := make([]byte, len(in))
		pos := len(out)

		for i := 0; i < len(in); {
			d := norm.NFC.NextBoundary(in[i:], true)
			cluster := in[i : i+d]
			pos -= len(cluster)
			copy(out[pos:], cluster)
			i += d
		}

		return cty.StringVal(string(out)), nil
	},
})
