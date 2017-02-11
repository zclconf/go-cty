package cty

// Value represents a value of a particular type, and is the interface by
// which operations are executed on typed values.
type Value struct {
	ty    Type
	value interface{}
}

// Type returns the type of the value.
func (v Value) Type() Type {
	return v.ty
}

// IsKnown returns true if the value is known. That is, if it is not
// the result of the unknown value constructor Unknown(...), and is not
// the result of an operation on another unknown value.
//
// Unknown values are only produced either directly or as a result of
// operating on other unknown values, and so an application that never
// introduces Unknown values can be guaranteed to never receive any either.
func (v Value) IsKnown() bool {
	return v.value != unknown
}

// IsNull returns true if the value is null. Values of any type can be
// null, but any operations on a null value will panic. No operation ever
// produces null, so an application that never introduces Null values can
// be guaranteed to never receive any either.
func (v Value) IsNull() bool {
	return v.value == nil
}
