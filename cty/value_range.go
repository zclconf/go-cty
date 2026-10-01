package cty

import (
	"fmt"
	"math"
	"strings"
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
	if v.IsMarked() {
		panic("Value.Range on marked value; must be unmarked first")
	}

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
		if lenVal := v.Length(); lenVal.IsKnown() {
			l, _ := lenVal.AsBigFloat().Int64()
			synth = &refinementCollection{
				minLen: int(l),
				maxLen: int(l),
			}
		} else {
			synth = &refinementCollection{
				minLen: 0,
				maxLen: math.MaxInt,
			}
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

// AsValue returns an value that coulod be used as a placeholder for any
// value included in the range.
//
// The result is usually at least partially unknown, but can potentially be
// fully known in cases where the range is tightly constrained enough that
// we can prove there's only one possible value.
//
// ValueRange has limited precision in what it can track for different types
// of value, so round-tripping through [Value.Range] and [ValueRange.AsValue]
// will typically lose information.
func (r ValueRange) AsValue() Value {
	return refinedValue(UnknownVal(r.ty), r.raw)
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
		if !rfn.min.IsKnown() {
			return NegativeInfinity, true
		}
		return rfn.min, rfn.minInc
	}
	return NegativeInfinity, false
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
		if !rfn.max.IsKnown() {
			return PositiveInfinity, true
		}
		return rfn.max, rfn.maxInc
	}
	return PositiveInfinity, false
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
func (r ValueRange) LengthLowerBound() int {
	if r.ty == DynamicPseudoType {
		// We don't even know if this is a collection yet.
		return 0
	}
	if !r.ty.IsCollectionType() {
		panic(fmt.Sprintf("LengthLowerBound for %#v", r.ty))
	}
	if rfn, ok := r.raw.(*refinementCollection); ok {
		return rfn.minLen
	}
	return 0
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
func (r ValueRange) LengthUpperBound() int {
	if r.ty == DynamicPseudoType {
		// We don't even know if this is a collection yet.
		return math.MaxInt
	}
	if !r.ty.IsCollectionType() {
		panic(fmt.Sprintf("LengthUpperBound for %#v", r.ty))
	}
	if rfn, ok := r.raw.(*refinementCollection); ok {
		return rfn.maxLen
	}
	return math.MaxInt
}

// Includes determines whether the given value is in the receiving range.
//
// It can return only three possible values:
//   - [cty.True] if the range definitely includes the value
//   - [cty.False] if the range definitely does not include the value
//   - An unknown value of [cty.Bool] if there isn't enough information to decide.
//
// This function is not fully comprehensive: it may return an unknown value
// in some cases where a definitive value could be computed in principle, and
// those same situations may begin returning known values in later releases as
// the rules are refined to be more complete. Currently the rules focus mainly
// on answering [cty.False], because disproving membership tends to be more
// useful than proving membership.
func (r ValueRange) Includes(v Value) Value {
	unknownResult := UnknownVal(Bool).RefineNotNull()

	if r.raw.null() == tristateTrue {
		if v.IsNull() {
			return True
		} else {
			return False
		}
	}
	if r.raw.null() == tristateFalse {
		if v.IsNull() {
			return False
		}
		// A definitely-not-null value could potentially match
		// but we won't know until we do some more checks below.
	}
	// If our range includes both null and non-null values and the value is
	// null then it's definitely in range.
	if v.IsNull() {
		return True
	}
	if len(v.Type().TestConformance(r.TypeConstraint())) != 0 {
		// If the value doesn't conform to the type constraint then it's
		// definitely not in the range.
		return False
	}
	if v.Type() == DynamicPseudoType {
		// If it's an unknown value of an unknown type then there's no
		// further tests we can make.
		return unknownResult
	}

	switch r.raw.(type) {
	case *refinementString:
		if v.IsKnown() {
			prefix := r.StringPrefix()
			got := v.AsString()

			if !strings.HasPrefix(got, prefix) {
				return False
			}
		}
	case *refinementCollection:
		lenVal := v.Length()
		minLen := NumberIntVal(int64(r.LengthLowerBound()))
		maxLen := NumberIntVal(int64(r.LengthUpperBound()))
		if minOk := lenVal.GreaterThanOrEqualTo(minLen); minOk.IsKnown() && minOk.False() {
			return False
		}
		if maxOk := lenVal.LessThanOrEqualTo(maxLen); maxOk.IsKnown() && maxOk.False() {
			return False
		}
	case *refinementNumber:
		minVal, minInc := r.NumberLowerBound()
		maxVal, maxInc := r.NumberUpperBound()
		var minOk, maxOk Value
		if minInc {
			minOk = v.GreaterThanOrEqualTo(minVal)
		} else {
			minOk = v.GreaterThan(minVal)
		}
		if maxInc {
			maxOk = v.LessThanOrEqualTo(maxVal)
		} else {
			maxOk = v.LessThan(maxVal)
		}
		if minOk.IsKnown() && minOk.False() {
			return False
		}
		if maxOk.IsKnown() && maxOk.False() {
			return False
		}
	}

	// If we fall out here then we don't have enough information to decide.
	return unknownResult
}

// numericRangeArithmetic is a helper we use to calculate derived numeric ranges
// for arithmetic on refined numeric values.
//
// op must be a monotone operation. numericRangeArithmetic adapts that operation
// into the equivalent interval arithmetic operation.
//
// The result is a superset of the range of the given operation against the
// given input ranges, if it's possible to calculate that without encountering
// an invalid operation. Currently the result is inexact due to ignoring
// the inclusiveness of the input bounds and just always returning inclusive
// bounds.
func numericRangeArithmetic(op func(a, b Value) Value, a, b ValueRange) func(*RefinementBuilder) *RefinementBuilder {
	wrapOp := func(a, b Value) (ret Value) {
		// Our functions have various panicking edge cases involving incompatible
		// uses of infinities. To keep things simple here we'll catch those
		// and just return an unconstrained number.
		defer func() {
			if v := recover(); v != nil {
				ret = UnknownVal(Number)
			}
		}()
		return op(a, b)
	}

	return func(builder *RefinementBuilder) *RefinementBuilder {
		aMin, _ := a.NumberLowerBound()
		aMax, _ := a.NumberUpperBound()
		bMin, _ := b.NumberLowerBound()
		bMax, _ := b.NumberUpperBound()

		v1 := wrapOp(aMin, bMin)
		v2 := wrapOp(aMin, bMax)
		v3 := wrapOp(aMax, bMin)
		v4 := wrapOp(aMax, bMax)

		newMin := mostNumberValue(Value.LessThan, v1, v2, v3, v4)
		newMax := mostNumberValue(Value.GreaterThan, v1, v2, v3, v4)

		if isInf := newMin.Equals(NegativeInfinity); isInf.IsKnown() && isInf.False() {
			builder = builder.NumberRangeLowerBound(newMin, true)
		}
		if isInf := newMax.Equals(PositiveInfinity); isInf.IsKnown() && isInf.False() {
			builder = builder.NumberRangeUpperBound(newMax, true)
		}
		return builder
	}
}

func mostNumberValue(op func(i, j Value) Value, v1 Value, vN ...Value) Value {
	r := v1
	for _, v := range vN {
		more := op(v, r)
		if !more.IsKnown() {
			return UnknownVal(Number)
		}
		if more.True() {
			r = v
		}
	}
	return r
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

// commonRange returns a [ValueRange] that contains everything that is included
// in the two other given value ranges, while attempting to return the tightest
// possible range.
func commonRange(a, b ValueRange) ValueRange {
	if !a.ty.Equals(b.ty) || a.raw == nil || b.raw == nil {
		return ValueRange{
			ty:  DynamicPseudoType,
			raw: nil, // no refinements at all, because only known-typed values can have refinements
		}
	}

	var rfn unknownValRefinement
	commonNullable := tristateUnknown
	if aNullable, bNullable := a.raw.null(), b.raw.null(); aNullable == bNullable {
		commonNullable = aNullable
	}
	nullable := refinementNullable{
		isNull: commonNullable,
	}
	switch aRfn := a.raw.(type) {
	case *refinementString:
		bRfn, ok := b.raw.(*refinementString)
		if !ok {
			break
		}
		if aRfn.prefix == bRfn.prefix {
			rfn = &refinementString{
				prefix:             aRfn.prefix,
				refinementNullable: nullable,
			}
			break
		}
		if commonLen := min(len(aRfn.prefix), len(bRfn.prefix)); commonLen != 0 {
			// The following works with bytes, rather than UTF-8 sequences or
			// grapheme clusters, because we assume that the prefix should
			// already have been normalized during the construction of these
			// refinements.
			commonPrefix := aRfn.prefix[:commonLen]
			for i := range commonLen {
				if bRfn.prefix[i] != commonPrefix[i] {
					commonPrefix = aRfn.prefix[:i]
					break
				}
			}
			if len(commonPrefix) != 0 {
				rfn = &refinementString{
					prefix:             commonPrefix,
					refinementNullable: nullable,
				}
				break
			}
		}
	case *refinementCollection:
		bRfn, ok := b.raw.(*refinementCollection)
		if !ok {
			break
		}
		rfn = &refinementCollection{
			minLen:             min(aRfn.minLen, bRfn.minLen),
			maxLen:             max(aRfn.maxLen, bRfn.maxLen),
			refinementNullable: nullable,
		}
	case *refinementNumber:
		bRfn, ok := b.raw.(*refinementNumber)
		if !ok {
			break
		}
		newMin := mostNumberValue(Value.LessThan, aRfn.min, bRfn.min)
		newMax := mostNumberValue(Value.GreaterThan, aRfn.max, bRfn.max)
		// For simplicity we currently always return inclusive bounds here,
		// which means our result is potentially slightly looser than it
		// could've been but we're still definitely returning something that
		// subsumes both ranges.
		rfn = &refinementNumber{
			min: newMin, minInc: true,
			max: newMax, maxInc: true,
			refinementNullable: nullable,
		}
	}
	if rfn == nil && nullable.isNull != tristateUnknown {
		// We do a little hoop-jumping here because RefinementBuilder knows
		// the various rules about how refinements are supposed to be
		// represented for different value types and we want to follow that
		// without duplicating those rules here.
		builder := UnknownVal(a.ty).Refine()
		switch nullable.isNull {
		case tristateTrue:
			builder = builder.Null()
		case tristateFalse:
			builder = builder.NotNull()
		}
		rfn = builder.currentRefinement()
	}
	return ValueRange{
		ty:  a.ty,
		raw: rfn,
	}
}
