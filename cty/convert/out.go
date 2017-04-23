package convert

import (
	"reflect"

	"github.com/apparentlymart/go-cty/cty"
)

// FromCtyValue assigns a cty.Value to a reflect.Value, which must be a pointer,
// using a fixed set of conversion rules.
//
// This function considers its audience to be the creator of the cty Value
// given, and thus the error messages it generates are (unlike with ToCtyValue)
// presented in cty terminology that is generally appropriate to return to
// end-users in applications where cty data structures are built from
// user-provided configuration.
//
// If an error is returned, the target data structure may have been partially
// populated, but the degree to which this is true is an implementation
// detail that the calling application should not rely on.
//
// The function will panic if given a non-pointer as the Go value target,
// since that is considered to be a bug in the calling program.
func FromCtyValue(val cty.Value, target interface{}) error {
	return nil
}
