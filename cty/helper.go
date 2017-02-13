package cty

import (
	"fmt"
)

// anyUnknown is a helper to easily check if a set of values contains any
// unknowns, for operations that short-circuit to return unknown in that case.
func anyUnknown(values ...Value) bool {
	for _, val := range values {
		if val.v == unknown {
			return true
		}
	}
	return false
}

// typeCheck tests whether all of the given values belong to the given type.
// If the given types are a mixture of the given type and the dynamic
// pseudo-type then a short-circuit dynamic value is returned. If the given
// values are all of the correct type but at least one is unknown then
// a short-circuit unknown value is returned. If any other types appear then
// an error is returned. Otherwise (finally!) the result is nil, nil.
func typeCheck(ty Type, values ...Value) (shortCircuit *Value, err error) {
	hasDynamic := false
	hasUnknown := false

	for i, val := range values {
		if val.ty == DynamicPseudoType {
			hasDynamic = true
			continue
		}

		if !val.Type().Equals(ty) {
			return nil, fmt.Errorf(
				"type mismatch: want %s but value %d is %s",
				ty.FriendlyName(),
				i, val.ty.FriendlyName(),
			)
		}

		if val.v == unknown {
			hasUnknown = true
		}
	}

	if hasDynamic {
		return &DynamicVal, nil
	}

	if hasUnknown {
		ret := UnknownVal(ty)
		return &ret, nil
	}

	return nil, nil
}

// mustTypeCheck is a wrapper around typeCheck that immediately panics if
// any error is returned.
func mustTypeCheck(ty Type, values ...Value) *Value {
	shortCircuit, err := typeCheck(ty, values...)
	if err != nil {
		panic(err)
	}
	return shortCircuit
}
