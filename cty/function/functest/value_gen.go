package functest

import (
	"math"
	"math/big"
	"math/rand"
	"strings"
	"unicode/utf8"

	"github.com/zclconf/go-cty/cty"
)

// ValueGen is a function type for Go functions that use a given random number
// generator to produce a cty value meeting some constraint documented for
// that particular function.
type ValueGen func(rand *rand.Rand) cty.Value

// Map creates a new ValueGen that wraps the reciever and applies the given
// map function to whatever result it produces.
func (gen ValueGen) Map(f func(v cty.Value) cty.Value) ValueGen {
	return func(rand *rand.Rand) cty.Value {
		return f(gen(rand))
	}
}

// Where creates a new ValueGen that wraps the reciever and keeps calling
// it repeatedly until it returns a value that the given function will
// return true for.
//
// Note that the check function ought to be one that returns true for most
// of the values in the domain of the wrapped check function, or else this
// function might spin for a long time before selecting a value. As a safety
// check against check functions that never succeed, this function will panic
// if it fails to produce a suitable value after 1,000 tries.
func (gen ValueGen) Where(f func(v cty.Value) bool) ValueGen {
	return func(rand *rand.Rand) cty.Value {
		for i := 0; i < 1000; i++ {
			candidate := gen(rand)
			if f(candidate) {
				return candidate
			}
		}
		panic("no valid result after 1000 attempts")
	}
}

// GenValueFromAnyOf creates a ValueGen that randomly chooses any one of the
// given ValueGens, with equal probability, and returns the value generated by
// that particular generator.
func GenValueFromAnyOf(gens ...ValueGen) ValueGen {
	return func(rand *rand.Rand) cty.Value {
		max := len(gens)
		gen := gens[rand.Intn(max)]
		return gen(rand)
	}
}

// MaybeDynamicVal creates a new ValueGen that wraps the receiver and adds a
// one in fifty chance of skipping a call to the wrapped generator and just
// returning cty.DynamicVal (an unknown value of unknown type) instead.
func (gen ValueGen) MaybeDynamicVal() ValueGen {
	return func(rand *rand.Rand) cty.Value {
		roll := rand.Intn(50)
		if roll == 0 {
			return cty.DynamicVal
		}
		return gen(rand)
	}
}

// MaybeUnknown creates a new ValueGen that wraps the receiver and adds a
// one in fifty chance of turning the result into an unknown value of the
// same type as the underlying generator produced.
func (gen ValueGen) MaybeUnknown() ValueGen {
	return func(rand *rand.Rand) cty.Value {
		roll := rand.Intn(50)
		v := gen(rand)
		if roll == 0 {
			return cty.UnknownVal(v.Type())
		}
		return v
	}
}

// MaybeMarked creates a new ValueGen that wraps the receiver and adds a
// one in fifty chance of turning the result into a marked version of the
// same value the generator produced.
func (gen ValueGen) MaybeMarked(mark interface{}) ValueGen {
	return func(rand *rand.Rand) cty.Value {
		roll := rand.Intn(50)
		v := gen(rand)
		if roll == 0 {
			return v.Mark(mark)
		}
		return v
	}
}

// MaybeAnnotated is a convenience helper that is the same as calling
// .MaybeUnknown().MaybeMarked("") on the same reciever, as an easy way to
// opt in to both of the special cases that functions are typically expected
// to handle gracefully (but often inadvertenly don't).
func (gen ValueGen) MaybeAnnotated() ValueGen {
	return gen.MaybeUnknown().MaybeMarked("")
}

// MaybeNull creates a new ValueGen that wraps the receiver and adds a
// one in fifty chance of turning the result into a null value of the
// same type as the underlying generator produced.
func (gen ValueGen) MaybeNull() ValueGen {
	return func(rand *rand.Rand) cty.Value {
		roll := rand.Intn(50)
		v := gen(rand)
		if roll == 0 {
			return cty.NullVal(v.Type())
		}
		return v
	}
}

// MaybeInfinity creates a new ValueGen that wraps the receiver and adds a
// one in 75 chance of turning the result into either cty.PositiveInfinity
// or cty.NegativeInfinity, with equal probabilities between the two
// integers.
//
// It generally only makes sense to use this method with a ValueGen that
// could potentially return numbers.
func (gen ValueGen) MaybeInfinity() ValueGen {
	return func(rand *rand.Rand) cty.Value {
		roll := rand.Intn(75 * 2)
		switch roll {
		case 0:
			return cty.PositiveInfinity
		case 1:
			return cty.NegativeInfinity
		default:
			return gen(rand)
		}
	}
}

// GenConstant creates a ValueGen that just always returns the given constant
// value.
func GenConstant(v cty.Value) ValueGen {
	return func(rand *rand.Rand) cty.Value {
		return v
	}
}

// GenBools creates a ValueGen that has an equal chance of returning either
// cty.True or cty.False.
func GenBools() ValueGen {
	return func(rand *rand.Rand) cty.Value {
		n := rand.Intn(2)
		if n == 0 {
			return cty.False
		}
		return cty.True
	}
}

// GenNumbers creates a ValueGen that returns arbitrary cty.Number-typed values.
//
// GenInteger can generate integers outside of the range representable by
// Go's built-in numeric types, so it's good for testing arithmetic functions
// that work with integers but not suitable for functions that expect a more
// constrained integer range.
//
// There is a 50% chance that the result will be an integer. If you don't need
// that bias in the results, consider GenFloats instead.
//
// GenNumbers alone will never generate infinities. Use the MaybeInfinity
// method on the result to add the possibility of infinities.
func GenNumbers() ValueGen {
	return GenValueFromAnyOf(GenIntegers(), GenFloats())
}

// GenIntegers creates a ValueGen that returns arbitrary cty.Number-typed values
// that are all guaranteed to be integers.
//
// GenIntegers can generate integers outside of the range representable by
// Go's built-in integer types, so it's good for testing arithmetic functions
// that work with integers but not suitable for functions that expect a more
// constrained integer range.
//
// GenIntegers alone will never generate infinities. Use the MaybeInfinity
// method on the result to add the possibility of infinities.
func GenIntegers() ValueGen {
	return func(rand *rand.Rand) cty.Value {
		// Our result will be the sum of two randomly-selected signed integers,
		// which can therefore generate both positive and negative numbers
		// outside of the range of an int64 or uint64.
		v1 := int64(rand.Uint64())
		v2 := int64(rand.Uint64())

		// What we're doing here is essentially rolling two dice, which
		// produces a very uneven probability distribution. As an imprecise
		// way to deal with that, we'll treat part of the second result as
		// a chance of ignoring that result altogether, and thus behaving
		// as if we only rolled one dice within the uint64 range.
		if v1 == 0 {
			// Zero is a common edge case, so we'll artificially increase
			// our probability of returning it by ignoring v2 if v1 is zero.
			return cty.NumberIntVal(0)
		}
		if v2 > (math.MaxInt64 - math.MaxInt32) {
			v2 = 0
		}

		var v1f big.Float
		v1f.SetPrec(128)
		v1f.SetInt64(v1)

		var v2f big.Float
		v2f.SetPrec(128)
		v2f.SetInt64(v2)

		v1f.Add(&v1f, &v2f)

		return cty.NumberVal(&v1f)
	}
}

// GenFloats creates a ValueGen that returns arbitrary cty.Number-typed
// values.
//
// GenFloats can generate numbers outside of the range representable by
// Go's built-in float64 type, so it's good for testing arithmetic functions
// but not suitable for functions that expect a more constrained number range.
//
// GenFloats alone will never generate infinities. Use the MaybeInfinity
// method on the result to add the possibility of infinities.
func GenFloats() ValueGen {
	return func(rand *rand.Rand) cty.Value {
		// Our result will be the sum of two randomly-selected signed floats,
		// which can therefore generate numbers outside of the range of a
		// float64.
		v1 := rand.NormFloat64()
		v2 := rand.NormFloat64()

		// What we're doing here is essentially rolling two dice, which
		// produces a very uneven probability distribution. As an imprecise
		// way to deal with that, we'll treat part of the second result as
		// a chance of ignoring that result altogether, and thus behaving
		// as if we only rolled one dice within the float64 range.
		if v1 == 0 {
			// Zero is a common edge case, so we'll artificially increase
			// our probability of returning it by ignoring v2 if v1 is zero.
			return cty.NumberIntVal(0)
		}
		if v2 > (math.MaxFloat64 - math.MaxFloat32) {
			v2 = 0.0
		}

		var v1f big.Float
		v1f.SetPrec(128)
		v1f.SetFloat64(v1)

		var v2f big.Float
		v2f.SetPrec(128)
		v2f.SetFloat64(v2)

		v1f.Add(&v1f, &v2f)

		return cty.NumberVal(&v1f)
	}
}

// GenStrings creates a ValueGen that returns arbitrary string values, which
// each contain between zero and twenty unicode code units where all are
// guaranteed to be UTF-8-encodeable but otherwise unconstrained.
//
// This function doesn't have any awareness of the Unicode character
// segmentation rules however, so it may generate unreasonable sequences that
// don't combine, such as emoji skin tone modifiers associated with non-emoji
// characters, or combining diacritics alongside characters that can't
// reasonably accept them. Perhaps in a future version it will become more
// reasonable though, so don't rely on it to generate unrealistic nonsense.
func GenStrings() ValueGen {
	const bmpLimit = rune(65536)
	const overallLimit = utf8.MaxRune
	const astralLimit = overallLimit - bmpLimit
	// We are four times as likely to generate characters in the basic
	// multilingual plane as to generate characters in the supplementary
	// (aka "astral") planes.
	const runeRandBmpRange = bmpLimit * 4
	const runeRandLimit = int(runeRandBmpRange + astralLimit)
	genRune := func(rand *rand.Rand) rune {
		for {
			n := rune(rand.Intn(runeRandLimit))
			if n < runeRandBmpRange {
				n = n >> 2 // divide by four to undo our weighting
			} else {
				// Adjust for our weighting
				n = n - runeRandBmpRange + bmpLimit
			}
			if utf8.ValidRune(n) {
				return n
			}
			// We'll keep trying until we find something UTF-8-encodable.
			// There are vastly more encodable than non-encodable
			// characters under utf8.MaxRune, so this shouldn't take long.
		}
	}
	return func(rand *rand.Rand) cty.Value {
		n := rand.Intn(maxGeneratedWidth)
		if n == 0 {
			return cty.StringVal("")
		}
		var buf strings.Builder
		buf.Grow(n * utf8.UTFMax)
		for i := 0; i < n; i++ {
			r := genRune(rand)
			buf.WriteRune(r)
		}
		return cty.StringVal(buf.String())
	}
}

// GenLists creates a ValueGen that returns list values whose elements are
// all within the range of the given generator, and with a length between the
// minimum and maximum provided, both inclusive.
//
// The given ValueGen must generate values that are all of the same type, or
// else this function will attempt to create an invalid list and therefore
// panic.
func GenLists(values ValueGen, minLen, maxLen int) ValueGen {
	return func(rand *rand.Rand) cty.Value {
		count := minLen
		if diff := maxLen - minLen; diff > 0 {
			count += rand.Intn(diff)
		}
		if count == 0 {
			// For an empty list we generate a value just to get
			// an element type, discarding the value itself.
			typeVal := values(rand)
			return cty.ListValEmpty(typeVal.Type())
		}
		elems := make([]cty.Value, count)
		for i := range elems {
			elems[i] = values(rand)
		}
		return cty.ListVal(elems)
	}
}

// GenTuples creates a ValueGen that returns list values whose elements are
// all within the range of the given generator, and with a length between the
// minimum and maximum provided, both inclusive.
func GenTuples(values ValueGen, minLen, maxLen int) ValueGen {
	return func(rand *rand.Rand) cty.Value {
		count := minLen
		if diff := maxLen - minLen; diff > 0 {
			count += rand.Intn(diff)
		}
		if count == 0 {
			return cty.EmptyTupleVal
		}
		elems := make([]cty.Value, count)
		for i := range elems {
			elems[i] = values(rand)
		}
		return cty.TupleVal(elems)
	}
}

// GenSets creates a ValueGen that returns set values whose elements are
// all within the range of the given generator, and with a length between the
// minimum and maximum provided, both inclusive.
//
// The given ValueGen must generate values that are all of the same type, or
// else this function will attempt to create an invalid set and therefore
// panic.
func GenSets(values ValueGen, minLen, maxLen int) ValueGen {
	return func(rand *rand.Rand) cty.Value {
		count := minLen
		if diff := maxLen - minLen; diff > 0 {
			count += rand.Intn(diff)
		}
		if count == 0 {
			// For an empty list we generate a value just to get
			// an element type, discarding the value itself.
			typeVal := values(rand)
			return cty.SetValEmpty(typeVal.Type())
		}
		elems := make([]cty.Value, count)
		for i := range elems {
			elems[i] = values(rand)
		}
		return cty.SetVal(elems)
	}
}

// GenMaps creates a ValueGen that returns map values whose elements are
// all within the ranges of the given generators, and with a length between the
// minimum and maximum provided, both inclusive.
//
// The keys ValueGen must generate only known, non-null, unmarked values of
// type cty.String to use as map keys, or this function will panic.
//
// The values ValueGen must generate values that are all of the same type, or
// else this function will attempt to create an invalid map and therefore
// panic.
func GenMaps(keys, values ValueGen, minLen, maxLen int) ValueGen {
	return func(rand *rand.Rand) cty.Value {
		count := minLen
		if diff := maxLen - minLen; diff > 0 {
			count += rand.Intn(diff)
		}
		if count == 0 {
			// For an empty list we generate a value just to get
			// an element type, discarding the value itself.
			typeVal := values(rand)
			return cty.MapValEmpty(typeVal.Type())
		}
		elems := make(map[string]cty.Value, count)
		for i := 0; i < count; i++ {
			// We need to find a key we're not already using.
			var key string
			for {
				key = keys(rand).AsString()
				if _, exists := elems[key]; !exists {
					break
				}
			}

			value := values(rand)
			elems[key] = value
		}
		return cty.MapVal(elems)
	}
}

// GenObjects creates a ValueGen that returns object values whose attributes
// are all within the ranges of the given generators, and with a length between
// the minimum and maximum provided, both inclusive.
//
// The keys ValueGen must generate only known, non-null, unmarked values of
// type cty.String to use as map keys, or this function will panic.
func GenObjects(keys, values ValueGen, minLen, maxLen int) ValueGen {
	return func(rand *rand.Rand) cty.Value {
		count := minLen
		if diff := maxLen - minLen; diff > 0 {
			count += rand.Intn(diff)
		}
		if count == 0 {
			return cty.EmptyObjectVal
		}
		elems := make(map[string]cty.Value, count)
		for i := 0; i < count; i++ {
			// We need to find a key we're not already using.
			var key string
			for {
				key = keys(rand).AsString()
				if _, exists := elems[key]; !exists {
					break
				}
			}

			value := values(rand)
			elems[key] = value
		}
		return cty.ObjectVal(elems)
	}
}

// GenAnyValuesOfType creates a ValueGen that returns arbitrary cty.Value
// values which conform to the given type constraint, including unknown
// and marked values.
//
// Unlike most of the value generator functions, this one has a chance of
// producing null, unknown, and marked values by default, without the need to
// wrap it afterwards with the MaybeNull, MaybeUnknown, or MaybeMarked methods.
//
// In the full space of possible cty values there are very wide and deep
// data structures that would be time-consuming or memory-consuming to run
// tests against, so as a measure of pragmatism this function has some
// reasonable limits: it will never generate a collection with more than
// 20 elements and it will never create a data structure with more than
// three levels of nesting.
func GenAnyValuesOfType(ty cty.Type) ValueGen {
	return func(rand *rand.Rand) cty.Value {
		return genValueOfType(rand, ty, true, maxGeneratedDepth)
	}
}

// GenAnyValues creates a ValueGen that returns arbitrary cty.Value
// values, of any type or structure, including unknown and marked values.
//
// Unlike most of the value generator functions, this one has a chance of
// producing null, unknown, and marked values by default, without the need to
// wrap it afterwards with the MaybeNull, MaybeUnknown, or MaybeMarked methods.
//
// In the full space of possible cty values there are very wide and deep
// data structures that would be time-consuming or memory-consuming to run
// tests against, so as a measure of pragmatism this function has some
// reasonable limits: it will never generate a collection with more than
// 20 elements and it will never create a data structure with more than
// three levels of nesting.
func GenAnyValues() ValueGen {
	return func(rand *rand.Rand) cty.Value {
		return genAnyValue(rand, true, maxGeneratedDepth)
	}
}

// GenAnySerializableValues creates a ValueGen that returns arbitrary cty.Value
// values, of any type or structure, except for unknown and marked values.
// Unknown and marked values are the ones that our general serializers
// typically can't handle, which is the motivation for this function's
// distinction from GenAnyValues.
//
// Unlike most of the value generator functions, this one has a chance of
// producing null values by default, without the need to wrap it afterwards
// with the MaybeNull method.
//
// Aside from the omission of unknown and marked values this generator has
// the same range as GenAnyValues, including the restrictions on width and
// depth of the generated data structures.
func GenAnySerializableValues() ValueGen {
	return func(rand *rand.Rand) cty.Value {
		return genAnyValue(rand, false, maxGeneratedDepth)
	}
}

// GenSerializableValuesOfType creates a ValueGen that returns arbitrar
// cty.Value values which conform to the given type constraint, excluding
// unknown and marked values.
//
// Unlike most of the value generator functions, this one has a chance of
// producing null by default, without the need to wrap it afterwards with the
// MaybeNull method.
//
// In the full space of possible cty values there are very wide and deep
// data structures that would be time-consuming or memory-consuming to run
// tests against, so as a measure of pragmatism this function has some
// reasonable limits: it will never generate a collection with more than
// 20 elements and it will never create a data structure with more than
// three levels of nesting.
func GenSerializableValuesOfType(ty cty.Type) ValueGen {
	return func(rand *rand.Rand) cty.Value {
		return genValueOfType(rand, ty, false, maxGeneratedDepth)
	}
}

const maxGeneratedDepth = 3
const maxGeneratedWidth = 20

func genValueOfType(rand *rand.Rand, ty cty.Type, unknownMarked bool, nestCount int) cty.Value {
	switch {

	case ty == cty.String:
		return GenStrings()(rand)

	case ty == cty.Number:
		return GenNumbers()(rand)

	case ty == cty.Bool:
		return GenBools()(rand)

	case ty.IsListType():
		ety := ty.ElementType()
		return GenLists(arbitrary(ety, unknownMarked, nestCount-1), 0, maxGeneratedWidth)(rand)

	case ty.IsSetType():
		ety := ty.ElementType()
		return GenSets(arbitrary(ety, unknownMarked, nestCount-1), 0, maxGeneratedWidth)(rand)

	case ty.IsMapType():
		ety := ty.ElementType()
		return GenMaps(GenStrings(), arbitrary(ety, unknownMarked, nestCount-1), 0, maxGeneratedWidth)(rand)

	case ty.IsTupleType():
		etys := ty.TupleElementTypes()
		if len(etys) == 0 {
			return cty.EmptyTupleVal
		}
		elems := make([]cty.Value, len(etys))
		for i, ety := range etys {
			elems[i] = genValueOfType(rand, ety, unknownMarked, nestCount-1)
		}
		return cty.TupleVal(elems)

	case ty.IsObjectType():
		atys := ty.AttributeTypes()
		if len(atys) == 0 {
			return cty.EmptyObjectVal
		}
		elems := make(map[string]cty.Value, len(atys))
		for k, aty := range atys {
			elems[k] = genValueOfType(rand, aty, unknownMarked, nestCount-1)
		}
		return cty.ObjectVal(elems)

	default:
		// The main way to get here is if the randomly-selected type was
		// cty.DynamicPseudoType, which therefore limits our options
		// significantly.
		//
		// We might also get here if there are capsule types in the mix, but
		// the other cases above ought to be comprehensive for all of the
		// built-in types.
		if !unknownMarked {
			return cty.NullVal(ty)
		}
		n := rand.Intn(4)
		if n == 0 {
			return cty.UnknownVal(ty)
		}
		return cty.NullVal(ty)
	}
}

func arbitrary(ty cty.Type, unknownMarked bool, nestCount int) ValueGen {
	return func(rand *rand.Rand) cty.Value {
		return genValueOfType(rand, ty, unknownMarked, nestCount)
	}
}

func genAnyValue(rand *rand.Rand, unknownMarked bool, nestCount int) cty.Value {
	ty := genAnyType(rand, nestCount)
	v := genValueOfType(rand, ty, unknownMarked, nestCount)
	n := rand.Intn(100)
	switch {
	case n < 10:
		v = cty.NullVal(v.Type())
	case unknownMarked && n < 20:
		v = cty.UnknownVal(v.Type())
	}
	if unknownMarked {
		// markedness is independent of nullness or unknownness
		n = rand.Intn(100)
		if n < 5 {
			v = v.Mark("")
		}
	}
	return v
}

func genAnyType(rand *rand.Rand, nestCount int) cty.Type {
	// Our decisions here are weighted towards generating primitive-typed
	// values, because lots of nested structures tends to make our operations
	// expensive and we're aiming to test correctness rather than performance
	// with these generators. While it's possible that a function could
	// have buggy behaviors that only appear for more complex structures,
	// that seems to have been pretty uncommon in practice and so we'd rather
	// not waste the time repeatedly running expensive tests for that rare
	// possibility.

	// We can always generate primitive types. If nestCount is greater than
	// zero then we can also potentially generate collection and structural
	// types. We also have a relatively low chance of generating
	// DynamicPseudoType.
	const boolWeight = 5 // special because there are only a few distinct bool values
	const primitiveWeight = 10
	const collectionWeight = 4
	const structuralWeight = 3
	const dynamicWeight = 1
	max := (boolWeight) + (primitiveWeight * 2) + (dynamicWeight)
	if nestCount > 0 {
		max += (collectionWeight * 3) + (structuralWeight * 2)
	}
	n := rand.Intn(max)
	switch n {
	case 0, 1, 2, 3, 4:
		return cty.Bool
	case 5, 6, 7, 8, 9, 10, 11, 12, 13, 14:
		return cty.Number
	case 15, 16, 17, 18, 19, 20, 21, 22, 23, 24:
		return cty.String
	case 25:
		return cty.DynamicPseudoType

	// The remainder of these cases are reachable only if nestCount > 0
	// and therefore we made "max" larger above.

	case 26, 27, 28, 29:
		ety := genAnyType(rand, nestCount-1)
		return cty.List(ety)
	case 30, 31, 32, 33:
		ety := genAnyType(rand, nestCount-1)
		return cty.Set(ety)
	case 34, 35, 36, 37:
		ety := genAnyType(rand, nestCount-1)
		return cty.Map(ety)
	case 38, 39, 40:
		elemCount := rand.Intn(maxGeneratedWidth)
		if elemCount == 0 {
			return cty.EmptyTuple
		}
		etys := make([]cty.Type, elemCount)
		for i := range etys {
			etys[i] = genAnyType(rand, nestCount-1)
		}
		return cty.Tuple(etys)
	case 41, 42, 43:
		elemCount := rand.Intn(maxGeneratedWidth)
		if elemCount == 0 {
			return cty.EmptyObject
		}
		atys := make(map[string]cty.Type, elemCount)
		for i := 0; i < elemCount; i++ {
			key := GenStrings()(rand)
			aty := genAnyType(rand, nestCount-1)
			atys[key.AsString()] = aty
		}
		return cty.Object(atys)
	default:
		// Shouldn't get here, because our cases above should cover the full
		// range of our rand.Intn call above.
		panic("unhandled random number")
	}
}
