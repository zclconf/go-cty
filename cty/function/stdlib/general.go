package stdlib

import (
	"fmt"
	"maps"
	"slices"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
)

var EqualFunc = function.New(&function.Spec{
	Description: `Returns true if the two given values are equal, or false otherwise.`,
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
	Type:         function.StaticReturnType(cty.Bool),
	RefineResult: refineNonNull,
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		return args[0].Equals(args[1]), nil
	},
})

var NotEqualFunc = function.New(&function.Spec{
	Description: `Returns false if the two given values are equal, or true otherwise.`,
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
	Type:         function.StaticReturnType(cty.Bool),
	RefineResult: refineNonNull,
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		return args[0].Equals(args[1]).Not(), nil
	},
})

var CoalesceFunc = function.New(&function.Spec{
	Description: `Returns the first of the given arguments that isn't null, or raises an error if there are no non-null arguments.`,
	Params:      []function.Parameter{},
	VarParam: &function.Parameter{
		Name:             "vals",
		Type:             cty.DynamicPseudoType,
		AllowUnknown:     true,
		AllowDynamicType: true,
		AllowNull:        true,
		AllowMarked:      true,
	},
	Type: func(args []cty.Value) (ret cty.Type, err error) {
		argTypes := make([]cty.Type, len(args))
		for i, val := range args {
			argTypes[i] = val.Type()
		}
		retType, _ := convert.UnifyUnsafe(argTypes)
		if retType == cty.NilType {
			return cty.NilType, fmt.Errorf("all arguments must have the same type")
		}
		return retType, nil
	},
	RefineResult: refineNonNull,
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		var extraMarks cty.ValueMarks
		defer func() {
			if len(extraMarks) != 0 && ret != cty.NilVal {
				ret = ret.WithMarks(extraMarks)
			}
		}()
		for i, argVal := range args {
			unmarkedVal, marks := argVal.Unmark()
			if argVal.IsNull() {
				if extraMarks == nil {
					extraMarks = make(cty.ValueMarks)
				}
				maps.Copy(extraMarks, marks)
				continue
			}
			rng := unmarkedVal.Range()
			if rng.DefinitelyNotNull() {
				return convert.Convert(argVal, retType)
			}
			// If we get here then we don't know whether this argument is
			// null or not. Therefore our result could potentially be this
			// or any subsequent value that could potentially be null, until
			// we find one that is definitely not null.
			remaining := args[i:]
			for i, argVal := range remaining {
				unmarkedVal, _ := argVal.Unmark()
				if unmarkedVal.Range().DefinitelyNotNull() {
					remaining = remaining[:i+1]
					return convert.Convert(cty.UnknownChoice(slices.Values(remaining), false), retType)
				}
			}
			return convert.Convert(cty.UnknownChoice(slices.Values(remaining), true), retType)
		}
		return cty.NilVal, fmt.Errorf("no non-null arguments")
	},
})

func refineNonNull(b *cty.RefinementBuilder) *cty.RefinementBuilder {
	return b.NotNull()
}

// Equal determines whether the two given values are equal, returning a
// bool value.
func Equal(a cty.Value, b cty.Value) (cty.Value, error) {
	return EqualFunc.Call([]cty.Value{a, b})
}

// NotEqual is the opposite of Equal.
func NotEqual(a cty.Value, b cty.Value) (cty.Value, error) {
	return NotEqualFunc.Call([]cty.Value{a, b})
}

// Coalesce returns the first of the given arguments that is not null. If
// all arguments are null, an error is produced.
func Coalesce(vals ...cty.Value) (cty.Value, error) {
	return CoalesceFunc.Call(vals)
}
