package json

import (
	"encoding/json"

	"github.com/apparentlymart/go-cty/cty"
	"github.com/apparentlymart/go-cty/cty/convert"
)

func unmarshal(dec *json.Decoder, t cty.Type, path cty.Path) (cty.Value, error) {
	tok, err := dec.Token()
	if err != nil {
		return cty.NilVal, path.NewError(err)
	}

	return unmarshalTok(tok, dec, t, path)
}

func unmarshalTok(tok json.Token, dec *json.Decoder, t cty.Type, path cty.Path) (cty.Value, error) {
	if tok == nil {
		return cty.NullVal(t), nil
	}

	if t == cty.DynamicPseudoType {
		return unmarshalDynamic(tok, dec, path)
	}

	switch {
	case t.IsPrimitiveType():
		return unmarshalPrimitive(tok, dec, t, path)
	case t.IsListType():
		return unmarshalList(tok, dec, t.ElementType(), path)
	case t.IsSetType():
		return unmarshalSet(tok, dec, t.ElementType(), path)
	case t.IsMapType():
		return unmarshalMap(tok, dec, t.ElementType(), path)
	case t.IsTupleType():
		return unmarshalTuple(tok, dec, t.TupleElementTypes(), path)
	case t.IsObjectType():
		return unmarshalObject(tok, dec, t.AttributeTypes(), path)
	default:
		return cty.NilVal, path.NewErrorf("unsupported type %s", t.FriendlyName())
	}
}

func unmarshalPrimitive(tok json.Token, dec *json.Decoder, t cty.Type, path cty.Path) (cty.Value, error) {

	switch t {
	case cty.Bool:
		switch v := tok.(type) {
		case bool:
			return cty.BoolVal(v), nil
		case string:
			val, err := convert.Convert(cty.StringVal(v), t)
			if err != nil {
				return cty.NilVal, path.NewError(err)
			}
			return val, nil
		default:
			return cty.NilVal, path.NewErrorf("bool is required")
		}
	case cty.Number:
		if v, ok := tok.(json.Number); ok {
			tok = string(v)
		}
		switch v := tok.(type) {
		case string:
			val, err := convert.Convert(cty.StringVal(v), t)
			if err != nil {
				return cty.NilVal, path.NewError(err)
			}
			return val, nil
		default:
			return cty.NilVal, path.NewErrorf("number is required")
		}
	case cty.String:
		switch v := tok.(type) {
		case string:
			return cty.StringVal(v), nil
		case json.Number:
			return cty.StringVal(string(v)), nil
		case bool:
			val, err := convert.Convert(cty.BoolVal(v), t)
			if err != nil {
				return cty.NilVal, path.NewError(err)
			}
			return val, nil
		default:
			return cty.NilVal, path.NewErrorf("string is required")
		}
	default:
		// should never happen
		panic("unsupported primitive type")
	}
}

func unmarshalList(tok json.Token, dec *json.Decoder, ety cty.Type, path cty.Path) (cty.Value, error) {
	if tok != json.Delim('[') {
		return cty.NilVal, path.NewErrorf("need JSON array for list")
	}

	var vals []cty.Value

	{
		path := append(path, nil)
		var idx int64

		for {
			path[len(path)-1] = cty.IndexStep{
				Key: cty.NumberIntVal(idx),
			}

			tok, err := dec.Token()
			if err != nil {
				return cty.NilVal, path.NewError(err)
			}

			if tok == json.Delim(']') {
				break
			}

			el, err := unmarshalTok(tok, dec, ety, path)
			if err != nil {
				return cty.NilVal, err
			}

			vals = append(vals, el)

			idx++
		}
	}

	if len(vals) == 0 {
		return cty.ListValEmpty(ety), nil
	}

	return cty.ListVal(vals), nil
}

func unmarshalSet(tok json.Token, dec *json.Decoder, ety cty.Type, path cty.Path) (cty.Value, error) {
	if tok != json.Delim('[') {
		return cty.NilVal, path.NewErrorf("need JSON array for set")
	}

	var vals []cty.Value

	{
		path := append(path, nil)
		var idx int64

		for {
			path[len(path)-1] = cty.IndexStep{
				Key: cty.UnknownVal(ety),
			}

			tok, err := dec.Token()
			if err != nil {
				return cty.NilVal, path.NewError(err)
			}

			if tok == json.Delim(']') {
				break
			}

			el, err := unmarshalTok(tok, dec, ety, path)
			if err != nil {
				return cty.NilVal, err
			}

			vals = append(vals, el)

			idx++
		}
	}

	if len(vals) == 0 {
		return cty.SetValEmpty(ety), nil
	}

	return cty.SetVal(vals), nil
}

func unmarshalMap(tok json.Token, dec *json.Decoder, ety cty.Type, path cty.Path) (cty.Value, error) {
	if tok != json.Delim('{') {
		return cty.NilVal, path.NewErrorf("need JSON object for map")
	}

	vals := make(map[string]cty.Value)

	{
		path := append(path, nil)

		for {
			path[len(path)-1] = cty.IndexStep{
				Key: cty.UnknownVal(cty.String),
			}

			tok, err := dec.Token()
			if err != nil {
				return cty.NilVal, path.NewError(err)
			}

			if tok == json.Delim('}') {
				break
			}

			k, ok := tok.(string)
			if !ok {
				return cty.NilVal, path.NewErrorf("invalid map key")
			}

			path[len(path)-1] = cty.IndexStep{
				Key: cty.StringVal(k),
			}

			el, err := unmarshal(dec, ety, path)
			if err != nil {
				return cty.NilVal, err
			}

			vals[k] = el
		}
	}

	if len(vals) == 0 {
		return cty.MapValEmpty(ety), nil
	}

	return cty.MapVal(vals), nil
}

func unmarshalTuple(tok json.Token, dec *json.Decoder, etys []cty.Type, path cty.Path) (cty.Value, error) {
	if tok != json.Delim('[') {
		return cty.NilVal, path.NewErrorf("need JSON array for tuple")
	}

	var vals []cty.Value

	{
		path := append(path, nil)
		var idx int

		for {
			path[len(path)-1] = cty.IndexStep{
				Key: cty.NumberIntVal(int64(idx)),
			}

			tok, err := dec.Token()
			if err != nil {
				return cty.NilVal, path.NewError(err)
			}

			if tok == json.Delim(']') {
				if len(vals) != len(etys) {
					return cty.NilVal, path[:len(path)-1].NewErrorf("not enough tuple elements (need %d)", len(etys))
				}
				break
			}

			if idx >= len(etys) {
				return cty.NilVal, path[:len(path)-1].NewErrorf("too many tuple elements (need %d)", len(etys))
			}
			ety := etys[idx]
			idx++

			el, err := unmarshalTok(tok, dec, ety, path)
			if err != nil {
				return cty.NilVal, err
			}

			vals = append(vals, el)
		}
	}

	if len(vals) == 0 {
		return cty.EmptyTupleVal, nil
	}

	return cty.TupleVal(vals), nil
}

func unmarshalObject(tok json.Token, dec *json.Decoder, atys map[string]cty.Type, path cty.Path) (cty.Value, error) {
	if tok != json.Delim('{') {
		return cty.NilVal, path.NewErrorf("need JSON object for object")
	}

	vals := make(map[string]cty.Value)

	{
		objPath := path           // some errors report from the object's perspective
		path := append(path, nil) // path to a specific attribute

		for {
			tok, err := dec.Token()
			if err != nil {
				return cty.NilVal, objPath.NewError(err)
			}

			if tok == json.Delim('}') {
				break
			}

			k, ok := tok.(string)
			if !ok {
				return cty.NilVal, path.NewErrorf("invalid object attribute")
			}

			path[len(path)-1] = cty.GetAttrStep{
				Name: k,
			}

			aty, ok := atys[k]
			if !ok {
				return cty.NilVal, objPath.NewErrorf("unsupported attribute %q", k)
			}

			el, err := unmarshal(dec, aty, path)
			if err != nil {
				return cty.NilVal, err
			}

			vals[k] = el
		}
	}

	// Make sure we have a value for every attribute
	for k, aty := range atys {
		if _, exists := vals[k]; !exists {
			vals[k] = cty.NullVal(aty)
		}
	}

	if len(vals) == 0 {
		return cty.EmptyObjectVal, nil
	}

	return cty.ObjectVal(vals), nil
}

func unmarshalDynamic(tok json.Token, dec *json.Decoder, path cty.Path) (cty.Value, error) {
	if tok != json.Delim('{') {
		return cty.NilVal, path.NewErrorf("need JSON object for dynamically-typed value")
	}

	var t cty.Type
	var valBody json.RawMessage // defer actual decoding until we know the type

	for {
		tok, err := dec.Token()
		if err != nil {
			return cty.NilVal, path.NewError(err)
		}

		if tok == json.Delim('}') {
			break
		}

		key, _ := tok.(string) // key == "" if tok is not a string

		switch key {
		case "type":
			err := dec.Decode(&t)
			if err != nil {
				return cty.NilVal, path.NewError(err)
			}
		case "value":
			err := dec.Decode(&valBody)
			if err != nil {
				return cty.NilVal, path.NewError(err)
			}
		default:
			return cty.NilVal, path.NewErrorf("invalid key %q in dynamically-typed value", key)
		}

	}

	if t == cty.NilType {
		return cty.NilVal, path.NewErrorf("missing type in dynamically-typed value")
	}
	if valBody == nil {
		return cty.NilVal, path.NewErrorf("missing value in dynamically-typed value")
	}

	val, err := Unmarshal([]byte(valBody), t)
	if err != nil {
		return cty.NilVal, path.NewError(err)
	}
	return val, nil
}
