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
// that from a Conversion value itself.
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
