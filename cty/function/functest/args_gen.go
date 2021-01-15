package functest

import (
	"math/rand"
	"reflect"

	"github.com/zclconf/go-cty/cty"
)

// ArgsGen is a function type for Go functions that use a given random number
// generator to produce an sequence of function arguments (a slice of cty.Value)
// meeting some constraints documented for that particular function.
type ArgsGen func(rand *rand.Rand) []cty.Value

// ReflectValues wraps the receiver in a function whose signature is comaptible
// with the Values field in the Config type in the standard library package
// testing/quick.
//
// The returned argument lists actually always have length one and consist
// of a reflect.Value wrapper around the generated []cty.Value, which is
// compatible with the Call method of Function in package function (in the
// parent directory of this package).
func (gen ArgsGen) ReflectValues() func([]reflect.Value, *rand.Rand) {
	return func(into []reflect.Value, rand *rand.Rand) {
		if len(into) != 1 {
			panic("ArgsGen.ReflectValues works only with functions that accept a single argument of type []cty.Value")
		}
		result := gen(rand)
		into[0] = reflect.ValueOf(result)
	}
}

// GenFixedArgs generates a fixed number of arguments, one per given ValueGen
// function.
func GenFixedArgs(gens ...ValueGen) ArgsGen {
	return func(rand *rand.Rand) []cty.Value {
		if len(gens) == 0 {
			return nil
		}
		var ret = make([]cty.Value, len(gens))
		for i, gen := range gens {
			ret[i] = gen(rand)
		}
		return ret
	}
}

// GenVarArgs generates a variable number of arguments all from the same
// generator.
//
// The minCount and maxCount arguments define the range of argument list
// lengths that may be produced, both of which are inclusive. Set both
// arguments to the same value to require a specific number of arguments.
func GenVarArgs(gen ValueGen, minCount, maxCount int) ArgsGen {
	return func(rand *rand.Rand) []cty.Value {
		count := minCount
		if diff := (maxCount - minCount); diff > 0 {
			count += rand.Intn(diff)
		}
		if count == 0 {
			return nil
		}
		var ret = make([]cty.Value, count)
		for i := range ret {
			ret[i] = gen(rand)
		}
		return ret
	}
}

// ConcatArgs returns an ArgsGen which runs all of the given ArgsGen in the
// given order and then concatenates the results together to return.
//
// This is useful, for example, for concatenating a GenFixedArgs result with
// a GenVarArgs result in order to generate arguments for a function that
// has a mixture of both fixed and variadic arguments.
func ConcatArgs(gens ...ArgsGen) ArgsGen {
	return func(rand *rand.Rand) []cty.Value {
		var ret []cty.Value
		for _, gen := range gens {
			ret = append(ret, gen(rand)...)
		}
		return ret
	}
}
