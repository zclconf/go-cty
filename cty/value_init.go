package cty

import (
	"math/big"
)

// BoolVal returns a Value of type Number whose internal value is the given
// bool.
func BoolVal(v bool) Value {
	return Value{
		ty: Bool,
		v:  v,
	}
}

// NumberVal returns a Value of type Number whose internal value is the given
// big.Float. The returned value becomes the owner of the big.Float object,
// and so it's forbidden for the caller to mutate the object after it's
// wrapped in this way.
func NumberVal(v *big.Float) Value {
	return Value{
		ty: Number,
		v:  v,
	}
}

// NumberIntVal returns a Value of type Number whose internal value is equal
// to the given integer.
func NumberIntVal(v int64) Value {
	return NumberVal(new(big.Float).SetInt64(v))
}

// NumberFloatVal returns a Value of type Number whose internal value is
// equal to the given float.
func NumberFloatVal(v float64) Value {
	return NumberVal(new(big.Float).SetFloat64(v))
}

// StringVal returns a Value of type String whose internal value is the
// given string.
func StringVal(v string) Value {
	return Value{
		ty: String,
		v:  v,
	}
}

// ObjectVal returns a Value of an object type whose structure is defined
// by the key names and value types in the given map.
func ObjectVal(attrs map[string]Value) Value {
	attrTypes := make(map[string]Type, len(attrs))
	attrVals := make(map[string]interface{}, len(attrs))

	for attr, val := range attrs {
		attrTypes[attr] = val.ty
		attrVals[attr] = val.v
	}

	return Value{
		ty: Object(attrTypes),
		v:  attrVals,
	}
}
