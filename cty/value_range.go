package cty

import (
	"fmt"
)

// Range returns an object that offers partial information about the range
// of the receiver.
//
// This is most relevant for unknown values, because it gives access to any
// optional additional constraints on the final value (specified by the source
// of the value using "refinements") beyond what we can assume from the value's
// type.
//
// Calling Range for a known value is a little strange, but it's supported by
// returning a [ValueRange] object that describes the exact value as closely
// as possible. Typically a caller should work directly with the exact value
// in that case, but some purposes might only need the level of detail
// offered by ranges and so can share code between both known and unknown
// values.
func (v Value) Range() ValueRange {
	// For an unknown value we just use its own refinements.
	if unk, isUnk := v.v.(*unknownType); isUnk {
		refinement := unk.refinement
		if refinement == nil {
			// We'll generate an unconstrained refinement, just to
			// simplify the code in ValueRange methods which can
			// therefore assume that there's always a refinement.
			refinement = &refinementNullable{isNull: tristateUnknown}
		}
		return ValueRange{v.Type(), refinement}
	}

	if v.IsNull() {
		// If we know a value is null then we'll just report that,
		// since no other refinements make sense for a definitely-null value.
		return ValueRange{
			v.Type(),
			&refinementNullable{isNull: tristateTrue},
		}
	}

	// For a known value we construct synthetic refinements that match
	// the value, just as a convenience for callers that want to share
	// codepaths between both known and unknown values.
	ty := v.Type()
	var synth unknownValRefinement
	switch {
	case ty == String:
		synth = &refinementString{
			prefix: v.AsString(),
		}
	case ty == Number:
		synth = &refinementNumber{
			min:    v,
			max:    v,
			minInc: true,
			maxInc: true,
		}
	case ty.IsCollectionType():
		synth = &refinementCollection{
			minLen: v.Length(),
			maxLen: v.Length(),
			minInc: true,
			maxInc: true,
		}
	default:
		// If we don't have anything else to say then we can at least
		// guarantee that the value isn't null.
		synth = &refinementNullable{}
	}

	// If we get down here then the value is definitely not null
	synth.setNull(tristateFalse)

	return ValueRange{ty, synth}
}

// ValueRange offers partial information about the range of a value.
//
// This is primarily interesting for unknown values, because it provides access
// to any additional known constraints (specified using "refinements") on the
// range of the value beyond what is represented by the value's type.
type ValueRange struct {
	ty  Type
	raw unknownValRefinement
}

// TypeConstraint returns a type constraint describing the value's type as
// precisely as possible with the available information.
func (r ValueRange) TypeConstraint() Type {
	return r.ty
}

// CouldBeNull returns true unless the value being described is definitely
// known to represent a non-null value.
func (r ValueRange) CouldBeNull() bool {
	if r.raw == nil {
		// A totally-unconstrained unknown value could be null
		return true
	}
	return r.raw.null() != tristateFalse
}

// DefinitelyNotNull returns true if there are no null values in the range.
func (r ValueRange) DefinitelyNotNull() bool {
	if r.raw == nil {
		// A totally-unconstrained unknown value could be null
		return false
	}
	return r.raw.null() == tristateFalse
}

// NumberLowerBound returns information about the lower bound of the range of
// a number value, or panics if the value is definitely not a number.
//
// If the value is nullable then the result represents the range of the number
// only if it turns out not to be null.
//
// The resulting value might itself be an unknown number if there is no
// known lower bound. In that case the "inclusive" flag is meaningless.
func (r ValueRange) NumberLowerBound() (min Value, inclusive bool) {
	if r.ty == DynamicPseudoType {
		// We don't even know if this is a number yet.
		return UnknownVal(Number), false
	}
	if r.ty != Number {
		panic(fmt.Sprintf("NumberLowerBound for %#v", r.ty))
	}
	if rfn, ok := r.raw.(*refinementNumber); ok && rfn.min != NilVal {
		return rfn.min, rfn.minInc
	}
	return UnknownVal(Number), false
}

// NumberUpperBound returns information about the upper bound of the range of
// a number value, or panics if the value is definitely not a number.
//
// If the value is nullable then the result represents the range of the number
// only if it turns out not to be null.
//
// The resulting value might itself be an unknown number if there is no
// known upper bound. In that case the "inclusive" flag is meaningless.
func (r ValueRange) NumberUpperBound() (max Value, inclusive bool) {
	if r.ty == DynamicPseudoType {
		// We don't even know if this is a number yet.
		return UnknownVal(Number), false
	}
	if r.ty != Number {
		panic(fmt.Sprintf("NumberUpperBound for %#v", r.ty))
	}
	if rfn, ok := r.raw.(*refinementNumber); ok && rfn.max != NilVal {
		return rfn.max, rfn.maxInc
	}
	return UnknownVal(Number), false
}

// StringPrefix returns a string that is guaranteed to be the prefix of
// the string value being described, or panics if the value is definitely not
// a string.
//
// If the value is nullable then the result represents the prefix of the string
// only if it turns out to not be null.
//
// If the resulting value is zero-length then the value could potentially be
// a string but it has no known prefix.
//
// cty.String values always contain normalized UTF-8 sequences; the result is
// also guaranteed to be a normalized UTF-8 sequence so the result also
// represents the exact bytes of the string value's prefix.
func (r ValueRange) StringPrefix() string {
	if r.ty == DynamicPseudoType {
		// We don't even know if this is a string yet.
		return ""
	}
	if r.ty != String {
		panic(fmt.Sprintf("StringPrefix for %#v", r.ty))
	}
	if rfn, ok := r.raw.(*refinementString); ok {
		return rfn.prefix
	}
	return ""
}

// LengthLowerBound returns information about the lower bound of the length of
// a collection-typed value, or panics if the value is definitely not a
// collection.
//
// If the value is nullable then the result represents the range of the length
// only if the value turns out not to be null.
//
// The resulting value might itself be an unknown number if there is no
// known lower bound. In that case the "inclusive" flag is meaningless.
func (r ValueRange) LengthLowerBound() (min Value, inclusive bool) {
	if r.ty == DynamicPseudoType {
		// We don't even know if this is a collection yet.
		return UnknownVal(Number), false
	}
	if !r.ty.IsCollectionType() {
		panic(fmt.Sprintf("LengthLowerBound for %#v", r.ty))
	}
	if rfn, ok := r.raw.(*refinementCollection); ok && rfn.minLen != NilVal {
		return rfn.minLen, rfn.minInc
	}
	return UnknownVal(Number), false
}

// LengthUpperBound returns information about the upper bound of the length of
// a collection-typed value, or panics if the value is definitely not a
// collection.
//
// If the value is nullable then the result represents the range of the length
// only if the value turns out not to be null.
//
// The resulting value might itself be an unknown number if there is no
// known upper bound. In that case the "inclusive" flag is meaningless.
func (r ValueRange) LengthUpperBound() (min Value, inclusive bool) {
	if r.ty == DynamicPseudoType {
		// We don't even know if this is a collection yet.
		return UnknownVal(Number), false
	}
	if !r.ty.IsCollectionType() {
		panic(fmt.Sprintf("LengthUpperBound for %#v", r.ty))
	}
	if rfn, ok := r.raw.(*refinementCollection); ok && rfn.maxLen != NilVal {
		return rfn.maxLen, rfn.maxInc
	}
	return UnknownVal(Number), false
}

// definitelyNotNull is a convenient helper for the common situation of checking
// whether a value could possibly be null.
//
// Returns true if the given value is either a known value that isn't null
// or an unknown value that has been refined to exclude null values from its
// range.
func definitelyNotNull(v Value) bool {
	if v.IsKnown() {
		return !v.IsNull()
	}
	return v.Range().DefinitelyNotNull()
}
