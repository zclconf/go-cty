package cty

import (
	"fmt"
)

type typeObject struct {
	typeImplSigil
	attrTypes map[string]Type
}

// Object creates a map type with the given attribute types.
//
// After a map is passed to this function the caller must no longer access it,
// since ownership is transferred to this library.
func Object(attrTypes map[string]Type) Type {
	return Type{
		typeObject{
			attrTypes: attrTypes,
		},
	}
}

func (t typeObject) Equals(other Type) bool {
	if ot, ok := other.typeImpl.(typeObject); ok {
		if len(t.attrTypes) != len(ot.attrTypes) {
			// Fast path: if we don't have the same number of attributes
			// then we can't possibly be equal. This also avoids the need
			// to test attributes in both directions below, since we know
			// there can't be extras in "other".
			return false
		}

		for attr, ty := range t.attrTypes {
			oty, ok := ot.attrTypes[attr]
			if !ok {
				return false
			}
			if !oty.Equals(ty) {
				return false
			}
		}

		return true
	}
	panic("not an object type")
}

func (t typeObject) FriendlyName() string {
	// There isn't really a friendly way to write an object type due to its
	// complexity, so we'll just do something English-ish. Callers will
	// probably want to make some extra effort to avoid ever printing out
	// an object type FriendlyName in its entirety. For example, could
	// produce an error message by diffing two object types and saying
	// something like "Expected attribute foo to be string, but got number".
	// TODO: Finish this
	return "object"
}

func (t typeObject) GoString() string {
	if len(t.attrTypes) == 0 {
		return "cty.EmptyObject"
	}
	return fmt.Sprintf("cty.Object(%#v)", t.attrTypes)
}

// EmptyObject is a shorthand for Object(map[string]Type{}), to more
// easily talk about the empty object type.
var EmptyObject Type

// EmptyObjectVal is the only possible non-null, non-unknown value of type
// EmptyObject.
var EmptyObjectVal Value

func init() {
	EmptyObject = Object(map[string]Type{})
	EmptyObjectVal = Value{
		ty: EmptyObject,
		v:  map[string]interface{}{},
	}
}
