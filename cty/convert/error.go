package convert

import (
	"fmt"

	"github.com/apparentlymart/go-cty/cty"
)

// Error is a custom error type returned by most conversion operations that
// describes the location with a possibly-deep data structure that an error
// occured during conversion.
//
// Error embeds the standard error type and adds an additional Path attribute
// describing the location where the error occured.
type Error struct {
	error
	Path cty.Path
}

func errorf(path cty.Path, f string, args ...interface{}) error {
	return Error{
		error: fmt.Errorf(f, args...),
		Path:  path,
	}
}
