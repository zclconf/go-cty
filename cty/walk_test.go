package cty

import (
	"fmt"
	"testing"
)

func TestWalk(t *testing.T) {
	type Call struct {
		Path string
		Type string
	}

	val := ObjectVal(map[string]Value{
		"string":        StringVal("hello"),
		"number":        NumberIntVal(10),
		"bool":          True,
		"list":          ListVal([]Value{True}),
		"list_empty":    ListValEmpty(Bool),
		"set":           SetVal([]Value{True}),
		"set_empty":     ListValEmpty(Bool),
		"tuple":         TupleVal([]Value{True}),
		"tuple_empty":   EmptyTupleVal,
		"map":           MapVal(map[string]Value{"true": True}),
		"map_empty":     MapValEmpty(Bool),
		"object":        ObjectVal(map[string]Value{"true": True}),
		"object_empty":  EmptyObjectVal,
		"null":          NullVal(List(String)),
		"unknown":       UnknownVal(Map(Bool)),
		"marked_string": StringVal("boop").Mark("blorp"),
		"marked_list":   ListVal([]Value{True}).Mark("blorp"),
		"marked_tuple":  TupleVal([]Value{True}).Mark("blorp"),
		"marked_set":    SetVal([]Value{True}).Mark("blorp"),
		"marked_object": ObjectVal(map[string]Value{"true": True}).Mark("blorp"),
		"marked_map":    MapVal(map[string]Value{"true": True}),
	})

	gotCalls := map[Call]struct{}{}
	wantCalls := []Call{
		{`cty.Path(nil)`, "object"},
		{`cty.Path{cty.GetAttrStep{Name:"string"}}`, "string"},
		{`cty.Path{cty.GetAttrStep{Name:"number"}}`, "number"},
		{`cty.Path{cty.GetAttrStep{Name:"bool"}}`, "bool"},
		{`cty.Path{cty.GetAttrStep{Name:"list"}}`, "list of bool"},
		{`cty.Path{cty.GetAttrStep{Name:"list"}, cty.IndexStep{Key:cty.NumberIntVal(0)}}`, "bool"},
		{`cty.Path{cty.GetAttrStep{Name:"list_empty"}}`, "list of bool"},
		{`cty.Path{cty.GetAttrStep{Name:"set"}}`, "set of bool"},
		{`cty.Path{cty.GetAttrStep{Name:"set"}, cty.IndexStep{Key:cty.True}}`, "bool"},
		{`cty.Path{cty.GetAttrStep{Name:"set_empty"}}`, "list of bool"},
		{`cty.Path{cty.GetAttrStep{Name:"tuple"}}`, "tuple"},
		{`cty.Path{cty.GetAttrStep{Name:"tuple"}, cty.IndexStep{Key:cty.NumberIntVal(0)}}`, "bool"},
		{`cty.Path{cty.GetAttrStep{Name:"tuple_empty"}}`, "tuple"},
		{`cty.Path{cty.GetAttrStep{Name:"map"}, cty.IndexStep{Key:cty.StringVal("true")}}`, "bool"},
		{`cty.Path{cty.GetAttrStep{Name:"map"}}`, "map of bool"},
		{`cty.Path{cty.GetAttrStep{Name:"map_empty"}}`, "map of bool"},
		{`cty.Path{cty.GetAttrStep{Name:"object"}}`, "object"},
		{`cty.Path{cty.GetAttrStep{Name:"object"}, cty.GetAttrStep{Name:"true"}}`, "bool"},
		{`cty.Path{cty.GetAttrStep{Name:"object_empty"}}`, "object"},
		{`cty.Path{cty.GetAttrStep{Name:"null"}}`, "list of string"},
		{`cty.Path{cty.GetAttrStep{Name:"unknown"}}`, "map of bool"},
		{`cty.Path{cty.GetAttrStep{Name:"marked_string"}}`, "string"},
		{`cty.Path{cty.GetAttrStep{Name:"marked_list"}}`, "list of bool"},
		{`cty.Path{cty.GetAttrStep{Name:"marked_list"}, cty.IndexStep{Key:cty.NumberIntVal(0)}}`, "bool"},
		{`cty.Path{cty.GetAttrStep{Name:"marked_set"}}`, "set of bool"},
		{`cty.Path{cty.GetAttrStep{Name:"marked_set"}, cty.IndexStep{Key:cty.True}}`, "bool"},
		{`cty.Path{cty.GetAttrStep{Name:"marked_object"}}`, "object"},
		{`cty.Path{cty.GetAttrStep{Name:"marked_object"}, cty.GetAttrStep{Name:"true"}}`, "bool"},
		{`cty.Path{cty.GetAttrStep{Name:"marked_tuple"}}`, "tuple"},
		{`cty.Path{cty.GetAttrStep{Name:"marked_tuple"}, cty.IndexStep{Key:cty.NumberIntVal(0)}}`, "bool"},
		{`cty.Path{cty.GetAttrStep{Name:"marked_map"}}`, "map of bool"},
		{`cty.Path{cty.GetAttrStep{Name:"marked_map"}, cty.IndexStep{Key:cty.StringVal("true")}}`, "bool"},
	}

	err := Walk(val, func(path Path, val Value) (bool, error) {
		gotCalls[Call{
			Path: fmt.Sprintf("%#v", path),
			Type: val.Type().FriendlyName(),
		}] = struct{}{}
		return true, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(gotCalls) != len(wantCalls) {
		t.Errorf("wrong number of calls %d; want %d", len(gotCalls), len(wantCalls))
	}

	for gotCall := range gotCalls {
		t.Logf("got call {%#q, %q}", gotCall.Path, gotCall.Type)
	}

	for _, wantCall := range wantCalls {
		if _, has := gotCalls[wantCall]; !has {
			t.Errorf("missing call {%#q, %q}", wantCall.Path, wantCall.Type)
		}
	}
}

type pathTransformer struct{}

func (pathTransformer) Enter(p Path, v Value) (Value, error) {
	return v, nil
}

func (pathTransformer) Exit(p Path, v Value) (Value, error) {
	if v.Type().IsPrimitiveType() {
		return StringVal(fmt.Sprintf("%#v", p)), nil
	}
	return v, nil
}

func TestTransformWithTransformer(t *testing.T) {
	val := ObjectVal(map[string]Value{
		"string":        StringVal("hello"),
		"number":        NumberIntVal(10),
		"bool":          True,
		"list":          ListVal([]Value{True}),
		"list_empty":    ListValEmpty(Bool),
		"set":           SetVal([]Value{True}),
		"set_empty":     ListValEmpty(Bool),
		"tuple":         TupleVal([]Value{True}),
		"tuple_empty":   EmptyTupleVal,
		"map":           MapVal(map[string]Value{"true": True}),
		"map_empty":     MapValEmpty(Bool),
		"object":        ObjectVal(map[string]Value{"true": True}),
		"object_empty":  EmptyObjectVal,
		"null":          NullVal(String),
		"unknown":       UnknownVal(Bool),
		"null_list":     NullVal(List(String)),
		"unknown_map":   UnknownVal(Map(Bool)),
		"marked_string": StringVal("hello").Mark("blorp"),
		"marked_list":   ListVal([]Value{True}).Mark("blorp"),
		"marked_set":    SetVal([]Value{True}).Mark("blorp"),
		"marked_tuple":  TupleVal([]Value{True}).Mark("blorp"),
		"marked_map":    MapVal(map[string]Value{"true": True}).Mark("blorp"),
		"marked_object": ObjectVal(map[string]Value{"true": True}).Mark("blorp"),
	})

	gotVal, err := TransformWithTransformer(val, pathTransformer{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	wantVal := ObjectVal(map[string]Value{
		"string":        StringVal(`cty.Path{cty.GetAttrStep{Name:"string"}}`),
		"number":        StringVal(`cty.Path{cty.GetAttrStep{Name:"number"}}`),
		"bool":          StringVal(`cty.Path{cty.GetAttrStep{Name:"bool"}}`),
		"list":          ListVal([]Value{StringVal(`cty.Path{cty.GetAttrStep{Name:"list"}, cty.IndexStep{Key:cty.NumberIntVal(0)}}`)}),
		"list_empty":    ListValEmpty(Bool),
		"set":           SetVal([]Value{StringVal(`cty.Path{cty.GetAttrStep{Name:"set"}, cty.IndexStep{Key:cty.True}}`)}),
		"set_empty":     ListValEmpty(Bool),
		"tuple":         TupleVal([]Value{StringVal(`cty.Path{cty.GetAttrStep{Name:"tuple"}, cty.IndexStep{Key:cty.NumberIntVal(0)}}`)}),
		"tuple_empty":   EmptyTupleVal,
		"map":           MapVal(map[string]Value{"true": StringVal(`cty.Path{cty.GetAttrStep{Name:"map"}, cty.IndexStep{Key:cty.StringVal("true")}}`)}),
		"map_empty":     MapValEmpty(Bool),
		"object":        ObjectVal(map[string]Value{"true": StringVal(`cty.Path{cty.GetAttrStep{Name:"object"}, cty.GetAttrStep{Name:"true"}}`)}),
		"object_empty":  EmptyObjectVal,
		"null":          StringVal(`cty.Path{cty.GetAttrStep{Name:"null"}}`),
		"unknown":       StringVal(`cty.Path{cty.GetAttrStep{Name:"unknown"}}`),
		"null_list":     NullVal(List(String)),
		"unknown_map":   UnknownVal(Map(Bool)),
		"marked_string": StringVal(`cty.Path{cty.GetAttrStep{Name:"marked_string"}}`),
		"marked_list":   ListVal([]Value{StringVal(`cty.Path{cty.GetAttrStep{Name:"marked_list"}, cty.IndexStep{Key:cty.NumberIntVal(0)}}`)}).Mark("blorp"),
		"marked_set":    SetVal([]Value{StringVal(`cty.Path{cty.GetAttrStep{Name:"marked_set"}, cty.IndexStep{Key:cty.True}}`)}).Mark("blorp"),
		"marked_tuple":  TupleVal([]Value{StringVal(`cty.Path{cty.GetAttrStep{Name:"marked_tuple"}, cty.IndexStep{Key:cty.NumberIntVal(0)}}`)}).Mark("blorp"),
		"marked_map":    MapVal(map[string]Value{"true": StringVal(`cty.Path{cty.GetAttrStep{Name:"marked_map"}, cty.IndexStep{Key:cty.StringVal("true")}}`)}).Mark("blorp"),
		"marked_object": ObjectVal(map[string]Value{"true": StringVal(`cty.Path{cty.GetAttrStep{Name:"marked_object"}, cty.GetAttrStep{Name:"true"}}`)}).Mark("blorp"),
	})

	if !gotVal.RawEquals(wantVal) {
		t.Errorf("wrong result\ngot:  %#v\nwant: %#v", gotVal, wantVal)
		if got, want := len(gotVal.Type().AttributeTypes()), len(gotVal.Type().AttributeTypes()); got != want {
			t.Errorf("wrong length %d; want %d", got, want)
		}
		for it := wantVal.ElementIterator(); it.Next(); {
			key, wantElem := it.Element()
			attr := key.AsString()
			if !gotVal.Type().HasAttribute(attr) {
				t.Errorf("missing attribute %q", attr)
				continue
			}
			gotElem := gotVal.GetAttr(attr)
			if !gotElem.RawEquals(wantElem) {
				t.Errorf("wrong value for attribute %q\ngot:  %#v\nwant: %#v", attr, gotElem, wantElem)
			}
		}
	}
}

type errorTransformer struct{}

func (errorTransformer) Enter(p Path, v Value) (Value, error) {
	return v, nil
}

func (errorTransformer) Exit(p Path, v Value) (Value, error) {
	ty := v.Type()
	if ty.IsPrimitiveType() {
		return v, nil
	}
	return v, p.NewError(fmt.Errorf("expected primitive type, was %#v", ty))
}

func TestTransformWithTransformer_error(t *testing.T) {
	val := ObjectVal(map[string]Value{
		"string": StringVal("hello"),
		"number": NumberIntVal(10),
		"bool":   True,
		"list":   ListVal([]Value{True}),
	})

	gotVal, err := TransformWithTransformer(val, errorTransformer{})
	if gotVal != DynamicVal {
		t.Fatalf("expected DynamicVal, got %#v", gotVal)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	pathError, ok := err.(PathError)
	if !ok {
		t.Fatalf("expected PathError, got %#v", err)
	}

	if got, want := pathError.Path, GetAttrPath("list"); !got.Equals(want) {
		t.Errorf("wrong path\n got: %#v\nwant: %#v", got, want)
	}
}

func TestTransform(t *testing.T) {
	val := ObjectVal(map[string]Value{
		"list": ListVal([]Value{True, True, False}),
		"set":  SetVal([]Value{True, False}),
		"map":  MapVal(map[string]Value{"a": True, "b": False}),
		"object": ObjectVal(map[string]Value{
			"a": True,
			"b": ListVal([]Value{False, False, False}),
		}),
	})
	wantVal := ObjectVal(map[string]Value{
		"list": ListVal([]Value{False, False, True}),
		"set":  SetVal([]Value{True, False}),
		"map":  MapVal(map[string]Value{"a": False, "b": True}),
		"object": ObjectVal(map[string]Value{
			"a": False,
			"b": ListVal([]Value{True, True, True}),
		}),
	})

	gotVal, err := Transform(val, func(p Path, v Value) (Value, error) {
		if v.Type().Equals(Bool) {
			return v.Not(), nil
		}
		return v, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !gotVal.RawEquals(wantVal) {
		t.Fatalf("wrong value\n got: %#v\nwant: %#v", gotVal, wantVal)
	}
}
