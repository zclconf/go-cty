package convert

import (
	"fmt"

	"github.com/apparentlymart/go-cty/cty"
)

// This file contains the public interface of this package, which is intended
// to be a small, convenient interface designed for easy integration into
// a hypothetical language type checker and interpreter.

// Conversion is a named function type representing a conversion from a
// value of one type to a value of another type.
//
// The source type for a conversion is always the source type given to
// the function that returned the Conversion, but there is no way to recover
// that from a Conversion value itself. If a Conversion is given a value
// that is not of its expected type (with the exception of DynamicPseudoType,
// which is always supported) then the function may panic or produce undefined
// results.
type Conversion func(in cty.Value) (out cty.Value, err error)

// GetConversion returns a Conversion between the given in and out Types if
// a safe one is available, or returns nil otherwise.
func GetConversion(in cty.Type, out cty.Type) Conversion {
	return nil
}

// GetConversionUnsafe returns a Conversion between the given in and out Types
// if either a safe or unsafe one is available, or returns nil otherwise.
func GetConversionUnsafe(in cty.Type, out cty.Type) Conversion {
	safe := GetConversion(in, out)
	if safe != nil {
		return safe
	}

	return nil
}

// Convert returns the result of converting the given value to the given type
// if an safe or unsafe conversion is available, or returns an error if such a
// conversion is impossible.
//
// This is a convenience wrapper around calling GetConversionUnsafe and then
// immediately passing the given value to the resulting function.
func Convert(in cty.Value, want cty.Type) (cty.Value, error) {
	conv := GetConversionUnsafe(in.Type(), want)
	if conv == nil {
		return cty.NilVal, fmt.Errorf("incorrect value type; %s required", want.FriendlyName())
	}
	return conv(in)
}

// CompareTypes returns a value that defines the partial order of types.
// If -1 is returned then Type a is a supertype of Type b. If 1 is returned
// then the converse is true. If 0 is returned then the two types have no
// such relationship.
//
// In cty the subtype/supertype relationships are somewhat loose and result
// from the availability of type conversions. The availability of a *safe*
// type conversion from a to b makes Type b a supertype of Type a. Conversely,
// the availability of an *unsafe* conversion makes Type b a *subtype* of
// Type a.
//
// cty.DynamicPseudoType is, as usual, a special case: it is treated as the
// universal supertype for comparison purposes, but since it is used as a type
// *placeholder* rather than as an actual type, callers seeking the closest
// common subtype of a set of types should disregard DynamicPseudoType as
// the solution unless it is the *only* type present.
func CompareTypes(a cty.Type, b cty.Type) int {
	return 0
}

// Unify attempts to find a common supertype of the given types. If this is
// possible, that type is returned along with a slice of necessary conversions
// for some of the given types.
//
// If no common supertype can be found, this function returns cty.NilType and
// a nil slice.
//
// If a common supertype *can* be found, the returned slice will always be
// non-nil and will contain a non-nil conversion for each given type that
// needs to be converted, with indices corresponding to the input slice.
// Any given type that does *not* need conversion (because it is already of
// the appropriate type) will have a nil Conversion.
//
// cty.DynamicPseudoType is, as usual, a special case. If the given type list
// contains a mixture of dynamic and non-dynamic types, the dynamic types are
// disregarded for type selection and a conversion is returned for them that
// will attempt a late conversion of the given value to the target type,
// failing with a conversion error if the eventual concrete type is not
// compatible. If *all* given types are DynamicPseudoType, or in the
// degenerate case of an empty slice of types, the returned type is itself
// cty.DynamicPseudoType and no conversions are attempted.
func Unify(types []cty.Type) (cty.Type, []Conversion) {
	return cty.NilType, nil
}
