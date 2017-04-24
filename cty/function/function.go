package function

import (
	"github.com/apparentlymart/go-cty/cty"
)

// Function represents a function. This is the main type in this package.
type Function struct {
	params   []Parameter
	varParam *Parameter
	typeFunc TypeFunc
	implFunc ImplFunc
}

// Spec is the specification of a function, used to instantiate
// a new Function.
type Spec struct {
	// Params is a description of the positional parameters for the function.
	// The standard checking logic rejects any calls that do not provide
	// arguments conforming to this definition, freeing the function
	// implementer from dealing with such inconsistencies.
	Params []Parameter

	// VarParam is an optional specification of additional "varargs" the
	// function accepts. If this is non-nil then callers may provide an
	// arbitrary number of additional arguments (after those matching with
	// the fixed parameters in Params) that conform to the given specification,
	// which will appear as additional values in the slices of values
	// provided to the type and implementation functions.
	VarParam *Parameter

	// Type is the TypeFunc that decides the return type of the function
	// given its arguments, which may be Unknown. See the documentation
	// of TypeFunc for more information.
	//
	// Use StaticReturnType if the function's return type does not vary
	// depending on its arguments.
	//
	// Type may be nil, in which case the ImplFunc given in Impl is instead
	// used to determine the result type. This is appropriate only for
	// functions whose implementation is not significantly more expensive
	// than the type check would be, but can avoid the need to write a separate
	// type checking function in that case.
	Type TypeFunc

	// Impl is the ImplFunc that implements the function's behavior.
	//
	// Functions are expected to behave as pure functions, and not create
	// any visible side-effects.
	//
	// If a TypeFunc is also provided, the value returned from Impl *must*
	// conform to the type it returns, or a call to the function will panic.
	Impl ImplFunc
}

// New creates a new function with the given specification.
//
// After passing a Spec to this function, the caller must not make any further
// modifications to it.
func New(spec *Spec) Function {
	typeFunc := spec.Type
	if typeFunc == nil {
		typeFunc = defaultTypeFunc(spec.Impl)
	}
	return Function{
		params:   spec.Params,
		varParam: spec.VarParam,
		typeFunc: typeFunc,
		implFunc: spec.Impl,
	}
}

// TypeFunc is a callback type for determining the return type of a function
// given its arguments.
//
// Any of the values passed to this function may be unknown, even if the
// parameters are not configured to accept unknowns.
//
// If any of the given values are *not* unknown, the TypeFunc may use the
// values for pre-validation and for choosing the return type. For example,
// a hypothetical JSON-unmarshalling function could return
// cty.DynamicPseudoType if the given JSON string is unknown, but return
// a concrete type based on the JSON structure if the JSON string is already
// known.
type TypeFunc func(args []cty.Value) (cty.Type, error)

// ImplFunc is a callback type for the main implementation of a function.
type ImplFunc func(args []cty.Value) (cty.Value, error)

// StaticReturnType returns a TypeFunc that always returns the given type.
//
// This is provided as a convenience for defining a function whose return
// type does not depend on the argument types.
func StaticReturnType(ty cty.Type) TypeFunc {
	return func([]cty.Value) (cty.Type, error) {
		return ty, nil
	}
}

// defaultTypeFunc wraps an implementation func and returns the type of
// the value it returns.
func defaultTypeFunc(implFunc ImplFunc) TypeFunc {
	return func(args []cty.Value) (cty.Type, error) {
		val, err := implFunc(args)
		if err != nil {
			return cty.Type{}, err
		}
		return val.Type(), nil
	}
}
