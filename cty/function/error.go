package function

import "fmt"

// ArgError represents an error with one of the arguments in a call. The
// attribute Index represents the zero-based index of the argument in question.
//
// Its error *may* be a cty.PathError, in which case the error actually
// pertains to a nested value within the data structure passed as the argument.
type ArgError struct {
	error
	Index int
}

func argErrorf(i int, f string, args ...interface{}) error {
	return ArgError{
		error: fmt.Errorf(f, args...),
		Index: i,
	}
}

func argError(i int, err error) error {
	return ArgError{
		error: err,
		Index: i,
	}
}
