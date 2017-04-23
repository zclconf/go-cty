package convert

import (
	"reflect"

	"github.com/apparentlymart/go-cty/cty"
)

// FromCtyValue assigns a cty.Value to a reflect.Value, which must be a pointer,
// using a fixed set of conversion rules.
func FromCtyValue(val cty.Value, target reflect.Value) error {
	return nil
}
