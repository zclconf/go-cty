package json

import (
	"bytes"
	"encoding/json"

	"github.com/apparentlymart/go-cty/cty"
)

func marshal(val cty.Value, t cty.Type, path cty.Path, b *bytes.Buffer) error {
	// If we're going to decode as DynamicPseudoType then we need to save
	// dynamic type information to recover the real type.
	if t == cty.DynamicPseudoType && val.Type() != cty.DynamicPseudoType {
		return marshalDynamic(val, path, b)
	}

	if val.IsNull() {
		b.WriteString("null")
		return nil
	}

	if !val.IsKnown() {
		return path.NewErrorf("value is not known")
	}

	// The caller should've guaranteed that the given val is conformant with
	// the given type t, so we'll proceed under that assumption here.

	switch {
	case t.IsPrimitiveType():
		switch t {
		case cty.String:
			json, err := json.Marshal(val.AsString())
			if err != nil {
				return path.NewErrorf("failed to serialize value: %s", err)
			}
			b.Write(json)
			return nil
		case cty.Number:
			b.WriteString(val.AsBigFloat().Text('f', -1))
			return nil
		case cty.Bool:
			if val.True() {
				b.WriteString("true")
			} else {
				b.WriteString("false")
			}
			return nil
		default:
			panic("unsupported primitive type")
		}
	default:
		panic("marshal not yet fully implemented")
	}
}

// marshalDynamic adds an extra wrapping object containing dynamic type
// information for the given value.
func marshalDynamic(val cty.Value, path cty.Path, b *bytes.Buffer) error {
	typeJSON, err := MarshalType(val.Type())
	if err != nil {
		return path.NewErrorf("failed to serialize type: %s", err)
	}
	b.WriteString(`{"value":`)
	marshal(val, val.Type(), path, b)
	b.WriteString(`,"type":`)
	b.Write(typeJSON)
	b.WriteRune('}')
	return nil
}
