package cty

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/apparentlymart/go-cty/cty/set"
)

func (val Value) GoString() string {
	if val == NilVal {
		return "cty.NilVal"
	}

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

	switch {
	case val.ty.IsSetType():
		vals := val.v.(set.Set).Values()
		if vals == nil || len(vals) == 0 {
			return fmt.Sprintf("cty.SetValEmpty()")
		} else {
			return fmt.Sprintf("cty.SetVal(%#v)", vals)
		}
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
	case ty.IsSetType():
		s1 := val.v.(set.Set)
		s2 := other.v.(set.Set)
		equal := true

		// Note that by our definition of sets it's never possible for two
		// sets that contain unknown values (directly or indicrectly) to
		// ever be equal, even if they are otherwise identical.

		// FIXME: iterating both lists and checking each item is not the
		// ideal implementation here, but it works with the primitives we
		// have in the set implementation. Perhaps the set implementation
		// can provide its own equality test later.
		s1.EachValue(func(v interface{}) {
			if !s2.Has(v) {
				equal = false
			}
		})
		s2.EachValue(func(v interface{}) {
			if !s1.Has(v) {
				equal = false
			}
		})

		result = equal
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

// GetAttr returns the value of the given attribute of the receiver, which
// must be of an object type that has an attribute of the given name.
// This method will panic if the receiver type is not compatible.
//
// The method will also panic if the given attribute name is not defined
// for the value's type. Use the attribute-related methods on Type to
// check for the validity of an attribute before trying to use it.
//
// This method may be called on a value whose type is DynamicPseudoType,
// in which case the result will also be DynamicVal.
func (val Value) GetAttr(name string) Value {
	if val.ty == DynamicPseudoType {
		return DynamicVal
	}

	if !val.ty.IsObjectType() {
		panic("value is not an object")
	}
	if !val.ty.HasAttribute(name) {
		panic("value has no attribute of that name")
	}

	return Value{
		ty: val.ty.AttributeType(name),
		v:  val.v.(map[string]interface{})[name],
	}
}

// Index returns the value of an element of the receiver, which must be
// either a map or a list. This method will panic if the receiver type is
// not compatible.
//
// The key value must be the correct type for the receving collection: a
// number if the collection is a list or a string if it is a map.
// In the case of a list, the given number must be convertable to int or this
// method will panic. The key may alternatively be of DynamicPseudoType, in
// which case the result itself is DynamicValue.
//
// The result is of the receiver collection's element type.
//
// This method may be called on a value whose type is DynamicPseudoType,
// in which case the result will also be the DynamicValue.
func (val Value) Index(key Value) Value {
	panic("Index not yet implemented")
}

// ForEachElement executes a given callback function for each element of
// the receiver, which must be a collection type or this method will panic.
//
// If the receiver is of a list type, the key passed to to the callback
// will be of type Number and the value will be of the list's element type.
//
// If the receiver is of a map type, the key passed to the callback will
// be of type String and the value will be of the map's element type.
// Elements are passed in ascending lexicographical order by key.
//
// If the receiver is of a set type, the key passed to the callback will be
// NilVal and should be disregarded. Elements are passed in an undefined but
// consistent order.
//
// Returns true if the iteration exited early due to the callback function
// returning true, or false if the loop ran to completion.
//
// ForEachElement is an integration method, so it cannot handle Unknown
// values. This method will panic if the receiver is Unknown.
func (val Value) ForEachElement(cb ElementIterator) bool {
	switch {
	case val.ty.IsListType():
		ety := val.ty.ElementType()

		for i, rawVal := range val.v.([]interface{}) {
			stop := cb(NumberIntVal(int64(i)), Value{
				ty: ety,
				v:  rawVal,
			})
			if stop {
				return true
			}
		}
		return false
	case val.ty.IsMapType():
		ety := val.ty.ElementType()

		// We iterate the keys in a predictable lexicographical order so
		// that results will always be stable given the same input map.
		rawMap := val.v.(map[string]interface{})
		keys := make([]string, 0, len(rawMap))
		for key := range rawMap {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			rawVal := rawMap[key]
			stop := cb(StringVal(key), Value{
				ty: ety,
				v:  rawVal,
			})
			if stop {
				return true
			}
		}
		return false
	case val.ty.IsSetType():
		ety := val.ty.ElementType()

		rawSet := val.v.(set.Set)
		stop := false
		rawSet.EachValue(func(ev interface{}) {
			if stop {
				return
			}
			stop = cb(NilVal, Value{
				ty: ety,
				v:  ev,
			})
		})
		return stop
	default:
		panic("ForEachElement on non-collection type")
	}
}
