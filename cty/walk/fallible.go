package walk

// Fallible is an adapter to allow returning a (result, error) pair through
// the walk functions when needed, without forcing that on all uses of
// the walk functions.
type Fallible[T any] struct {
	v fallibleOutcome
}

// Result expands the Fallible into a typical (result, error) pair, for
// use with idiomatic Go error handling patterns.
func (f Fallible[T]) Result() (T, error) {
	switch v := f.v.(type) {
	case fallibleSuccess[T]:
		return v[0], nil
	case fallibleFailure:
		var zero T
		return zero, v[0]
	default:
		// Should not be possible to get here
		panic("unhandled fallible outcome")
	}
}

// IsSuccess returns true if the Fallible is representing a successful result.
func (f Fallible[T]) IsSuccess() bool {
	_, ok := f.v.(fallibleSuccess[T])
	return ok
}

// IsSuccess returns true if the Fallible is representing an error.
func (f Fallible[T]) IsError() bool {
	_, ok := f.v.(fallibleFailure)
	return ok
}

// Err returns a non-nil error if the fallible represents failure, or a nil
// error if it represents success.
func (f Fallible[T]) Err() error {
	if failed, ok := f.v.(fallibleFailure); ok {
		return failed[0]
	}
	return nil
}

// Success creates a Fallible that represents success with the given value.
func Success[T any](v T) Fallible[T] {
	return Fallible[T]{fallibleSuccess[T]{v}}
}

// Error creates a Fallible that represents failure with a given error.
//
// The given error must not be nil or this function will panic.
func Error[T any](err error) Fallible[T] {
	if err == nil {
		panic("walk.FallibleFailure with nil error")
	}
	return Fallible[T]{fallibleFailure{err}}
}

// Void represents the absense of a value, for walks that don't need to
// actually produce a result.
//
// The only possible value of Void is nil.
type Void interface {
	thereAreNoValuesOfThisType()
}

type fallibleOutcome interface {
	fallibleOutcomeSigil()
}

type fallibleSuccess[T any] [1]T

func (fallibleSuccess[T]) fallibleOutcomeSigil() {}

type fallibleFailure [1]error

func (fallibleFailure) fallibleOutcomeSigil() {}
