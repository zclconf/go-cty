package convert

import (
	"sort"
	"strings"

	"github.com/zclconf/go-cty/cty"
)

func conversionUnionToUnion(in, out cty.Type, unsafe bool) conversion {
	// A union-to-union conversion is valid as long as every variant
	// in the source union is compatible with a variant in the
	// destination union. The destination union is allowed to have
	// additional variants, which are guaranteed to not be selected
	// if conversion succeeds.
	inVtys := in.UnionVariants()
	outVtys := out.UnionVariants()
	convs := make(map[string]conversion, len(inVtys))
	for name, inVty := range inVtys {
		outVty, ok := outVtys[name]
		if !ok {
			return nil
		}
		if !inVty.Equals(outVty) {
			conv := getConversion(inVty, outVty, unsafe)
			if conv == nil {
				return nil
			}
			convs[name] = conv
		} else {
			convs[name] = nil // no conversion required
		}
	}

	return func(v cty.Value, path cty.Path) (cty.Value, error) {
		variantName, variantVal := v.UnionVariant()
		conv := convs[variantName]
		if conv != nil {
			var err error
			path := path.GetAttr(variantName)
			variantVal, err = conv(variantVal, path)
			if err != nil {
				return cty.NilVal, err
			}
		}
		return cty.UnionVal(out, variantName, variantVal), nil
	}
}

func conversionObjectToUnion(in, out cty.Type, unsafe bool) conversion {
	// Conversion from object type to union type is never "safe" because
	// successful conversion requires that exactly one attribute of the
	// object type be set, and we can't guarantee that from type
	// information alone.
	if !unsafe {
		return nil
	}

	// Each of the object type's attributes must match with one of the
	// union type's variants.
	atys := in.AttributeTypes()
	if len(atys) == 0 {
		// An empty object cannot convert to any union type because it would
		// be unable to maintain the invariant that exactly one of its
		// attributes is non-null.
		return nil
	}

	vtys := out.UnionVariants()
	aconvs := make(map[string]conversion, len(atys))
	for name, aty := range atys {
		vty, ok := vtys[name]
		if !ok {
			return nil
		}
		if !vty.Equals(aty) {
			conv := getConversion(aty, vty, unsafe)
			if conv == nil {
				return nil
			}
			aconvs[name] = conv
		} else {
			aconvs[name] = nil // attribute is relevant but no conversion required
		}
	}

	// If we get here then we have a conversion (if needed) from each attribute
	// of the object type to the corresponding union variant, so we have enough
	// information to attempt a conversion.
	return func(val cty.Value, path cty.Path) (cty.Value, error) {
		var variantName string
		var conv conversion
		var sourceVal cty.Value
		for attrName, attrConv := range aconvs {
			v := val.GetAttr(attrName)
			if v.IsNull() {
				continue
			}
			if sourceVal != cty.NilVal {
				// More than one attribute is non-null
				attrNames := make([]string, 0, len(aconvs))
				for name := range aconvs {
					if !val.GetAttr(name).IsNull() {
						attrNames = append(attrNames, name)
					}
				}
				sort.Strings(attrNames)
				return cty.NilVal, path.NewErrorf("only one of the following attributes may be set: %s", strings.Join(attrNames, ", "))
			}
			variantName = attrName
			conv = attrConv
			sourceVal = v
		}
		if sourceVal == cty.NilVal {
			// No relevant attributes are non-null
			attrNames := make([]string, 0, len(aconvs))
			for name := range aconvs {
				attrNames = append(attrNames, name)
			}
			sort.Strings(attrNames)
			return cty.NilVal, path.NewErrorf("exactly one of the following attributes must be set: %s", strings.Join(attrNames, ", "))
		}

		variantVal := sourceVal
		if conv != nil {
			var err error
			attrPath := path.GetAttr(variantName)
			variantVal, err = conv(sourceVal, attrPath)
			if err != nil {
				return cty.NilVal, err
			}
		}
		return cty.UnionVal(out, variantName, variantVal), nil
	}
}
