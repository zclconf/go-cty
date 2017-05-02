package convert

import (
	"github.com/apparentlymart/go-cty/cty"
)

// conversionCollectionToList returns a conversion that will apply the given
// conversion to all of the elements of a collection (something that supports
// ForEachElement and LengthInt) and then returns the result as a list.
//
// "conv" can be nil if the elements are expected to already be of the
// correct type and just need to be re-wrapped into a list. (For example,
// if we're converting from a set into a list of the same element type.)
func conversionCollectionToList(ety cty.Type, conv conversion) conversion {
	return func(val cty.Value, path cty.Path) (cty.Value, error) {
		elems := make([]cty.Value, 0, val.LengthInt())
		i := int64(0)
		path = append(path, nil)
		var err error
		val.ForEachElement(func(key cty.Value, val cty.Value) bool {
			path[len(path)-1] = cty.IndexStep{
				Key: cty.NumberIntVal(i),
			}

			if conv != nil {
				val, err = conv(val, path)
				if err != nil {
					return true
				}
			}
			elems = append(elems, val)

			i++
			return false
		})
		if err != nil {
			return cty.NilVal, err
		}

		if len(elems) == 0 {
			return cty.ListValEmpty(ety), nil
		}

		return cty.ListVal(elems), nil
	}
}
