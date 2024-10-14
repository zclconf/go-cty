package cty

import (
	"fmt"
)

type typeUnion struct {
	typeImplSigil
	Variants map[string]Type
}

// EmptyUnion is a union without any variants.
//
// There are no known, non-null values of this type because a valid value
// of a union type requires exactly one variant to be selected.
var EmptyUnion = Union(nil)

// Union creates a union type with the given variants.
//
// Variants may be of any _exact_ type, but are forbidden from including
// cty.DynamicPseudoType. If any variant includes such an unspecified type,
// this function will panic.
//
// After a map is passed to this function the caller must no longer access it,
// since ownership is transferred to this library.
func Union(variants map[string]Type) Type {
	variantsNorm := make(map[string]Type, len(variants))
	for k, v := range variants {
		if v.HasDynamicTypes() {
			panic("union variant must not include cty.DynamicPseudoType")
		}
		variantsNorm[NormalizeString(k)] = v
	}

	return Type{
		typeUnion{
			Variants: variants,
		},
	}
}

func (t typeUnion) Equals(other Type) bool {
	if ot, ok := other.typeImpl.(typeUnion); ok {
		if len(t.Variants) != len(ot.Variants) {
			// Fast path: if we don't have the same number of variants
			// then we can't possibly be equal. This also avoids the need
			// to test variants in both directions below, since we know
			// there can't be extras in "other".
			return false
		}

		for attr, ty := range t.Variants {
			oty, ok := ot.Variants[attr]
			if !ok {
				return false
			}
			if !oty.Equals(ty) {
				return false
			}
		}

		return true
	}
	return false
}

func (t typeUnion) FriendlyName(mode friendlyTypeNameMode) string {
	// There isn't really a friendly way to write a union type due to its
	// complexity, so we'll just do something English-ish. Callers will
	// probably want to make some extra effort to avoid ever printing out
	// a union type FriendlyName in its entirety. For example, could
	// produce an error message by diffing two union types and saying
	// something like "Expected variant foo to be string, but got number".
	return "union"
}

func (t typeUnion) GoString() string {
	if len(t.Variants) == 0 {
		return "cty.EmptyUnion"
	}
	return fmt.Sprintf("cty.Union(%#v)", t.Variants)
}

// unionVal is the internal representation of a union value, capturing both
// the selected variant name and the raw representation of the value of that
// variant's type.
type unionVal struct {
	variant string
	value   any
}

// IsUnionType returns true if the given type is a union type, regardless
// of its variants.
func (t Type) IsUnionType() bool {
	_, ok := t.typeImpl.(typeUnion)
	return ok
}

// HasUnionVariant returns true if the receiver has a union variant with the
// given name, regardless of its type. Will panic if the reciever isn't a
// union type; use [Type.IsUnionType] to determine whether this operation will
// succeed.
func (t Type) HasUnionVariant(name string) bool {
	name = NormalizeString(name)
	if ot, ok := t.typeImpl.(typeUnion); ok {
		_, hasVariant := ot.Variants[name]
		return hasVariant
	}
	panic("HasUnionVariant on non-union Type")
}

// UnionVariantType returns the type of the variant with the given name. Will
// panic if the receiver is not a union type (use IsUnionType to confirm)
// or if the union type has no such variant (use HasUnionVariant to confirm).
func (t Type) UnionVariantType(name string) Type {
	name = NormalizeString(name)
	if ot, ok := t.typeImpl.(typeUnion); ok {
		aty, hasAttr := ot.Variants[name]
		if !hasAttr {
			panic("no such variant")
		}
		return aty
	}
	panic("UnionVariantType on non-union Type")
}

// UnionVariants returns a map from variant names to their associated
// types. Will panic if the receiver is not a union type (use IsUnionType
// to confirm).
//
// The returned map is part of the internal state of the type, and is provided
// for read access only. It is forbidden for any caller to modify the returned
// map. For many purposes the variant-related methods of Value are more
// appropriate and more convenient to use.
func (t Type) UnionVariants() map[string]Type {
	if ot, ok := t.typeImpl.(typeUnion); ok {
		return ot.Variants
	}
	panic("UnionVariants on non-union Type")
}

// unionVariantVal is an internal helper that decomposes a union-typed
// value into the name of the selected variant and the value assigned
// to it.
//
// This panics if the given value is not known, non-null, unmarked, and
// of a union type.
func unionVariantVal(v Value) (string, Value) {
	ty := v.Type()
	if !ty.IsUnionType() {
		panic("not a union type")
	}
	internal := v.v.(unionVal)
	vty := ty.UnionVariantType(internal.variant)
	return internal.variant, Value{
		ty: vty,
		v:  internal.value,
	}
}
