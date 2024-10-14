package convert

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/zclconf/go-cty/cty"
)

// MismatchMessage is a helper to return an English-language description of
// the differences between got and want, phrased as a reason why got does
// not conform to want.
//
// This function does not itself attempt conversion, and so it should generally
// be used only after a conversion has failed, to report the conversion failure
// to an English-speaking user. The result will be confusing got is actually
// conforming to or convertable to want.
//
// The shorthand helper function Convert uses this function internally to
// produce its error messages, so callers of that function do not need to
// also use MismatchMessage.
//
// This function is similar to Type.TestConformance, but it is tailored to
// describing conversion failures and so the messages it generates relate
// specifically to the conversion rules implemented in this package.
func MismatchMessage(got, want cty.Type) string {
	switch {

	case got.IsObjectType() && want.IsObjectType():
		// If both types are object types then we may be able to say something
		// about their respective attributes.
		return mismatchMessageObjects(got, want)

	case got.IsUnionType() && want.IsUnionType():
		// If both types are union types then we can hopefully say something
		// about their respective variants.
		return mismatchMessageUnions(got, want)

	case got.IsObjectType() && want.IsUnionType():
		// If we were attempting an object-to-union conversion then we
		// can hopefully say something useful about why it failed.
		return mismatchMessageObjectToUnion(got, want)

	case got.IsTupleType() && want.IsListType() && want.ElementType() == cty.DynamicPseudoType:
		// If conversion from tuple to list failed then it's because we couldn't
		// find a common type to convert all of the tuple elements to.
		return "all list elements must have the same type"

	case got.IsTupleType() && want.IsSetType() && want.ElementType() == cty.DynamicPseudoType:
		// If conversion from tuple to set failed then it's because we couldn't
		// find a common type to convert all of the tuple elements to.
		return "all set elements must have the same type"

	case got.IsObjectType() && want.IsMapType() && want.ElementType() == cty.DynamicPseudoType:
		// If conversion from object to map failed then it's because we couldn't
		// find a common type to convert all of the object attributes to.
		return "all map elements must have the same type"

	case (got.IsTupleType() || got.IsObjectType()) && want.IsCollectionType():
		return mismatchMessageCollectionsFromStructural(got, want)

	case got.IsCollectionType() && want.IsCollectionType():
		return mismatchMessageCollectionsFromCollections(got, want)

	default:
		// If we have nothing better to say, we'll just state what was required.
		return want.FriendlyNameForConstraint() + " required"
	}
}

func mismatchMessageObjects(got, want cty.Type) string {
	// Per our conversion rules, "got" is allowed to be a superset of "want",
	// and so we'll produce error messages here under that assumption.
	gotAtys := got.AttributeTypes()
	wantAtys := want.AttributeTypes()

	// If we find missing attributes then we'll report those in preference,
	// but if not then we will report a maximum of one non-conforming
	// attribute, just to keep our messages relatively terse.
	// We'll also prefer to report a recursive type error from an _unsafe_
	// conversion over a safe one, because these are subjectively more
	// "serious".
	var missingAttrs []string
	var unsafeMismatchAttr string
	var safeMismatchAttr string

	for name, wantAty := range wantAtys {
		gotAty, exists := gotAtys[name]
		if !exists {
			if !want.AttributeOptional(name) {
				missingAttrs = append(missingAttrs, name)
			}
			continue
		}

		if gotAty.Equals(wantAty) {
			continue // exact match, so no problem
		}

		// We'll now try to convert these attributes in isolation and
		// see if we have a nested conversion error to report.
		// We'll try an unsafe conversion first, and then fall back on
		// safe if unsafe is possible.

		// If we already have an unsafe mismatch attr error then we won't bother
		// hunting for another one.
		if unsafeMismatchAttr != "" {
			continue
		}
		if conv := GetConversionUnsafe(gotAty, wantAty); conv == nil {
			unsafeMismatchAttr = fmt.Sprintf("attribute %q: %s", name, MismatchMessage(gotAty, wantAty))
		}

		// If we already have a safe mismatch attr error then we won't bother
		// hunting for another one.
		if safeMismatchAttr != "" {
			continue
		}
		if conv := GetConversion(gotAty, wantAty); conv == nil {
			safeMismatchAttr = fmt.Sprintf("attribute %q: %s", name, MismatchMessage(gotAty, wantAty))
		}
	}

	// We should now have collected at least one problem. If we have more than
	// one then we'll use our preference order to decide what is most important
	// to report.
	switch {

	case len(missingAttrs) != 0:
		sort.Strings(missingAttrs)
		switch len(missingAttrs) {
		case 1:
			return fmt.Sprintf("attribute %q is required", missingAttrs[0])
		case 2:
			return fmt.Sprintf("attributes %q and %q are required", missingAttrs[0], missingAttrs[1])
		default:
			sort.Strings(missingAttrs)
			var buf bytes.Buffer
			for _, name := range missingAttrs[:len(missingAttrs)-1] {
				fmt.Fprintf(&buf, "%q, ", name)
			}
			fmt.Fprintf(&buf, "and %q", missingAttrs[len(missingAttrs)-1])
			return fmt.Sprintf("attributes %s are required", buf.Bytes())
		}

	case unsafeMismatchAttr != "":
		return unsafeMismatchAttr

	case safeMismatchAttr != "":
		return safeMismatchAttr

	default:
		// We should never get here, but if we do then we'll return
		// just a generic message.
		return "incorrect object attributes"
	}
}

func mismatchMessageUnions(got, want cty.Type) string {
	var missingVariants []string
	var incompatibleVariantsSafe []string
	var incompatibleVariantsUnsafe []string

	givenVtys := got.UnionVariants()
	allowedVtys := got.UnionVariants()
	for name, givenVty := range givenVtys {
		allowedVty, ok := allowedVtys[name]
		if !ok {
			missingVariants = append(missingVariants, name)
			continue
		}
		if givenVty.Equals(allowedVty) {
			continue // definitely okay
		}
		if conv := GetConversion(givenVty, allowedVty); conv == nil {
			incompatibleVariantsSafe = append(incompatibleVariantsSafe, fmt.Sprintf(
				"%s: %s",
				name, MismatchMessage(givenVty, allowedVty),
			))
			continue
		}
		if conv := GetConversionUnsafe(givenVty, allowedVty); conv == nil {
			incompatibleVariantsUnsafe = append(incompatibleVariantsUnsafe, fmt.Sprintf(
				"%s: %s",
				name, MismatchMessage(givenVty, allowedVty),
			))
		}
	}

	// We return a message about the most severe of the violations, to
	// avoid returning a horrendously-long error message.
	switch {
	case len(missingVariants) != 0:
		sort.Strings(missingVariants)
		switch len(missingVariants) {
		case 1:
			return fmt.Sprintf("unexpected union variant %q", missingVariants[0])
		default:
			return fmt.Sprintf("unexpected union variants: %s", strings.Join(missingVariants, ", "))
		}
	case len(incompatibleVariantsSafe) != 0:
		sort.Strings(incompatibleVariantsSafe)
		// To keep the message relatively concise, we'll return just the first
		// incompatibility.
		return "incompatible type for variant " + incompatibleVariantsSafe[0]
	case len(incompatibleVariantsUnsafe) != 0:
		sort.Strings(incompatibleVariantsUnsafe)
		// To keep the message relatively concise, we'll return just the first
		// incompatibility.
		return "incompatible type for variant " + incompatibleVariantsUnsafe[0]
	default:
		// Generic (but rather useless) error message just in case something
		// else is wrong that we didn't handle properly.
		return "incompatible union type"
	}
}

func mismatchMessageObjectToUnion(got, want cty.Type) string {
	if len(got.AttributeTypes()) == 0 {
		return "empty object cannot convert to union type"
	}

	var unwantedAttrs []string
	var incompatibleAttrs []string
	atys := got.AttributeTypes()
	vtys := want.UnionVariants()
	for name, aty := range atys {
		vty, ok := vtys[name]
		if !ok {
			unwantedAttrs = append(unwantedAttrs, name)
			continue
		}
		if !aty.Equals(vty) {
			if conv := GetConversionUnsafe(aty, vty); conv == nil {
				mismatch := MismatchMessage(aty, vty)
				incompatibleAttrs = append(incompatibleAttrs, fmt.Sprintf("%q: %s", name, mismatch))
			}
		}
	}
	switch {
	case len(unwantedAttrs) != 0:
		sort.Strings(unwantedAttrs)
		switch len(unwantedAttrs) {
		case 1:
			return fmt.Sprintf("tuple type has no variant named %q", unwantedAttrs[0])
		case 2:
			return fmt.Sprintf("tuple type has no variants named %q or %q", unwantedAttrs[0], unwantedAttrs[1])
		default:
			return fmt.Sprintf("tuple type does not have any of the following variant names: %s", strings.Join(unwantedAttrs, ","))
		}
	case len(incompatibleAttrs) != 0:
		sort.Strings(incompatibleAttrs)
		// There isn't really any compact way for us to report more than one
		// incompatible error message, so we'll just report whichever one
		// sorts lexically first.
		return "incompatible type for tuple variant " + incompatibleAttrs[0]
	default:
		// If all else fails, a generic (and rather useless) error message
		return "incompatible object type for union conversion"
	}
}

func mismatchMessageCollectionsFromStructural(got, want cty.Type) string {
	// First some straightforward cases where the kind is just altogether wrong.
	switch {
	case want.IsListType() && !got.IsTupleType():
		return want.FriendlyNameForConstraint() + " required"
	case want.IsSetType() && !got.IsTupleType():
		return want.FriendlyNameForConstraint() + " required"
	case want.IsMapType() && !got.IsObjectType():
		return want.FriendlyNameForConstraint() + " required"
	}

	// If the kinds are matched well enough then we'll move on to checking
	// individual elements.
	wantEty := want.ElementType()
	switch {
	case got.IsTupleType():
		for i, gotEty := range got.TupleElementTypes() {
			if gotEty.Equals(wantEty) {
				continue // exact match, so no problem
			}
			if conv := getConversion(gotEty, wantEty, true); conv != nil {
				continue // conversion is available, so no problem
			}
			return fmt.Sprintf("element %d: %s", i, MismatchMessage(gotEty, wantEty))
		}

		// If we get down here then something weird is going on but we'll
		// return a reasonable fallback message anyway.
		return fmt.Sprintf("all elements must be %s", wantEty.FriendlyNameForConstraint())

	case got.IsObjectType():
		for name, gotAty := range got.AttributeTypes() {
			if gotAty.Equals(wantEty) {
				continue // exact match, so no problem
			}
			if conv := getConversion(gotAty, wantEty, true); conv != nil {
				continue // conversion is available, so no problem
			}
			return fmt.Sprintf("element %q: %s", name, MismatchMessage(gotAty, wantEty))
		}

		// If we get down here then something weird is going on but we'll
		// return a reasonable fallback message anyway.
		return fmt.Sprintf("all elements must be %s", wantEty.FriendlyNameForConstraint())

	default:
		// Should not be possible to get here since we only call this function
		// with got as structural types, but...
		return want.FriendlyNameForConstraint() + " required"
	}
}

func mismatchMessageCollectionsFromCollections(got, want cty.Type) string {
	// First some straightforward cases where the kind is just altogether wrong.
	switch {
	case want.IsListType() && !(got.IsListType() || got.IsSetType()):
		return want.FriendlyNameForConstraint() + " required"
	case want.IsSetType() && !(got.IsListType() || got.IsSetType()):
		return want.FriendlyNameForConstraint() + " required"
	case want.IsMapType() && !got.IsMapType():
		return want.FriendlyNameForConstraint() + " required"
	}

	// If the kinds are matched well enough then we'll check the element types.
	gotEty := got.ElementType()
	wantEty := want.ElementType()
	noun := "element type"
	switch {
	case want.IsListType():
		noun = "list element type"
	case want.IsSetType():
		noun = "set element type"
	case want.IsMapType():
		noun = "map element type"
	}
	return fmt.Sprintf("incorrect %s: %s", noun, MismatchMessage(gotEty, wantEty))
}
