package cty

import (
	"fmt"
	"iter"
	"maps"
)

// unknownType is the placeholder type used for the sigil value representing
// "Unknown", to make it unambigiously distinct from any other possible value.
type unknownType struct {
	// refinement is an optional object which, if present, describes some
	// additional constraints we know about the range of real values this
	// unknown value could be a placeholder for.
	refinement unknownValRefinement

	// nestedMarks optionally represents marks that are expected to appear
	// in at least one location inside a final complex-typed value, as opposed
	// to on the complex-typed value itself. For example, an unknown list might
	// track that a particular mark is expected to appear within at least one
	// of the final list's elements without conservatively attributing that mark
	// to the list itself, and so e.g. taking the length of the list would not
	// produce a marked value, but indexing into it would.
	//
	// When cty operations populate this automatically it usually only includes
	// the subset of marks that are considered to be "provenance" marks, because
	// some older callers use marks in a way where nested tracking would
	// cause overly-conservative behavior. Values constructed directly by a
	// caller are allowed to have non-nested marks in here too, though.
	nestedMarks ValueMarks
}

// totallyUnknown is the representation a a value we know nothing about at
// all. Subsequent refinements of an unknown value will cause creation of
// other values of unknownType that can represent additional constraints
// on the unknown value, but all unknown values start as totally unknown
// and we will also typically lose all unknown value refinements when
// round-tripping through serialization formats.
var totallyUnknown any = &unknownType{}

// UnknownVal returns an Value that represents an unknown value of the given
// type. Unknown values can be used to represent a value that is
// not yet known. Its meaning is undefined in cty, but it could be used by
// an calling application to allow partial evaluation.
//
// Unknown values of any type can be created of any type. All operations on
// Unknown values themselves return Unknown.
func UnknownVal(t Type) Value {
	if t == DynamicPseudoType {
		// We do our best to have only one representation of this situation,
		// although this is not guaranteed when unknown values are constructed
		// in different ways.
		return DynamicVal
	}
	return Value{
		ty: t,
		v:  totallyUnknown,
	}
}

// UnknownValWithNestedMarks is a variant of [UnknownVal] which also tracks
// that all of the values in the given sequence are marks that are assumed to
// apply to at least one nested value inside the final known value.
//
// Only complex-typed values can have nested marks, so if this is called with
// a primitive type the nested marks are silently ignored and this is therefore
// equivalent to [UnknownVal]. The "marks" sequence is allowed to be nil, in
// which case this is exactly equivalent to calling [UnknownVal] with the same
// type.
//
// When cty operations automatically track nested marks for unknown values they
// are typically filtered to include only "provenance marks" as described in
// [ctymarks.ProvenenanceMark], but when explicitly constructing a value with
// this function it's valid to specify non-provenance marks too.
func UnknownValWithNestedMarks(t Type, marks iter.Seq[any]) Value {
	if marks == nil || t.IsPrimitiveType() {
		return UnknownVal(t)
	}
	var nestedMarks ValueMarks
	for mark := range marks {
		if nestedMarks == nil {
			nestedMarks = make(ValueMarks)
		}
		nestedMarks[mark] = struct{}{}
	}
	if t.IsSetType() {
		// When marked values are included on individual elements passed to
		// SetVal they actually all just aggregate on the set itself because
		// set elements cannot retain their own individual marks. We'll mimic
		// that here for the equivalent unknown value behavior, putting
		// the marks directly on the unknown value instead of nested within.
		return UnknownVal(t).WithMarks(nestedMarks)
	}
	return Value{
		ty: t,
		v: &unknownType{
			nestedMarks: nestedMarks,
		},
	}
}

func (t unknownType) GoString() string {
	// This is the stringification of our internal unknown marker. The
	// stringification of the public representation of unknowns is in
	// Value.GoString.
	return fmt.Sprintf("cty.unknownType{refinement:%#v}", t.refinement)
}

type pseudoTypeDynamic struct {
	typeImplSigil
}

// DynamicPseudoType represents the dynamic pseudo-type.
//
// This type can represent situations where a type is not yet known. Its
// meaning is undefined in cty, but it could be used by a calling
// application to allow expression type checking with some types not yet known.
// For example, the application might optimistically permit any operation on
// values of this type in type checking, allowing a partial type-check result,
// and then repeat the check when more information is known to get the
// final, concrete type.
//
// It is a pseudo-type because it is used only as a sigil to the calling
// application. "Unknown" is the only valid value of this pseudo-type, so
// operations on values of this type will always short-circuit as per
// the rules for that special value.
var DynamicPseudoType Type

func (t pseudoTypeDynamic) Equals(other Type) bool {
	_, ok := other.typeImpl.(pseudoTypeDynamic)
	return ok
}

func (t pseudoTypeDynamic) FriendlyName(mode friendlyTypeNameMode) string {
	switch mode {
	case friendlyTypeConstraintName:
		return "any type"
	default:
		return "dynamic"
	}
}

func (t pseudoTypeDynamic) GoString() string {
	return "cty.DynamicPseudoType"
}

// DynamicVal is the only valid value of the pseudo-type dynamic.
// This value can be used as a placeholder where a value or expression's
// type and value are both unknown, thus allowing partial evaluation. See
// the docs for DynamicPseudoType for more information.
//
// Some applications historically treated DynamicVal as trivially comparable
// with itself using the Go "==" operator, and so that is now guaranteed to
// be valid with one important caveat: applications that either use
// [UnknownValWithNestedMarks] with [DynamicPseudoType] or that perform
// operations on values that contain provenance marks ([ctymarks.ProvenanceMark])
// will encounter unknown values of [DynamicPseudoType] which DO NOT trivially
// compare equal with [DynamicVal], and so should perform a more precise test
// for whatever they were intending to achieve in such a comparison. For
// example, consider the following to recognize "unknown value of a
// completely-unknown type":
//
//	val.Type() == cty.DynamicPseudoType && !val.IsKnown()
var DynamicVal Value

func init() {
	DynamicPseudoType = Type{
		pseudoTypeDynamic{},
	}
	DynamicVal = Value{
		ty: DynamicPseudoType,
		v:  totallyUnknown,
	}
}

// unknownValueNestedMarks is an internal helper for obtaining the set of
// nested marks for an unknown value.
//
// The result is a direct reference to the value's marks set, and so the caller
// must treat it as immutable to avoid breaking the expectations of other
// holders of the value.
func (v Value) unknownValueNestedMarks() ValueMarks {
	if v.IsMarked() {
		v, marks := v.Unmark()
		nestedMarks := v.unknownValueNestedMarks()
		// Note that we intentionally modify marks rather than nestedMarks
		// hre because Unmark produces a separate map for each call, whereas
		// unknownValueNestedMarks returns a shared reference to the same
		// marks set inside v which we must therefore treat as immutable.
		maps.Copy(marks, nestedMarks)
		return marks
	}
	unk, ok := v.v.(*unknownType)
	if !ok {
		panic("unknownValueNestedMarks is only valid for unknown values")
	}
	return unk.nestedMarks
}

func (v Value) unknownWithNewNestedMarks(newMarks ValueMarks) Value {
	if v.IsMarked() {
		v, vMarks := v.Unmark()
		return v.unknownWithNewNestedMarks(newMarks).WithMarks(vMarks)
	}
	if v.Type().IsPrimitiveType() {
		// Primitive-typed unknown values are not permitted to have nested
		// marks, but we'll just silently ignore the request instead of
		// failing because this is only an internal helper function we use
		// to reduce the boilerplate of transferring nested marks between
		// values.
		return v
	}

	unk, ok := v.v.(*unknownType)
	if !ok {
		panic("unknownWithNewNestedMarks is only valid for unknown values")
	}
	return Value{
		ty: v.ty,
		v: &unknownType{
			refinement:  unk.refinement,
			nestedMarks: newMarks,
		},
	}
}
