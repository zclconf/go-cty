package cty

// CollectionType is a specialization of Type for types that are a container
// for multiple elements of another single type. Certain Value operations
// are valid for all collection types.
type CollectionType interface {
	Type
	ElementType() Type
}

// IsCollectionType returns true if the given type supports the operations
// that are defined for all collection types.
func IsCollectionType(t Type) bool {
	_, ok := t.(CollectionType)
	return ok
}
