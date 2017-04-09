package cty

import (
	"fmt"
)

type typeSet struct {
	typeImplSigil
	elementType Type
}

// Set creates a set type with the given element Type.
//
// Set types are CollectionType implementations.
func Set(elem Type) Type {
	return Type{
		typeSet{
			elementType: elem,
		},
	}
}

// Equals returns true if the other Type is a map whose element type is
// equal to that of the receiver.
func (t typeSet) Equals(other Type) bool {
	ot, isSet := other.typeImpl.(typeSet)
	if !isSet {
		return false
	}

	return t.elementType.Equals(ot.elementType)
}

func (t typeSet) FriendlyName() string {
	return "set of " + t.elementType.FriendlyName()
}

func (t typeSet) ElementType() Type {
	return t.elementType
}

func (t typeSet) GoString() string {
	return fmt.Sprintf("cty.Map(%#v)", t.elementType)
}

// IsSetType returns true if the given type is a list type, regardless of its
// element type.
func (t Type) IsSetType() bool {
	_, ok := t.typeImpl.(typeSet)
	return ok
}

// SetElementType is a convenience method that checks if the given type is
// a set type, returning a pointer to its element type if so and nil
// otherwise. This is intended to allow convenient conditional branches,
// like so:
//
//     if et := t.SetElementType(); et != nil {
//         // Do something with *et
//     }
func (t Type) SetElementType() *Type {
	if lt, ok := t.typeImpl.(typeSet); ok {
		return &lt.elementType
	}
	return nil
}
