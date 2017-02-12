package cty

import (
	"fmt"
)

// TypeList instances represent specific list types. Each distinct ElementType
// creates a distinct, non-equal list type.
type typeList struct {
	typeImpl
	elementType Type
}

// List creates a map type with the given element Type.
//
// List types are CollectionType implementations.
func List(elem Type) Type {
	return typeList{
		elementType: elem,
	}
}

// Equals returns true if the other Type is a list whose element type is
// equal to that of the receiver.
func (t typeList) Equals(other Type) bool {
	ot, isList := other.(typeList)
	if !isList {
		return false
	}

	return t.elementType.Equals(ot.elementType)
}

func (t typeList) FriendlyName() string {
	return "list of " + t.elementType.FriendlyName()
}

func (t typeList) ElementType() Type {
	return t.elementType
}

func (t typeList) GoString() string {
	return fmt.Sprintf("cty.List(%#v)", t.elementType)
}

// IsListType returns true if the given type is a list type, regardless of its
// element type.
func IsListType(t Type) bool {
	_, ok := t.(typeList)
	return ok
}

// ListElementType is a convenience method that checks if the given type is
// a list type, returning its element type if so and nil otherwise. This is
// intended to allow convenient conditional branches, like so:
//
//     if et := ListElementType(t); et != nil {
//         // Do something with "et"
//     }
func ListElementType(t Type) Type {
	if lt, ok := t.(typeList); ok {
		return lt.elementType
	}
	return nil
}
