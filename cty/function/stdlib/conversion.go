package stdlib

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strconv"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
)

// MakeToFunc constructs a "to..." function, like "tostring", which converts
// its argument to a specific type or type kind.
//
// The given type wantTy can be any type constraint that cty's "convert" package
// would accept. In particular, this means that you can pass
// cty.List(cty.DynamicPseudoType) to mean "list of any single type", which
// will then cause cty to attempt to unify all of the element types when given
// a tuple.
func MakeToFunc(wantTy cty.Type) function.Function {
	return function.New(&function.Spec{
		Description: fmt.Sprintf("Converts the given value to %s, or raises an error if that conversion is impossible.", wantTy.FriendlyName()),
		Params: []function.Parameter{
			{
				Name: "v",
				// We use DynamicPseudoType rather than wantTy here so that
				// all values will pass through the function API verbatim and
				// we can handle the conversion logic within the Type and
				// Impl functions. This allows us to customize the error
				// messages to be more appropriate for an explicit type
				// conversion, whereas the cty function system produces
				// messages aimed at _implicit_ type conversions.
				Type:             cty.DynamicPseudoType,
				AllowNull:        true,
				AllowDynamicType: true,
			},
		},
		Type: func(args []cty.Value) (cty.Type, error) {
			gotTy := args[0].Type()
			if gotTy.Equals(wantTy) {
				return wantTy, nil
			}
			conv := convert.GetConversionUnsafe(args[0].Type(), wantTy)
			if conv == nil {
				// We'll use some specialized errors for some trickier cases,
				// but most we can handle in a simple way.
				switch {
				case gotTy.IsTupleType() && wantTy.IsTupleType():
					return cty.NilType, function.NewArgErrorf(0, "incompatible tuple type for conversion: %s", convert.MismatchMessage(gotTy, wantTy))
				case gotTy.IsObjectType() && wantTy.IsObjectType():
					return cty.NilType, function.NewArgErrorf(0, "incompatible object type for conversion: %s", convert.MismatchMessage(gotTy, wantTy))
				default:
					return cty.NilType, function.NewArgErrorf(0, "cannot convert %s to %s", gotTy.FriendlyName(), wantTy.FriendlyNameForConstraint())
				}
			}
			// If a conversion is available then everything is fine.
			return wantTy, nil
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			// We didn't set "AllowUnknown" on our argument, so it is guaranteed
			// to be known here but may still be null.
			ret, err := convert.Convert(args[0], retType)
			if err != nil {
				// Because we used GetConversionUnsafe above, conversion can
				// still potentially fail in here. For example, if the user
				// asks to convert the string "a" to bool then we'll
				// optimistically permit it during type checking but fail here
				// once we note that the value isn't either "true" or "false".
				gotTy := args[0].Type()
				switch {
				case gotTy == cty.String && wantTy == cty.Bool:
					what := "string"
					if !args[0].IsNull() {
						what = strconv.Quote(args[0].AsString())
					}
					return cty.NilVal, function.NewArgErrorf(0, `cannot convert %s to bool; only the strings "true" or "false" are allowed`, what)
				case gotTy == cty.String && wantTy == cty.Number:
					what := "string"
					if !args[0].IsNull() {
						what = strconv.Quote(args[0].AsString())
					}
					return cty.NilVal, function.NewArgErrorf(0, `cannot convert %s to number; given string must be a decimal representation of a number`, what)
				default:
					return cty.NilVal, function.NewArgErrorf(0, "cannot convert %s to %s", gotTy.FriendlyName(), wantTy.FriendlyNameForConstraint())
				}
			}
			return ret, nil
		},
	})
}

// ToUnionFunc is a function which takes an object-typed value and returns
// a union-typed value based on it.
//
// The given object may have any number of attributes of arbitrary types
// but exactly one of them must be non-null and all others must be null.
// The one that is not null then becomes the selected variant of the resulting
// union value.
//
// This function automatically infers a union type based on the given object
// type. To convert an object type to a specific, predefined union type use
// [MakeToFunc] with that union type instead of using this function.
var ToUnionFunc = function.New(&function.Spec{
	Description: "Constructs a union-typed value based on an example object-typed value.",
	Params: []function.Parameter{
		{
			Name: "variants",
			Type: cty.DynamicPseudoType,
		},
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		arg := args[0]
		objTy := arg.Type()
		if objTy.HasDynamicTypes() {
			if arg.IsWhollyKnown() {
				// If there are no unknown values but still unknown types
				// then we must have some unknown-typed nulls or empty
				// collections with unknown element types, in which case
				// we cannot proceed.
				return cty.NilType, function.NewArgErrorf(0, "object contains null values of unknown type and/or collections of unknown element type")
			}
			// If we have some unknown values then we'll optimistically assume
			// that we'll get more type information once those unknown values
			// are resolved, but we cannot predict our result type yet.
			return cty.DynamicPseudoType, nil
		}
		if !objTy.IsObjectType() {
			return cty.NilType, function.NewArgErrorf(0, "must be of an object type")
		}
		// If all of the preconditions hold then we can just directly
		// translate our object attributes into union variants. We'll
		// check whether the attribute _values_ are suitable in the Impl
		// function.
		atys := objTy.AttributeTypes()
		if len(atys) == 0 {
			return cty.NilType, function.NewArgErrorf(0, "object must have at least one attribute")
		}
		return cty.Union(atys), nil
	},
	RefineResult: refineNonNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		if retType == cty.DynamicPseudoType {
			// Our Type function decided it doesn't have enough information
			// yet, so we don't either.
			return cty.DynamicVal, nil
		}
		// Otherwise, we should definitely now have a union type and
		// so our goal is to decide which variant to set based on which
		// single attribute is set in the given object.
		arg := args[0]
		allVariants := slices.Collect(maps.Keys(retType.UnionVariants()))
		sort.Strings(allVariants)

		var variantName string
		var variantVal cty.Value
		for _, attrName := range allVariants {
			attrVal := arg.GetAttr(attrName)
			rng := attrVal.Range()
			if rng.DefinitelyNotNull() {
				if variantVal != cty.NilVal {
					return cty.NilVal, function.NewArgErrorf(0, "cannot set both %q and %q variants in union", variantName, attrName)
				}
				variantName = attrName
				variantVal = attrVal
				continue
			}
			if !attrVal.IsKnown() {
				// We don't yet have enough information to decide whether
				// this attribute is or is not null, so we cannot return
				// a known result yet.
				return cty.UnknownVal(retType), nil
			}
		}
		if variantVal == cty.NilVal {
			// We didn't find any non-null attribute value, so we can't
			// produce a valid union value.
			return cty.NilVal, function.NewArgErrorf(0, "must set exactly one attribute to a non-null value")
		}
		return cty.UnionVal(retType, variantName, variantVal), nil
	},
})

func ToUnion(variants cty.Value) (cty.Value, error) {
	return ToUnionFunc.Call([]cty.Value{variants})
}

// AssertNotNullFunc is a function which does nothing except return an error
// if the argument given to it is null.
//
// This could be useful in some cases where the automatic refinment of
// nullability isn't precise enough, because the result is guaranteed to not
// be null and can therefore allow downstream comparisons to null to return
// a known value even if the value is otherwise unknown.
var AssertNotNullFunc = function.New(&function.Spec{
	Description: "Returns the given value varbatim if it is non-null, or raises an error if it's null.",
	Params: []function.Parameter{
		{
			Name: "v",
			Type: cty.DynamicPseudoType,
			// NOTE: We intentionally don't set AllowNull here, and so
			// the function system will automatically reject a null argument
			// for us before calling Impl.
		},
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		return args[0].Type(), nil
	},
	RefineResult: refineNonNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		// Our argument doesn't set AllowNull: true, so we're guaranteed to
		// have a non-null value in args[0].
		return args[0], nil
	},
})

func AssertNotNull(v cty.Value) (cty.Value, error) {
	return AssertNotNullFunc.Call([]cty.Value{v})
}
