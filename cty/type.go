package cty

// Type represents value types within the type system.
//
// This is a closed interface type, meaning that only the concrete
// implementations provided within this package are considered valid.
type Type interface {
	// typeSigil is a do-nothing method that exists only to express
	// that a type is an implementation of Type.
	typeSigil() typeImpl

	// Equals returns true if the other given Type exactly equals the
	// receiver Type.
	Equals(other Type) bool

	// FriendlyName returns a human-friendly *English* name for the given
	// type.
	FriendlyName() string
}

// Base implementation of Type to embed into concrete implementations
// to signal that they are implementations of Type.
type typeImpl struct{}

func (t typeImpl) typeSigil() typeImpl {
	return typeImpl{}
}
