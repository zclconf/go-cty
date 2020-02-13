package convert

import (
	"github.com/zclconf/go-cty/cty"
)

// conversionObjectToObject returns a conversion that will make the input
// object type conform to the output object type, if possible.
//
// Conversion is possible only if the output type is a subset of the input
// type, meaning that each attribute of the output type has a corresponding
// attribute in the input type where a recursive conversion is available.
//
// Shallow object conversions work the same for both safe and unsafe modes,
// but the safety flag is passed on to recursive conversions and may thus
// limit the above definition of "subset".
func conversionObjectToObject(in cty.Type, out cty.Type, unsafe bool) conversion {
	inAtys := in.AttributeTypes()
	outAtys := out.AttributeTypes()

	return conversionObjectOrMapToObject(false, inAtys, outAtys, unsafe)
}

// conversionMapToObject works the same as conversionObjectToObject, however
// at this point we don’t know if the input map has all the keys that the
// output object requires. As such it is always unsafe.
func conversionMapToObject(mapType cty.Type, objectType cty.Type, unsafe bool) conversion {
	outAtys := objectType.AttributeTypes()
	atyCount := len(outAtys)
	inAtys := make(map[string]cty.Type, atyCount)

	if atyCount != 0 {
		mapEty := mapType.ElementType()

		// For now we just assume the map will have all the necessary keys so at
		// least we can fail early if we can’t convert from the map's element
		// type to the output object's values.
		for name := range outAtys {
			inAtys[name] = mapEty
		}
	}

	return conversionObjectOrMapToObject(true, inAtys, outAtys, unsafe)
}

// conversionObjectOrMapToObject first checks if the values in the input can
// be converted to the output object's, and returns a conversion if successful.
func conversionObjectOrMapToObject(inIsMap bool, inAtys map[string]cty.Type, outAtys map[string]cty.Type, unsafe bool) conversion {
	attrConvs := make(map[string]conversion)

	for name, outAty := range outAtys {
		inAty, exists := inAtys[name]
		if !exists {
			// No conversion is available, then.
			return nil
		}

		if inAty.Equals(outAty) {
			// No conversion needed, but we'll still record the attribute
			// in our map for later reference.
			attrConvs[name] = nil
			continue
		}

		attrConvs[name] = getConversion(inAty, outAty, unsafe)
		if attrConvs[name] == nil {
			// If a recursive conversion isn't available, then our top-level
			// configuration is impossible too.
			return nil
		}
	}

	// If we get here then a conversion is possible, using the attribute
	// conversions given in attrConvs.
	return func(val cty.Value, path cty.Path) (cty.Value, error) {
		path = append(path, nil)
		pathStep := &path[len(path)-1]

		if len(attrConvs) == 0 {
			return cty.EmptyObjectVal, nil
		}

		if inIsMap {
			// This is the first opportunity to check if the input map has all
			// of the keys wanted in the output object.
			for name := range outAtys {
				if val.HasIndex(cty.StringVal(name)).False() {
					return cty.NilVal, path.NewErrorf("missing required attribute %s", name)
				}
			}
		}

		attrVals := make(map[string]cty.Value, len(attrConvs))

		for it := val.ElementIterator(); it.Next(); {
			nameVal, val := it.Element()
			var err error

			name := nameVal.AsString()
			*pathStep = cty.GetAttrStep{
				Name: name,
			}

			conv, exists := attrConvs[name]
			if !exists {
				continue
			}
			if conv != nil {
				val, err = conv(val, path)
				if err != nil {
					return cty.NilVal, err
				}
			}

			attrVals[name] = val
		}

		return cty.ObjectVal(attrVals), nil
	}
}
