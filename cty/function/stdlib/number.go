package stdlib

import (
	"fmt"
	"math/big"

	"github.com/apparentlymart/go-cty/cty"
	"github.com/apparentlymart/go-cty/cty/function"
)

var AddFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "a",
			Type:             cty.Number,
			AllowDynamicType: true,
		},
		{
			Name:             "b",
			Type:             cty.Number,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.Number),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		return args[0].Add(args[1]), nil
	},
})

var SubtractFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "a",
			Type:             cty.Number,
			AllowDynamicType: true,
		},
		{
			Name:             "b",
			Type:             cty.Number,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.Number),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		return args[0].Subtract(args[1]), nil
	},
})

var MultiplyFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "a",
			Type:             cty.Number,
			AllowDynamicType: true,
		},
		{
			Name:             "b",
			Type:             cty.Number,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.Number),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		return args[0].Multiply(args[1]), nil
	},
})

var DivideFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "a",
			Type:             cty.Number,
			AllowDynamicType: true,
		},
		{
			Name:             "b",
			Type:             cty.Number,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.Number),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		return args[0].Divide(args[1]), nil
	},
})

var ModuloFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "a",
			Type:             cty.Number,
			AllowDynamicType: true,
		},
		{
			Name:             "b",
			Type:             cty.Number,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.Number),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		return args[0].Modulo(args[1]), nil
	},
})

var NegateFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "num",
			Type:             cty.Number,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.Number),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		return args[0].Negate(), nil
	},
})

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

// Add returns the sum of the two given numbers.
func Add(a cty.Value, b cty.Value) (cty.Value, error) {
	return AddFunc.Call([]cty.Value{a, b})
}

// Subtract returns the difference between the two given numbers.
func Subtract(a cty.Value, b cty.Value) (cty.Value, error) {
	return SubtractFunc.Call([]cty.Value{a, b})
}

// Multiply returns the product of the two given numbers.
func Multiply(a cty.Value, b cty.Value) (cty.Value, error) {
	return MultiplyFunc.Call([]cty.Value{a, b})
}

// Divide returns a divided by b, where both a and b are numbers.
func Divide(a cty.Value, b cty.Value) (cty.Value, error) {
	return DivideFunc.Call([]cty.Value{a, b})
}

// Negate returns the given number multipled by -1.
func Negate(num cty.Value) (cty.Value, error) {
	return NegateFunc.Call([]cty.Value{num})
}

// Modulo returns the remainder of a divided by b under integer division,
// where both a and b are numbers.
func Modulo(a cty.Value, b cty.Value) (cty.Value, error) {
	return ModuloFunc.Call([]cty.Value{a, b})
}

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
