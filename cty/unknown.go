package cty

import (
	"fmt"
	"maps"
)

// unknownType is the placeholder type used for the sigil value representing
// "Unknown", to make it unambigiously distinct from any other possible value.
type unknownType struct {
	// refinement is an optional object which, if present, describes some
	// additional constraints we know about the range of real values this
	// unknown value could be a placeholder for.
	refinement unknownValRefinement

	// nestedMarks represents marks that elements or attributes of the unknown
	// value may have, separately from marks on the collection or structural
	// value that the unknown value is directly representing.
	//
	// This is a conservative approximation that doesn't attempt to represent
	// exactly where in the potential final data structure the marks will
	// appear. It's here primarily just so that functions like
	// [Value.UnmarkDeep] can return an approximation of what that function
	// return on the known final value, while still keeping those nested marks
	// out of the result of the shallow [Value.Unmark].
	//
	// This could potentially grow to support more precise tracking of locations
	// of marks in the nested data structure later if we learn of a good reason
	// to do that, but it's not yet clear whether that complexity is warranted.
	// Conceptually the marks given here apply to any descendent of the unknown
	// value, regardless of the downstream path.
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
	return Value{
		ty: t,
		v:  totallyUnknown,
	}
}

// UnknownValWithNestedMarks is a variant of the exported [UnknownVal] which
// also records some nested marks inside its result. "Nested marks" in this
// case is an approximation for there being marks on the elements or attributes
// of an unknown value of a collection or structural type.
//
// Primitive-typed values, capsule-typed values, and values of empty structural
// types can never have nested marks so this panics if called with such a type.
func UnknownValWithNestedMarks(t Type, nestedMarks ValueMarks) Value {
	if !typeCanHaveNestedMarks(t) {
		panic(fmt.Sprintf("type %#v cannot have nested marks", t))
	}
	if len(nestedMarks) == 0 {
		return UnknownVal(t)
	}
	return Value{
		ty: t,
		v: &unknownType{
			nestedMarks: nestedMarks,
		},
	}
}

// typeCanHaveNestedMarks returns true if the given type is one whose known
// values could potentially have marks inside them, and therefore it's also
// valid for unknown values of that type to track nested marks.
func typeCanHaveNestedMarks(ty Type) bool {
	if ty.IsObjectType() {
		return len(ty.AttributeTypes()) != 0
	}
	if ty.IsTupleType() {
		return len(ty.TupleElementTypes()) != 0
	}
	return ty == DynamicPseudoType || ty.IsCollectionType()
}

// unknownMarkedDescendentPlaceholder is an internal helper that provides a
// placeholder value that represents an arbitrary descendent of the given
// unknown value if and only if the given value is an unknown value with
// nested marks. The result is always an unknown value of unknown type.
//
// This function returns [NilVal] in situations where there is no placeholder
// because there are not any nested marks.
func unknownMarkedDescendentPlaceholder(v Value) Value {
	unk, ok := v.v.(*unknownType)
	if !ok {
		return NilVal // not an unknown value
	}
	if len(unk.nestedMarks) == 0 {
		return NilVal // no marks, so not interesting to report
	}
	return DynamicVal.WithMarks(unk.nestedMarks).WithSameMarks(v)
}

func unknownValWithMoreNestedMarks(v Value, nestedMarks ValueMarks) Value {
	v, marks := v.Unmark()
	if v.IsKnown() {
		panic("unknownValWithMoreNestedMarks on known value")
	}
	if len(nestedMarks) == 0 {
		return v
	}
	unk := *v.v.(*unknownType) // shallow copy, since we're only going to modify nestedMarks
	if len(unk.nestedMarks) != 0 {
		nestedMarks = maps.Clone(nestedMarks)
		maps.Copy(nestedMarks, unk.nestedMarks)
	}
	unk.nestedMarks = nestedMarks
	v.v = &unk
	return v.WithMarks(marks)
}

func (t unknownType) GoString() string {
	// This is the stringification of our internal unknown marker. The
	// stringification of the public representation of unknowns is in
	// Value.GoString.
	return "cty.unknown"
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
