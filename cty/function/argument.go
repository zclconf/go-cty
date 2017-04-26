package function

import (
	"github.com/apparentlymart/go-cty/cty"
)

// Parameter represents a parameter to a function.
type Parameter struct {
	// Name is an optional name for the argument. This package ignores this
	// value, but callers may use it for documentation, etc.
	Name string

	// A type that any argument for this parameter must conform to.
	// cty.DynamicPseudoType can be used, either at top-level or nested
	// in a parameterized type, to indicate that any type should be
	// permitted, to allow the definition of type-generic functions.
	Type cty.Type

	// If AllowNull is set then null values may be passed into this
	// argument's slot in both the type-check function and the implementation
	// function. If not set, such values are rejected by the built-in
	// checking rules.
	AllowNull bool

	// If AllowUnknown is set then unknown values may be passed into this
	// argument's slot in the implementation function. If not set, any
	// unknown values will cause the function to immediately return
	// an unkonwn value without calling the implementation function, thus
	// freeing the function implementer from dealing with this case.
	AllowUnknown bool

	// If AllowDynamicType is set then DynamicVal may be passed into this
	// argument's slot in the implementation function. If not set, any
	// dynamic values will cause the function to immediately return
	// DynamicVal value without calling the implementation function, thus
	// freeing the function implementer from dealing with this case.
	AllowDynamicType bool
}
