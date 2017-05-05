package stdlib

import (
	"fmt"
	"math/big"

	"github.com/apparentlymart/go-cty/cty"
	"github.com/apparentlymart/go-cty/cty/function"
)

var MinFunc = function.New(&function.Spec{
	Params: []function.Parameter{},
	VarParam: &function.Parameter{
		Name:             "numbers",
		Type:             cty.Number,
		AllowDynamicType: true,
	},
	Type: function.StaticReturnType(cty.Number),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		if len(args) == 0 {
			return cty.NilVal, fmt.Errorf("must pass at least one number")
		}

		min := cty.PositiveInfinity
		for _, num := range args {
			if num.LessThan(min).True() {
				min = num
			}
		}

		return min, nil
	},
})

var MaxFunc = function.New(&function.Spec{
	Params: []function.Parameter{},
	VarParam: &function.Parameter{
		Name:             "numbers",
		Type:             cty.Number,
		AllowDynamicType: true,
	},
	Type: function.StaticReturnType(cty.Number),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		if len(args) == 0 {
			return cty.NilVal, fmt.Errorf("must pass at least one number")
		}

		max := cty.NegativeInfinity
		for _, num := range args {
			if num.GreaterThan(max).True() {
				max = num
			}
		}

		return max, nil
	},
})

var IntFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "num",
			Type:             cty.Number,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.Number),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		bf := args[0].AsBigFloat()
		if bf.IsInt() {
			return args[0], nil
		}
		bi, _ := bf.Int(nil)
		bf = (&big.Float{}).SetInt(bi)
		return cty.NumberVal(bf), nil
	},
})

// Min returns the minimum number from the given numbers.
func Min(numbers ...cty.Value) (cty.Value, error) {
	return MinFunc.Call(numbers)
}

// Max returns the maximum number from the given numbers.
func Max(numbers ...cty.Value) (cty.Value, error) {
	return MaxFunc.Call(numbers)
}

// Int removes the fractional component of the given number returning an
// integer representing the whole number component, rounding towards zero.
// For example, -1.5 becomes -1.
//
// If an infinity is passed to Int, an error is returned.
func Int(num cty.Value) (cty.Value, error) {
	if num == cty.PositiveInfinity || num == cty.NegativeInfinity {
		return cty.NilVal, fmt.Errorf("can't truncate infinity to an integer")
	}
	return IntFunc.Call([]cty.Value{num})
}
