package convert

import (
	"math/big"
	"reflect"

	"math"

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
	tVal := reflect.ValueOf(target)
	if tVal.Kind() != reflect.Ptr {
		panic("target value is not a pointer")
	}
	if tVal.IsNil() {
		panic("target value is nil pointer")
	}

	// 'path' starts off as empty but will grow for each level of recursive
	// call we make, so by the time fromCtyValue returns it is likely to have
	// unused capacity on the end of it, depending on how deeply-recursive
	// the given cty.Value is.
	path := make(cty.Path, 0)
	return fromCtyValue(val, tVal, path)
}

func fromCtyValue(val cty.Value, target reflect.Value, path cty.Path) error {
	ty := val.Type()

	deepTarget := fromCtyPopulatePtr(target, false)

	// If we're decoding into a cty.Value then we just pass through the
	// value as-is, to enable partial decoding. This is the only situation
	// where unknown values are permitted.
	if deepTarget.Kind() == reflect.Struct && deepTarget.Type().AssignableTo(valueType) {
		deepTarget.Set(reflect.ValueOf(val))
		return nil
	}

	// Lists and maps can be nil without indirection, but everything else
	// requires a pointer and we set it immediately to nil.
	// (fromCtyList and fromCtyMap must therefore deal with val.IsNull, while
	// other types can assume no nulls after this point.)
	if val.IsNull() && !val.Type().IsListType() && !val.Type().IsMapType() {
		target = fromCtyPopulatePtr(target, true)
		if target.Kind() != reflect.Ptr {
			return errorf(path, "null value is not allowed")
		}

		target.Set(reflect.Zero(target.Type()))
		return nil
	}

	target = deepTarget

	if !val.IsKnown() {
		return errorf(path, "value must be known")
	}

	// Converting into interface{} is allowed, in which case we use a default
	// set of conversions that are non-lossy but may not be convienient to
	// the caller.
	if target.Kind() == reflect.Interface && emptyInterfaceType.AssignableTo(target.Type()) {
		return fromCtyDynamic(val, target, path)
	}

	switch ty {
	case cty.Bool:
		return fromCtyBool(val, target, path)
	case cty.Number:
		return fromCtyNumber(val, target, path)
	case cty.String:
		return fromCtyString(val, target, path)
	}

	switch {
	case ty.IsListType():
		return fromCtyList(val, target, path)
	case ty.IsMapType():
		return fromCtyMap(val, target, path)
	case ty.IsSetType():
		return fromCtySet(val, target, path)
	case ty.IsObjectType():
		return fromCtyObject(val, target, path)
	}

	// We should never fall out here; reaching here indicates a bug in this
	// function.
	return errorf(path, "unsupported source type %#v", ty)
}

func fromCtyBool(val cty.Value, target reflect.Value, path cty.Path) error {
	switch target.Kind() {

	case reflect.Bool:
		if val.True() {
			target.Set(reflect.ValueOf(true))
		} else {
			target.Set(reflect.ValueOf(false))
		}
		return nil

	default:
		return likelyRequiredTypesError(path, target)

	}
}

func fromCtyNumber(val cty.Value, target reflect.Value, path cty.Path) error {
	bf := val.AsBigFloat()

	switch target.Kind() {

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fromCtyNumberInt(bf, target, path)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fromCtyNumberUInt(bf, target, path)

	case reflect.Float32, reflect.Float64:
		return fromCtyNumberFloat(bf, target, path)

	case reflect.Struct:
		return fromCtyNumberBig(bf, target, path)

	default:
		return likelyRequiredTypesError(path, target)

	}
}

func fromCtyNumberInt(bf *big.Float, target reflect.Value, path cty.Path) error {
	// Doing this with switch rather than << arithmetic because << with
	// result >32-bits is not portable to 32-bit systems.
	var min int64
	var max int64
	switch target.Type().Bits() {
	case 8:
		min = math.MinInt8
		max = math.MaxInt8
	case 16:
		min = math.MinInt16
		max = math.MaxInt16
	case 32:
		min = math.MinInt32
		max = math.MaxInt32
	case 64:
		min = math.MinInt64
		max = math.MaxInt64
	default:
		panic("weird number of bits in target int")
	}

	iv, accuracy := bf.Int64()
	if accuracy != big.Exact || iv < min || iv > max {
		return errorf(path, "value must be a whole number, between %d and %d", min, max)
	}

	target.Set(reflect.ValueOf(iv).Convert(target.Type()))

	return nil
}

func fromCtyNumberUInt(bf *big.Float, target reflect.Value, path cty.Path) error {
	// Doing this with switch rather than << arithmetic because << with
	// result >32-bits is not portable to 32-bit systems.
	var max uint64
	switch target.Type().Bits() {
	case 8:
		max = math.MaxUint8
	case 16:
		max = math.MaxUint16
	case 32:
		max = math.MaxUint32
	case 64:
		max = math.MaxUint64
	default:
		panic("weird number of bits in target uint")
	}

	iv, accuracy := bf.Uint64()
	if accuracy != big.Exact || iv > max {
		return errorf(path, "value must be a whole number, between 0 and %d inclusive", max)
	}

	target.Set(reflect.ValueOf(iv).Convert(target.Type()))

	return nil
}

func fromCtyNumberFloat(bf *big.Float, target reflect.Value, path cty.Path) error {
	switch target.Kind() {
	case reflect.Float32:
		fv, accuracy := bf.Float32()
		if accuracy != big.Exact {
			// We allow the precision to be truncated as part of our conversion,
			// but we don't want to silently introduce infinities.
			if math.IsInf(float64(fv), 0) {
				return errorf(path, "value must be between %f and %f inclusive", -math.MaxFloat32, math.MaxFloat32)
			}
		}
		target.Set(reflect.ValueOf(fv))
		return nil
	case reflect.Float64:
		fv, accuracy := bf.Float64()
		if accuracy != big.Exact {
			// We allow the precision to be truncated as part of our conversion,
			// but we don't want to silently introduce infinities.
			if math.IsInf(fv, 0) {
				return errorf(path, "value must be between %f and %f inclusive", -math.MaxFloat64, math.MaxFloat64)
			}
		}
		target.Set(reflect.ValueOf(fv))
		return nil
	default:
		panic("unsupported kind of float")
	}
}

func fromCtyNumberBig(bf *big.Float, target reflect.Value, path cty.Path) error {
	switch {

	case bigFloatType.AssignableTo(target.Type()):
		// Easy!
		target.Set(reflect.ValueOf(bf).Elem())
		return nil

	case bigIntType.AssignableTo(target.Type()):
		bi, accuracy := bf.Int(nil)
		if accuracy != big.Exact {
			return errorf(path, "value must be a whole number")
		}
		target.Set(reflect.ValueOf(bi).Elem())
		return nil

	default:
		return likelyRequiredTypesError(path, target)
	}
}

func fromCtyString(val cty.Value, target reflect.Value, path cty.Path) error {
	switch target.Kind() {

	case reflect.String:
		target.Set(reflect.ValueOf(val.AsString()))
		return nil

	default:
		return likelyRequiredTypesError(path, target)

	}
}

func fromCtyList(val cty.Value, target reflect.Value, path cty.Path) error {
	switch target.Kind() {

	case reflect.Slice:
		if val.IsNull() {
			target.Set(reflect.Zero(target.Type()))
			return nil
		}

		length := val.LengthInt()
		tv := reflect.MakeSlice(target.Type(), length, length)

		path = append(path, nil)

		i := 0
		var err error
		val.ForEachElement(func(key cty.Value, val cty.Value) bool {
			path[len(path)-1] = &cty.IndexStep{
				Key: cty.NumberIntVal(int64(i)),
			}

			targetElem := tv.Index(i)
			err = fromCtyValue(val, targetElem, path)
			if err != nil {
				return true
			}

			i++
			return false
		})
		if err != nil {
			return err
		}

		path = path[:len(path)-1]

		target.Set(tv)
		return nil

	case reflect.Array:
		if val.IsNull() {
			return errorf(path, "null value is not allowed")
		}

		length := val.LengthInt()
		if length != target.Len() {
			return errorf(path, "must be a list of length %d", target.Len())
		}

		path = append(path, nil)

		i := 0
		var err error
		val.ForEachElement(func(key cty.Value, val cty.Value) bool {
			path[len(path)-1] = &cty.IndexStep{
				Key: cty.NumberIntVal(int64(i)),
			}

			targetElem := target.Index(i)
			err = fromCtyValue(val, targetElem, path)
			if err != nil {
				return true
			}

			i++
			return false
		})
		if err != nil {
			return err
		}

		path = path[:len(path)-1]

		return nil

	default:
		return likelyRequiredTypesError(path, target)

	}
}

func fromCtyMap(val cty.Value, target reflect.Value, path cty.Path) error {

	switch target.Kind() {

	case reflect.Map:
		if val.IsNull() {
			target.Set(reflect.Zero(target.Type()))
			return nil
		}

		tv := reflect.MakeMap(target.Type())
		et := target.Type().Elem()

		path = append(path, nil)

		var err error
		val.ForEachElement(func(key cty.Value, val cty.Value) bool {
			path[len(path)-1] = &cty.IndexStep{
				Key: key,
			}

			ks := key.AsString()

			targetElem := reflect.New(et)
			err = fromCtyValue(val, targetElem, path)

			tv.SetMapIndex(reflect.ValueOf(ks), targetElem.Elem())

			return err != nil
		})
		if err != nil {
			return err
		}

		path = path[:len(path)-1]

		target.Set(tv)
		return nil

	default:
		return likelyRequiredTypesError(path, target)

	}
}

func fromCtySet(val cty.Value, target reflect.Value, path cty.Path) error {
	switch target.Kind() {

	case reflect.Slice:
		if val.IsNull() {
			target.Set(reflect.Zero(target.Type()))
			return nil
		}

		length := val.LengthInt()
		tv := reflect.MakeSlice(target.Type(), length, length)

		i := 0
		var err error
		val.ForEachElement(func(key cty.Value, val cty.Value) bool {
			targetElem := tv.Index(i)
			err = fromCtyValue(val, targetElem, path)
			if err != nil {
				return true
			}

			i++
			return false
		})
		if err != nil {
			return err
		}

		target.Set(tv)
		return nil

	case reflect.Array:
		if val.IsNull() {
			return errorf(path, "null value is not allowed")
		}

		length := val.LengthInt()
		if length != target.Len() {
			return errorf(path, "must be a set of length %d", target.Len())
		}

		i := 0
		var err error
		val.ForEachElement(func(key cty.Value, val cty.Value) bool {
			targetElem := target.Index(i)
			err = fromCtyValue(val, targetElem, path)
			if err != nil {
				return true
			}

			i++
			return false
		})
		if err != nil {
			return err
		}

		return nil

	// TODO: decode into set.Set instance

	default:
		return likelyRequiredTypesError(path, target)

	}
}

func fromCtyObject(val cty.Value, target reflect.Value, path cty.Path) error {

	switch target.Kind() {

	case reflect.Struct:

		attrTypes := val.Type().AttributeTypes()
		targetFields := structTagIndices(target.Type())

		path = append(path, nil)

		for k, i := range targetFields {
			if _, exists := attrTypes[k]; !exists {
				// If the field in question isn't able to represent nil,
				// that's an error.
				fk := target.Field(i).Kind()
				switch fk {
				case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface:
					// okay
				default:
					return errorf(path, "missing required attribute %q", k)
				}
			}
		}

		for k := range attrTypes {
			path[len(path)-1] = &cty.GetAttrStep{
				Name: k,
			}

			fieldIdx, exists := targetFields[k]
			if !exists {
				return errorf(path, "unsupported attribute %q", k)
			}

			ev := val.GetAttr(k)

			targetField := target.Field(fieldIdx)
			err := fromCtyValue(ev, targetField, path)
			if err != nil {
				return err
			}
		}

		path = path[:len(path)-1]

		return nil

	default:
		return likelyRequiredTypesError(path, target)

	}
}

func fromCtyDynamic(val cty.Value, target reflect.Value, path cty.Path) error {
	// TODO: implement this
	panic("decode into interface{} not yet supported")
}

// fromCtyPopulatePtr recognizes when target is a pointer type and allocates
// a value to assign to that pointer, which it returns.
//
// If the given value has multiple levels of indirection, like **int, these
// will be processed in turn so that the return value is guaranteed to be
// a non-pointer.
//
// As an exception, if decodingNull is true then the returned value will be
// the final level of pointer, if any, so that the caller can assign it
// as nil to represent a null value. If the given target value is not a pointer
// at all then the returned value will be just the given target, so the caller
// must test if the returned value is a pointer before trying to assign nil
// to it.
func fromCtyPopulatePtr(target reflect.Value, decodingNull bool) reflect.Value {
	for {
		if target.Kind() == reflect.Interface && !target.IsNil() {
			e := target.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() && (!decodingNull || e.Elem().Kind() == reflect.Ptr) {
				target = e
			}
		}

		if target.Kind() != reflect.Ptr {
			break
		}

		// Stop early if we're decodingNull and we've found our last indirection
		if target.Elem().Kind() != reflect.Ptr && decodingNull && target.CanSet() {
			break
		}

		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}

		target = target.Elem()
	}
	return target
}

// likelyRequiredTypesError returns an error that states which types are
// acceptable by making some assumptions about what types we support for
// each target Go kind. It's not a precise science but it allows us to return
// an error message that is cty-user-oriented rather than Go-oriented.
//
// Generally these error messages should be a matter of last resort, since
// the calling application should be validating user-provided value types
// before decoding anyway.
func likelyRequiredTypesError(path cty.Path, target reflect.Value) error {
	switch target.Kind() {

	case reflect.Bool:
		return errorf(path, "bool value is required")

	case reflect.String:
		return errorf(path, "string value is required")

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fallthrough
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fallthrough
	case reflect.Float32, reflect.Float64:
		return errorf(path, "number value is required")

	case reflect.Slice, reflect.Array:
		return errorf(path, "list or set value is required")

	case reflect.Map:
		return errorf(path, "map or object value is required")

	case reflect.Struct:
		switch {

		case target.Type().AssignableTo(bigFloatType) || target.Type().AssignableTo(bigIntType):
			return errorf(path, "number value is required")

		case target.Type().AssignableTo(setType):
			return errorf(path, "set or list value is required")

		default:
			return errorf(path, "object value is required")

		}

	default:
		// We should avoid getting into this path, since this error
		// message is rather useless.
		return errorf(path, "incorrect type")

	}
}
