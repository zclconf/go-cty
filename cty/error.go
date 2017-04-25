package cty

import (
	"fmt"
)

// PathError is a specialization of error that represents where in a
// potentially-deep data structure an error occured, using a Path.
type PathError struct {
	error
	Path Path
}

func errorf(path Path, f string, args ...interface{}) error {
	// We need to copy the Path because often our caller builds it by
	// continually mutating the same underlying buffer.
	sPath := make(Path, len(path))
	copy(sPath, path)
	return PathError{
		error: fmt.Errorf(f, args...),
		Path:  sPath,
	}
}
