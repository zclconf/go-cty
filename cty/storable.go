package cty

import (
	"errors"
)

// Storable is a wrapper for marshalling and unmarshalling cty Values,
// currently supporting only JSON encoding.
//
// Values themselves cannot be directly JSON-serialized because the cty
// type system is not directly compatible with JSON's type system, and so
// type information (at the cty layer) must be provided in order to recover
// the original Value after unmarshalling.
//
// Storable can be used with encoding/json to wrap a Value to be stored
// without its associated type information, and can later recover an
// equal Value by passing an identical type to the Value method.
type Storable struct {
	value interface{}
}

// Storable wraps the receiver in a Storable object ready to be serialized.
//
// The resulting object can be passed to Marshal in encoding/json. After later
// being unmarshalled, the original value can be recovered by passing the
// same type to the Value method of the Storable object.
func (val Value) Storable() Storable {
	// TODO: Implement
	return Storable{nil}
}

// Value unwraps a Storable using the given type information.
//
// If the given type is identical to that of the type that was stored then
// the result is guaranteed to be equal to what was stored. If the given type
// differs then recovery will be attempted, trying to convert values to
// conform to the given type, which may fail depending on how divergent the
// stored values are from the target type. If recovery via conversion is
// not possible then an error is returned and the value is not meaningful.
func (s Storable) Value(t Type) (Value, error) {
	// TODO: Implement
	return DynamicVal, nil
}

// MarshalJSON implements interface Marshaler from encoding/json, allowing
// Storables to be encoded as JSON.
func (s Storable) MarshalJSON() ([]byte, error) {
	// TODO: Implement
	return nil, errors.New("not implemented")
}

// UnmarshalJSON implements interface Unmarshaler from encoding/json, allowing
// Storables to be decoded from JSON.
func (s *Storable) UnmarshalJSON(d []byte) error {
	// TODO: Implement
	return errors.New("not implemented")
}
