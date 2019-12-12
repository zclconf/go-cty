package cty

import (
	"reflect"
)

// CapsuleOps represents a set of overloaded operations for a capsule type.
//
// Each field is a reference to a function that can either be nil or can be
// set to an implementation of the corresponding operation. If an operation
// function is nil then it isn't supported for the given capsule type.
type CapsuleOps struct {
	// GoString provides the GoString implementation for values of the
	// corresponding type. Conventionally this should return a string
	// representation of an expression that would produce an equivalent
	// value.
	GoString func(val interface{}) string

	// TypeGoString provides the GoString implementation for the corresponding
	// capsule type itself.
	TypeGoString func(goTy reflect.Type) string

	// Equals provides the implementation of the Equals operation. This is
	// called only with known, non-null values of the corresponding type,
	// but if the corresponding type is a compound type then it must be
	// ready to detect and handle nested unknown or null values, usually
	// by recursively calling Value.Equals on those nested values.
	//
	// The result value must always be of type cty.Bool, or the Equals
	// operation will panic.
	//
	// If RawEquals is set without also setting Equals, the RawEquals
	// implementation will be used as a fallback implementation. That fallback
	// is appropriate only for leaf types that do not contain any nested
	// cty.Value that would need to distinguish Equals vs. RawEquals for their
	// own equality.
	//
	// If RawEquals is nil then Equals must also be nil, selecting the default
	// pointer-identity comparison instead.
	Equals func(a, b interface{}) Value

	// RawEquals provides the implementation of the RawEquals operation.
	// This is called only with known, non-null values of the corresponding
	// type, but if the corresponding type is a compound type then it must be
	// ready to detect and handle nested unknown or null values, usually
	// by recursively calling Value.RawEquals on those nested values.
	//
	// If RawEquals is nil, values of the corresponding type are compared by
	// pointer identity of the encapsulated value.
	RawEquals func(a, b interface{}) bool

	// ConversionFrom can provide conversions from the corresponding type to
	// some other type when values of the corresponding type are used with
	// the "convert" package. (The main cty package does not use this operation.)
	//
	// This function itself returns a function, allowing it to switch its
	// behavior depending on the given source type. Return nil to indicate
	// that no such conversion is available.
	ConversionFrom func(src Type) func(interface{}, Path) (Value, error)

	// ConversionTo can provide conversions to the corresponding type from
	// some other type when values of the corresponding type are used with
	// the "convert" package. (The main cty package does not use this operation.)
	//
	// This function itself returns a function, allowing it to switch its
	// behavior depending on the given destination type. Return nil to indicate
	// that no such conversion is available.
	ConversionTo func(dst Type) func(Value, Path) (interface{}, error)
}

// noCapsuleOps is a pointer to a CapsuleOps with no functions set, which
// is used as the default operations value when a type is created using
// the Capsule function.
var noCapsuleOps = &CapsuleOps{}

func (ops *CapsuleOps) assertValid() {
	if ops.RawEquals == nil && ops.Equals != nil {
		panic("Equals cannot be set without RawEquals")
	}
}

// CapsuleOps returns a pointer to the CapsuleOps value for a capsule type,
// or panics if the receiver is not a capsule type.
//
// The caller must not modify the CapsuleOps.
func (ty Type) CapsuleOps() *CapsuleOps {
	if !ty.IsCapsuleType() {
		panic("not a capsule-typed value")
	}

	return ty.typeImpl.(*capsuleType).Ops
}
