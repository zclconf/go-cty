package cty

import (
	"fmt"
	"math/big"
	"strings"
)

// Refine creates a [RefinementBuilder] with which to annotate the reciever
// with zero or more additional refinements that constrain the range of
// the value.
//
// Calling methods on a RefinementBuilder for a known value essentially just
// serves as assertions about the range of that value, leading to panics if
// those assertions don't hold in practice. This is mainly supported just to
// make programs that rely on refinements automatically self-check by using
// the refinement codepath unconditionally on both placeholders and final
// values for those placeholders. It's always a bug to refine the range of
// an unknown value and then later substitute an exact value outside of the
// refined range.
//
// Calling methods on a RefinementBuilder for an unknown value is perhaps
// more useful because the newly-refined value will then be a placeholder for
// a smaller range of values and so it may be possible for other operations
// on the unknown value to return a known result despite the exact value not
// yet being known.
//
// It is never valid to refine [DynamicVal], because that value is a
// placeholder for a value about which we knkow absolutely nothing. A value
// must at least have a known root type before it can support further
// refinement.
func (v Value) Refine() *RefinementBuilder {
	if unk, isUnk := v.v.(*unknownType); isUnk && unk.refinement != nil {
		// We're refining a value that's already been refined before, so
		// we'll start from a copy of its existing refinements.
		wip := unk.refinement.copy()
		return &RefinementBuilder{v, wip}
	}

	ty := v.Type()
	var wip unknownValRefinement
	switch {
	case ty == String:
		wip = &refinementString{}
	case ty == Number:
		wip = &refinementNumber{}
	case ty.IsCollectionType():
		wip = &refinementCollection{
			// A collection can never have a negative length, so we'll
			// start with that already constrained.
			minLen: Zero,
			minInc: true,
		}
	case ty == Bool || ty.IsObjectType() || ty.IsTupleType() || ty.IsCapsuleType():
		// For other known types we'll just track nullability
		wip = &refinementNullable{}
	default:
		// we leave "wip" as nil for all other types, representing that
		// they don't support refinements at all and so any call on the
		// RefinementBuilder should fail.

		// NOTE: We intentionally don't allow any refinements for
		// cty.DynamicPseudoType here, even though it could be nice in principle
		// to at least track non-nullness for those, because it's historically
		// been valid to directly compare values with cty.DynamicVal using
		// the Go "==" operator and recording a refinement for an untyped
		// unknown value would break existing code relying on that.
	}

	return &RefinementBuilder{v, wip}
}

// RefineNotNull is a shorthand for Value.Refine().NotNull().NewValue(), because
// declaring that a unknown value isn't null is by far the most common use of
// refinements.
func (v Value) RefineNotNull() Value {
	return v.Refine().NotNull().NewValue()
}

// RefinementBuilder is a supporting type for the [Value.Refine] method,
// using the builder pattern to apply zero or more constraints before
// constructing a new value with all of those constraints applied.
//
// Most of the methods of this type return the same reciever to allow
// for method call chaining. End call chains with a call to
// [RefinementBuilder.NewValue] to obtain the newly-refined value.
type RefinementBuilder struct {
	orig Value
	wip  unknownValRefinement
}

func (b *RefinementBuilder) assertRefineable() {
	if b.wip == nil {
		panic(fmt.Sprintf("cannot refine a %#v value", b.orig.Type()))
	}
}

// NotNull constrains the value as definitely not being null.
//
// NotNull is valid when refining values of the following types:
//   - number, boolean, and string values
//   - list, set, or map types of any element type
//   - values of object types
//   - values of collection types
//   - values of capsule types
//
// When refining any other type this function will panic.
//
// In particular note that it is not valid to constrain an untyped value
// -- a value whose type is `cty.DynamicPseudoType` -- as being non-null.
// An unknown value of an unknown type is always completely unconstrained.
func (b *RefinementBuilder) NotNull() *RefinementBuilder {
	b.assertRefineable()

	if b.orig.IsKnown() && b.orig.IsNull() {
		panic("refining null value as non-null")
	}

	b.wip.setNull(tristateFalse)

	return b
}

// Null constrains the value as definitely null.
//
// Null is valid for the same types as [RefinementBuilder.NotNull].
// When refining any other type this function will panic.
//
// Explicitly cnstraining a value to be null is strange because that suggests
// that the caller does actually know the value -- there is only one null
// value for each type constraint -- but this is here for symmetry with the
// fact that a [ValueRange] can also represent that a value is definitely null.
func (b *RefinementBuilder) Null() *RefinementBuilder {
	b.assertRefineable()

	if b.orig.IsKnown() && !b.orig.IsNull() {
		panic("refining non-null value as null")
	}

	b.wip.setNull(tristateTrue)

	return b
}

// NumericRange constrains the upper and/or lower bounds of a number value,
// or panics if this builder is not refining a number value.
//
// The two given values are interpreted as inclusive bounds and either one
// may be an unknown number if only one of the two bounds is currently known.
// If either of the given values is not a non-null number value then this
// function will panic.
func (b *RefinementBuilder) NumberRangeInclusive(min, max Value) *RefinementBuilder {
	return b.numberRange(min, max, true, true)
}

// CollectionLengthLowerBound constrains the lower bound of the length of a
// collection value, or panics if this builder is not refining a collection
// value.
//
// The lower bound must be a known, non-null number or this function will
// panic.
func (b *RefinementBuilder) CollectionLengthLowerBound(min Value, inclusive bool) *RefinementBuilder {
	b.assertRefineable()

	wip, ok := b.wip.(*refinementCollection)
	if !ok {
		panic(fmt.Sprintf("cannot refine collection length bounds for a %#v value", b.orig.Type()))
	}

	if min.IsNull() {
		panic("collection length bound is null")
	}
	if !min.IsKnown() {
		panic("collection length bound is unknown")
	}

	if b.orig.IsKnown() {
		realLen := b.orig.Length()
		if gt := min.GreaterThan(realLen); gt.IsKnown() && gt.True() {
			panic(fmt.Sprintf("refining collection of length %#v with minimum bound %#v", realLen, min))
		}
	}

	if wip.minLen != NilVal {
		var ok bool
		if wip.minInc {
			ok = min.GreaterThanOrEqualTo(wip.minLen).True()
		} else {
			ok = min.GreaterThan(wip.minLen).True()
		}
		if !ok {
			panic("refined collection length lower bound is inconsistent with existing lower bound")
		}
	}

	wip.minLen = min
	wip.minInc = inclusive
	wip.assertConsistentLengthBounds()

	return b
}

// CollectionLengthUpperBound constrains the upper bound of the length of a
// collection value, or panics if this builder is not refining a collection
// value.
//
// The upper bound must be a known, non-null number or this function will
// panic.
func (b *RefinementBuilder) CollectionLengthUpperBound(max Value, inclusive bool) *RefinementBuilder {
	b.assertRefineable()

	wip, ok := b.wip.(*refinementCollection)
	if !ok {
		panic(fmt.Sprintf("cannot refine collection length bounds for a %#v value", b.orig.Type()))
	}

	if max.IsNull() {
		panic("collection length bound is null")
	}
	if !max.IsKnown() {
		panic("collection length bound is unknown")
	}

	if b.orig.IsKnown() {
		realLen := b.orig.Length()
		if gt := max.LessThan(realLen); gt.IsKnown() && gt.True() {
			panic(fmt.Sprintf("refining collection of length %#v with maximum bound %#v", realLen, max))
		}
	}

	if wip.maxLen != NilVal {
		var ok bool
		if wip.maxInc {
			ok = max.LessThanOrEqualTo(wip.minLen).True()
		} else {
			ok = max.LessThan(wip.minLen).True()
		}
		if !ok {
			panic("refined collection length upper bound is inconsistent with existing upper bound")
		}
	}

	wip.maxLen = max
	wip.maxInc = inclusive
	wip.assertConsistentLengthBounds()

	return b
}

// StringPrefix constrains the prefix of a string value, or panics if this
// builder is not refining a string value.
//
// The given prefix will be Unicode normalized in the same way that a
// cty.StringVal would be. However, since prefix is just a substring the
// normalization may produce a non-matching prefix string if the given prefix
// splits a sequence of combining characters. For correct results always ensure
// that the prefix ends at a grapheme cluster boundary.
func (b *RefinementBuilder) StringPrefix(prefix string) *RefinementBuilder {
	b.assertRefineable()

	wip, ok := b.wip.(*refinementString)
	if !ok {
		panic(fmt.Sprintf("cannot refine string prefix for a %#v value", b.orig.Type()))
	}

	// We must apply the same Unicode processing we'd normally use for a
	// cty string so that the prefix will be comparable.
	prefix = NormalizeString(prefix)

	// If we have a known string value then the given prefix must actually
	// match it.
	if b.orig.IsKnown() && !b.orig.IsNull() {
		have := b.orig.AsString()
		matchLen := len(have)
		if l := len(prefix); l < matchLen {
			matchLen = l
		}
		have = have[:matchLen]
		new := prefix[:matchLen]
		if have != new {
			panic("refined prefix is inconsistent with known value")
		}
	}

	// If we already have a refined prefix then the overlapping parts of that
	// and the new prefix must match.
	{
		matchLen := len(wip.prefix)
		if l := len(prefix); l < matchLen {
			matchLen = l
		}

		have := wip.prefix[:matchLen]
		new := prefix[:matchLen]
		if have != new {
			panic("refined prefix is inconsistent with previous refined prefix")
		}
	}

	// We'll only save the new prefix if it's longer than the one we already
	// had.
	if len(prefix) > len(wip.prefix) {
		wip.prefix = prefix
	}

	return b
}

func (b *RefinementBuilder) numberRange(min, max Value, minInc, maxInc bool) *RefinementBuilder {
	b.assertRefineable()

	wip, ok := b.wip.(*refinementNumber)
	if !ok {
		panic(fmt.Sprintf("cannot refine numeric range for a %#v value", b.orig.Type()))
	}
	// After this point b.orig is guaranteed to have type cty.Number

	if min.Type() != Number || max.Type() != Number {
		panic("refining numeric range with a non-numeric bound")
	}
	if min.IsNull() || max.IsNull() {
		panic("refining numeric range with a null bound")
	}

	uncomparable := func(v Value) bool {
		return v.IsNull() || !v.IsKnown()
	}
	checkMinRangeFunc := func(inclusive bool) func(Value, Value) bool {
		if inclusive {
			return func(a, b Value) bool {
				if uncomparable(a) || uncomparable(b) {
					return true // default to valid if we're not sure
				}
				return a.GreaterThanOrEqualTo(b).True()
			}
		} else {
			return func(a, b Value) bool {
				if uncomparable(a) || uncomparable(b) {
					return true // default to valid if we're not sure
				}
				return a.GreaterThan(b).True()
			}
		}
	}
	checkMaxRangeFunc := func(inclusive bool) func(Value, Value) bool {
		if inclusive {
			return func(a, b Value) bool {
				if uncomparable(a) || uncomparable(b) {
					return true // default to valid if we're not sure
				}
				return a.LessThanOrEqualTo(b).True()
			}
		} else {
			return func(a, b Value) bool {
				if uncomparable(a) || uncomparable(b) {
					return true // default to valid if we're not sure
				}
				return a.LessThan(b).True()
			}
		}
	}

	// If our original value is known then it must be in the given range.
	if v := b.orig; v.IsKnown() && !v.IsNull() {
		if !checkMinRangeFunc(minInc)(v, min) {
			panic(fmt.Sprintf("refining %#v with invalid lower bound %#v", v, min))
		}
		if !checkMaxRangeFunc(maxInc)(v, max) {
			panic(fmt.Sprintf("refining %#v with invalid upper bound %#v", v, min))
		}
	}

	// If we already have bounds then the new bounds must be consistent with them.
	if wip.min != NilVal && !checkMinRangeFunc(wip.minInc)(wip.min, min) {
		panic(fmt.Sprintf("new refined lower bound %#v conflicts with previous %#v", min, wip.min))
	}
	if wip.max != NilVal && !checkMaxRangeFunc(wip.maxInc)(wip.max, max) {
		panic(fmt.Sprintf("new refined upper bound %#v conflicts with previous %#v", min, wip.min))
	}

	// We only record known bounds. An unknown value for either bound means
	// it's either unbounded or we'll retain a prevously-recorded bound.
	if min.IsKnown() {
		wip.min = min
		wip.minInc = minInc
	}
	if max.IsKnown() {
		wip.max = max
		wip.maxInc = maxInc
	}
	wip.assertConsistentBounds()

	return b
}

// NewValue completes the refinement process by constructing a new value
// that is guaranteed to meet all of the previously-specified refinements.
//
// If the original value being refined was known then the result is exactly
// that value, because otherwise the previous refinement calls would have
// panicked reporting the refinements as invalid for the value.
//
// If the original value was unknown then the result is typically also unknown
// but may have additional refinements compared to the original. If the applied
// refinements have reduced the range to a single exact value then the result
// might be that known value.
func (b *RefinementBuilder) NewValue() Value {
	if b.orig.IsKnown() {
		return b.orig
	}

	// We have a few cases where the value has been refined enough that we now
	// know exactly what the value is, or at least we can produce a more
	// detailed approximation of it.
	switch b.wip.null() {
	case tristateTrue:
		// There is only one null value of each type so this is now known.
		return NullVal(b.orig.Type())
	case tristateFalse:
		// If we know it's definitely not null then we might have enough
		// information to construct a known, non-null value.
		if rfn, ok := b.wip.(*refinementNumber); ok {
			// If both bounds are inclusive and equal then our value can
			// only be the same number as the bounds.
			if rfn.maxInc && rfn.minInc {
				if rfn.min != NilVal && rfn.max != NilVal {
					eq := rfn.min.Equals(rfn.max)
					if eq.IsKnown() && eq.True() {
						return rfn.min
					}
				}
			}
		} else if rfn, ok := b.wip.(*refinementCollection); ok {
			// If both length bounds are inclusive and equal then we know our
			// length is the same number as the bounds.
			if rfn.maxInc && rfn.minInc {
				if rfn.minLen != NilVal && rfn.maxLen != NilVal {
					eq := rfn.minLen.Equals(rfn.maxLen)
					if eq.IsKnown() && eq.True() {
						knownLen := rfn.minLen
						ty := b.orig.Type()
						if knownLen == Zero {
							// If we know the length is zero then we can construct
							// a known value of any collection kind.
							switch {
							case ty.IsListType():
								return ListValEmpty(ty.ElementType())
							case ty.IsSetType():
								return SetValEmpty(ty.ElementType())
							case ty.IsMapType():
								return MapValEmpty(ty.ElementType())
							}
						} else if ty.IsListType() {
							// If we know the length of the list then we can
							// create a known list with unknown elements instead
							// of a wholly-unknown list.
							if knownLen, acc := knownLen.AsBigFloat().Int64(); acc == big.Exact {
								elems := make([]Value, knownLen)
								unk := UnknownVal(ty.ElementType())
								for i := range elems {
									elems[i] = unk
								}
								return ListVal(elems)
							}
						} else if ty.IsSetType() && knownLen == NumberIntVal(1) {
							// If we know we have a one-element set then we
							// know the one element can't possibly coalesce with
							// anything else and so we can create a known set with
							// an unknown element.
							return SetVal([]Value{UnknownVal(ty.ElementType())})
						}
					}
				}
			}
		}
	}

	return Value{
		ty: b.orig.ty,
		v:  &unknownType{refinement: b.wip},
	}
}

// unknownValRefinment is an interface pretending to be a sum type representing
// the different kinds of unknown value refinements we support for different
// types of value.
type unknownValRefinement interface {
	unknownValRefinementSigil()
	copy() unknownValRefinement
	null() tristateBool
	setNull(tristateBool)
	rawEqual(other unknownValRefinement) bool
	GoString() string
}

type refinementString struct {
	refinementNullable
	prefix string
}

func (r *refinementString) unknownValRefinementSigil() {}

func (r *refinementString) copy() unknownValRefinement {
	ret := *r
	// Everything in refinementString is immutable, so a shallow copy is sufficient.
	return &ret
}

func (r *refinementString) rawEqual(other unknownValRefinement) bool {
	{
		other, ok := other.(*refinementString)
		if !ok {
			return false
		}
		return (r.refinementNullable.rawEqual(&other.refinementNullable) &&
			r.prefix == other.prefix)
	}
}

func (r *refinementString) GoString() string {
	var b strings.Builder
	b.WriteString(r.refinementNullable.GoString())
	if r.prefix != "" {
		fmt.Fprintf(&b, ".StringPrefix(%q)", r.prefix)
	}
	return b.String()
}

type refinementNumber struct {
	refinementNullable
	min, max       Value
	minInc, maxInc bool
}

func (r *refinementNumber) unknownValRefinementSigil() {}

func (r *refinementNumber) copy() unknownValRefinement {
	ret := *r
	// Everything in refinementNumber is immutable, so a shallow copy is sufficient.
	return &ret
}

func (r *refinementNumber) rawEqual(other unknownValRefinement) bool {
	{
		other, ok := other.(*refinementNumber)
		if !ok {
			return false
		}
		return (r.refinementNullable.rawEqual(&other.refinementNullable) &&
			r.min.RawEquals(other.min) &&
			r.max.RawEquals(other.max) &&
			r.minInc == other.minInc &&
			r.maxInc == other.maxInc)
	}
}

func (r *refinementNumber) GoString() string {
	var b strings.Builder
	b.WriteString(r.refinementNullable.GoString())
	if r.min != NilVal {
		fmt.Fprintf(&b, ".NumberLowerBound(%#v, %t)", r.min, r.minInc)
	}
	if r.max != NilVal {
		fmt.Fprintf(&b, ".NumberUpperBound(%#v, %t)", r.max, r.maxInc)
	}
	return b.String()
}

func (r *refinementNumber) assertConsistentBounds() {
	if r.min != NilVal && r.max != NilVal && r.max.LessThan(r.min).True() {
		panic("number upper bound is less than lower bound")
	}
}

type refinementCollection struct {
	refinementNullable
	minLen, maxLen Value
	minInc, maxInc bool
}

func (r *refinementCollection) unknownValRefinementSigil() {}

func (r *refinementCollection) copy() unknownValRefinement {
	ret := *r
	// Everything in refinementCollection is immutable, so a shallow copy is sufficient.
	return &ret
}

func (r *refinementCollection) rawEqual(other unknownValRefinement) bool {
	{
		other, ok := other.(*refinementCollection)
		if !ok {
			return false
		}
		return (r.refinementNullable.rawEqual(&other.refinementNullable) &&
			r.minLen.RawEquals(other.minLen) &&
			r.maxLen.RawEquals(other.maxLen) &&
			r.minInc == other.minInc &&
			r.maxInc == other.maxInc)
	}
}

func (r *refinementCollection) GoString() string {
	var b strings.Builder
	b.WriteString(r.refinementNullable.GoString())
	if r.minLen != NilVal && r.minLen != Zero {
		// (a lower bound of zero is the default)
		fmt.Fprintf(&b, ".CollectionLengthLowerBound(%#v, %t)", r.minLen, r.minInc)
	}
	if r.maxLen != NilVal {
		fmt.Fprintf(&b, ".CollectionLengthUpperBound(%#v, %t)", r.maxLen, r.maxInc)
	}
	return b.String()
}

func (r *refinementCollection) assertConsistentLengthBounds() {
	if r.minLen != NilVal && r.maxLen != NilVal && r.maxLen.LessThan(r.minLen).True() {
		panic("collection length upper bound is less than lower bound")
	}
}

type refinementNullable struct {
	isNull tristateBool
}

func (r *refinementNullable) unknownValRefinementSigil() {}

func (r *refinementNullable) copy() unknownValRefinement {
	ret := *r
	// Everything in refinementJustNull is immutable, so a shallow copy is sufficient.
	return &ret
}

func (r *refinementNullable) null() tristateBool {
	return r.isNull
}

func (r *refinementNullable) setNull(v tristateBool) {
	r.isNull = v
}

func (r *refinementNullable) rawEqual(other unknownValRefinement) bool {
	{
		other, ok := other.(*refinementNullable)
		if !ok {
			return false
		}
		return r.isNull == other.isNull
	}
}

func (r *refinementNullable) GoString() string {
	switch r.isNull {
	case tristateFalse:
		return ".NotNull()"
	case tristateTrue:
		return ".Null()"
	default:
		return ""
	}
}

type tristateBool rune

const tristateTrue tristateBool = 'T'
const tristateFalse tristateBool = 'F'
const tristateUnknown tristateBool = 0
