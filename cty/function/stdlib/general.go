package stdlib

import (
	"github.com/apparentlymart/go-cty/cty"
	"github.com/apparentlymart/go-cty/cty/function"
)

var EqualFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "a",
			Type:             cty.DynamicPseudoType,
			AllowUnknown:     true,
			AllowDynamicType: true,
			AllowNull:        true,
		},
		{
			Name:             "b",
			Type:             cty.DynamicPseudoType,
			AllowUnknown:     true,
			AllowDynamicType: true,
			AllowNull:        true,
		},
	},
	Type: function.StaticReturnType(cty.Bool),
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		return args[0].Equals(args[1]), nil
	},
})

// Equal determines whether the two given values are equal, returning a
// bool value.
func Equal(a cty.Value, b cty.Value) (cty.Value, error) {
	return EqualFunc.Call([]cty.Value{a, b})
}
