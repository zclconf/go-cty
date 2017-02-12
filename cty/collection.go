package cty

type collectionTypeImpl interface {
	ElementType() Type
}

// IsCollectionType returns true if the given type supports the operations
// that are defined for all collection types.
func (t Type) IsCollectionType() bool {
	_, ok := t.typeImpl.(collectionTypeImpl)
	return ok
}
