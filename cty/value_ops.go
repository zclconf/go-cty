package cty

import (
	"fmt"
	"math/big"
)

func (val Value) GoString() string {
	if val.ty == DynamicPseudoType {
		return "cty.DynamicValue"
	}

	if !val.IsKnown() {
		return fmt.Sprintf("cty.Unknown(%#v)", val.ty)
	}
	if val.IsNull() {
		return fmt.Sprintf("cty.Null(%#v)", val.ty)
	}

	// By the time we reach here we've dealt with all of the exceptions around
	// unknowns and nulls, so we're guaranteed that the values are the
	// canonical internal representation of the given type.

	switch val.ty {
	case Bool:
		return fmt.Sprintf("cty.BoolVal(%#v)", val.v)
	case Number:
		fv := val.v.(*big.Float)
		// We'll try to use NumberIntVal or NumberFloatVal if we can, since
		// the fully-general initializer call is pretty ugly-looking.
		if fv.IsInt() {
			return fmt.Sprintf("cty.NumberIntVal(%#v)", fv)
		}
		if rfv, accuracy := fv.Float64(); accuracy == big.Exact {
			return fmt.Sprintf("cty.NumberFloatVal(%#v)", rfv)
		}
		return fmt.Sprintf("cty.NumberVal(new(big.Float).Parse(\"%#v\", 10))", fv)
	case String:
		return fmt.Sprintf("cty.StringVal(%#v)", val.v)
	}

	// Default exposes implementation details, so should actually cover
	// all of the cases above for good caller UX.
	return fmt.Sprintf("cty.Value{ty: %#v, v: %#v}", val.ty, val.v)
}

// Equals returns True if the receiver and the given other value have the
// same type and are exactly equal in value.
//
// The usual short-circuit rules apply, so the result can be unknown or typed
// as dynamic if either of the given values are. Use RawEquals to compare
// if two values are equal *ignoring* the short-circuit rules.
func (val Value) Equals(other Value) Value {
	if val.ty == DynamicPseudoType || other.ty == DynamicPseudoType {
		return UnknownVal(Bool)
	}

	if !val.ty.Equals(other.ty) {
		return BoolVal(false)
	}

	if !(val.IsKnown() && other.IsKnown()) {
		return UnknownVal(Bool)
	}

	if val.IsNull() || other.IsNull() {
		if val.IsNull() && other.IsNull() {
			return BoolVal(true)
		}
		return BoolVal(false)
	}

	ty := val.ty
	result := false

	switch {
	case ty == Number:
		result = val.v.(*big.Float).Cmp(other.v.(*big.Float)) == 0
	case ty == Bool:
		result = val.v.(bool) == other.v.(bool)
	case ty == String:
		// Simple equality is safe because we NFC-normalize strings as they
		// enter our world from StringVal, and so we can assume strings are
		// always in normal form.
		result = val.v.(string) == other.v.(string)
	case ty.IsObjectType():
		oty := ty.typeImpl.(typeObject)
		result = true
		for attr, aty := range oty.attrTypes {
			lhs := Value{
				ty: aty,
				v:  val.v.(map[string]interface{})[attr],
			}
			rhs := Value{
				ty: aty,
				v:  other.v.(map[string]interface{})[attr],
			}
			eq := lhs.Equals(rhs)
			if !eq.IsKnown() {
				return UnknownVal(Bool)
			}
			if eq.False() {
				result = false
				break
			}
		}
	case ty.IsListType():
		ety := ty.typeImpl.(typeList).elementType
		if len(val.v.([]interface{})) == len(other.v.([]interface{})) {
			result = true
			for i := range val.v.([]interface{}) {
				lhs := Value{
					ty: ety,
					v:  val.v.([]interface{})[i],
				}
				rhs := Value{
					ty: ety,
					v:  other.v.([]interface{})[i],
				}
				eq := lhs.Equals(rhs)
				if !eq.IsKnown() {
					return UnknownVal(Bool)
				}
				if eq.False() {
					result = false
					break
				}
			}
		}
	case ty.IsMapType():
		ety := ty.typeImpl.(typeMap).elementType
		if len(val.v.(map[string]interface{})) == len(other.v.(map[string]interface{})) {
			result = true
			for k := range val.v.(map[string]interface{}) {
				if _, ok := other.v.(map[string]interface{})[k]; !ok {
					result = false
					break
				}
				lhs := Value{
					ty: ety,
					v:  val.v.(map[string]interface{})[k],
				}
				rhs := Value{
					ty: ety,
					v:  other.v.(map[string]interface{})[k],
				}
				eq := lhs.Equals(rhs)
				if !eq.IsKnown() {
					return UnknownVal(Bool)
				}
				if eq.False() {
					result = false
					break
				}
			}
		}

	default:
		// should never happen
		panic(fmt.Errorf("unsupported value type %#v in Equals", ty))
	}

	return BoolVal(result)
}

// True returns true if the receiver is True, false if False, and panics if
// the receiver is not of type Bool.
//
// This is a helper function to help write application logic that works with
// values, rather than a first-class operation. It does not work with unknown
// or null values. For more robust handling with unknown value
// short-circuiting, use val.Equals(cty.True).
func (val Value) True() bool {
	if val.ty != Bool {
		panic("not bool")
	}
	return val.Equals(True).v.(bool)
}

// False is the opposite of True.
func (val Value) False() bool {
	return !val.True()
}

// RawEquals returns true if and only if the two given values have the same
// type and equal value, ignoring the usual short-circuit rules about
// unknowns and dynamic types.
//
// This method is more appropriate for testing than for real use, since it
// skips over usual semantics around unknowns but as a consequence allows
// testing the result of another operation that is expected to return unknown.
// It returns a primitive Go bool rather than a Value to remind us that it
// is not a first-class value operation.
func (val Value) RawEquals(other Value) bool {
	// First some exceptions to skip over the short-circuit behavior we'd
	// normally expect, thus ensuring we can call Equals and reliably get
	// back a known Bool.
	if !val.ty.Equals(other.ty) {
		return false
	}
	if (!val.IsKnown()) && (!other.IsKnown()) {
		return true
	}
	if (val.IsKnown() && !other.IsKnown()) || (other.IsKnown() && !val.IsKnown()) {
		return false
	}
	if val.ty == DynamicPseudoType && other.ty == DynamicPseudoType {
		return true
	}

	result := val.Equals(other)
	return result.v.(bool)
}

// Add returns the sum of the receiver and the given other value. Both values
// must be numbers; this method will panic if not.
func (val Value) Add(other Value) Value {
	if shortCircuit := mustTypeCheck(Number, val, other); shortCircuit != nil {
		return *shortCircuit
	}

	ret := new(big.Float)
	ret.Add(val.v.(*big.Float), other.v.(*big.Float))
	return NumberVal(ret)
}

// Sub returns receiver minus the given other value. Both values must be
// numbers; this method will panic if not.
func (val Value) Sub(other Value) Value {
	if shortCircuit := mustTypeCheck(Number, val, other); shortCircuit != nil {
		return *shortCircuit
	}

	return val.Add(other.Neg())
}

// Neg returns the numeric negative of the receiver, which must be a number.
// This method will panic when given a value of any other type.
func (val Value) Neg() Value {
	if shortCircuit := mustTypeCheck(Number, val); shortCircuit != nil {
		return *shortCircuit
	}

	ret := new(big.Float).Neg(val.v.(*big.Float))
	return NumberVal(ret)
}
