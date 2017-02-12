package cty

// primitiveType is the hidden implementation of the various primitive types
// that are exposed as variables in this package.
type primitiveType struct {
	typeImpl
	Name string
}

func (t *primitiveType) Equals(other Type) bool {
	if otherP, ok := other.(*primitiveType); ok {
		return otherP.Name == t.Name
	}
	return false
}

func (t *primitiveType) FriendlyName() string {
	return t.Name
}

// Number is the numeric type. Number values are arbitrary-precision
// decimal numbers, which can then be converted into Go's various numeric
// types only if they are in the appropriate range.
var Number Type = &primitiveType{Name: "number"}

// String is the string type. String values are sequences of unicode codepoints
// encoded internally as UTF-8.
var String Type = &primitiveType{Name: "string"}

// Bool is the boolean type. The two values of this type are True and False.
var Bool Type = &primitiveType{Name: "bool"}

// True is the truthy value of type Bool
var True = trueValue
var trueValue = Value{
	ty: Bool,
	v:  true,
}

// False is the falsey value of type Bool
var False = falseValue
var falseValue = Value{
	ty: Bool,
	v:  false,
}

// IsPrimitiveType returns true if and only if the given type is a primitive
// type, which means it's either number, string, or bool. Any two primitive
// types can be safely compared for equality using the standard == operator
// without panic, which is not a guarantee that holds for all types. Primitive
// types can therefore also be used in switch statements.
func IsPrimitiveType(t Type) bool {
	_, ok := t.(*primitiveType)
	return ok
}
