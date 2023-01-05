package walk

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"
)

func TestTransformValue_Identity(t *testing.T) {
	t.Parallel()

	// This is a kinda-contrived transformation that should just reproduce
	// whatever is fed into it exactly, which then allows us to make sure that
	// all possible input values are handled correctly and the transformer has
	// enough information to retain any characteristic of its input that it
	// might need to.

	capTy := cty.Capsule("capsule", reflect.TypeOf(""))
	capStr := "hello"

	tests := []cty.Value{
		cty.DynamicVal,
		cty.NullVal(cty.String),
		cty.UnknownVal(cty.Bool),
		cty.UnknownVal(cty.DynamicPseudoType),
		cty.StringVal("hello"),
		cty.True,
		cty.NumberIntVal(5),
		cty.TupleVal([]cty.Value{
			cty.StringVal("a"),
			cty.True,
		}),
		cty.ListVal([]cty.Value{
			cty.StringVal("a"),
			cty.StringVal("b"),
		}),
		cty.SetVal([]cty.Value{
			cty.StringVal("a"),
			cty.StringVal("b"),
		}),
		cty.ObjectVal(map[string]cty.Value{
			"string": cty.StringVal("a"),
			"bool":   cty.True,
		}),
		cty.MapVal(map[string]cty.Value{
			"a": cty.StringVal("a-val"),
			"b": cty.StringVal("b-val"),
		}),
		cty.CapsuleVal(capTy, &capStr),
		cty.True.Mark("oh no"),
		cty.ListVal([]cty.Value{
			cty.StringVal("a"),
			cty.StringVal("b"),
		}).Mark("oh no"),
		cty.ListVal([]cty.Value{
			cty.StringVal("a"),
			cty.StringVal("b").Mark("oh no"),
		}),
		cty.TupleVal([]cty.Value{
			cty.StringVal("a"),
			cty.StringVal("b"),
		}).Mark("oh no"),
		cty.TupleVal([]cty.Value{
			cty.StringVal("a"),
			cty.StringVal("b").Mark("oh no"),
		}),
		cty.SetVal([]cty.Value{
			cty.StringVal("a"),
			cty.StringVal("b"),
		}).Mark("oh no"), // NOTE: Sets cannot have individually-marked elements
		cty.ObjectVal(map[string]cty.Value{
			"string": cty.StringVal("a"),
			"bool":   cty.True,
		}).Mark("oh no"),
		cty.ObjectVal(map[string]cty.Value{
			"string": cty.StringVal("a"),
			"bool":   cty.True.Mark("oh no"),
		}),
		cty.MapVal(map[string]cty.Value{
			"a": cty.StringVal("a-val"),
			"b": cty.StringVal("b-val"),
		}).Mark("oh no"),
		cty.MapVal(map[string]cty.Value{
			"a": cty.StringVal("a-val"),
			"b": cty.StringVal("b-val").Mark("oh no"),
		}),
	}

	for _, test := range tests {
		t.Run(test.GoString(), func(t *testing.T) {
			var transformer valueIdentityTransformer
			got := TransformValue[cty.Value](test, &transformer)
			if !test.RawEquals(got) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test)
			}
		})
	}
}

type valueIdentityTransformer struct {
	marksStack [][]cty.PathValueMarks
}

var _ ValueTransformer[cty.Value] = (*valueIdentityTransformer)(nil)

func (t *valueIdentityTransformer) EnterValue(input cty.Value, path cty.Path) cty.Value {
	value, marks := input.UnmarkDeepWithPaths()
	t.pushMarks(marks)
	return value
}

func (t *valueIdentityTransformer) ExitLeafValue(input cty.Value, path cty.Path) cty.Value {
	return input.MarkWithPaths(t.popMarks())
}

func (t *valueIdentityTransformer) ExitMappingValue(input map[string]cty.Value, orig cty.Value, path cty.Path) cty.Value {
	ty := orig.Type()
	switch {
	case ty.IsMapType():
		return cty.MapVal(input).MarkWithPaths(t.popMarks())
	case ty.IsObjectType():
		return cty.ObjectVal(input).MarkWithPaths(t.popMarks())
	default:
		panic(fmt.Sprintf("unsupported type %#v", ty))
	}
}

func (t *valueIdentityTransformer) ExitSequenceValue(input []cty.Value, orig cty.Value, path cty.Path) cty.Value {
	ty := orig.Type()
	switch {
	case ty.IsListType():
		return cty.ListVal(input).MarkWithPaths(t.popMarks())
	case ty.IsSetType():
		return cty.SetVal(input).MarkWithPaths(t.popMarks())
	case ty.IsTupleType():
		return cty.TupleVal(input).MarkWithPaths(t.popMarks())
	default:
		panic(fmt.Sprintf("unsupported type %#v", ty))
	}
}

func (t *valueIdentityTransformer) pushMarks(marks []cty.PathValueMarks) {
	t.marksStack = append(t.marksStack, marks)
}

func (t *valueIdentityTransformer) popMarks() []cty.PathValueMarks {
	l := len(t.marksStack)
	ret, new := t.marksStack[l-1], t.marksStack[:l-1]
	t.marksStack = new
	return ret
}

func TestTransformValue_JSON(t *testing.T) {
	t.Parallel()

	// This is a semi-realistic test of transforming a cty.Value into
	// a slice of json.Token values. In practice encoding/json only uses
	// json.Token for _decoding_ and so there's not anything really useful
	// to do with that result, but it stands in for deriving a corresponding
	// value in some other infoset while only relying on stuff in the Go
	// standard library.

	tests := []struct {
		input cty.Value
		want  []json.Token
	}{
		{
			cty.NullVal(cty.String),
			[]json.Token{nil},
		},
		{
			cty.StringVal("hello"),
			[]json.Token{"hello"},
		},
		{
			cty.NumberIntVal(1),
			[]json.Token{json.Number("1")},
		},
		{
			cty.True,
			[]json.Token{true},
		},
		{
			cty.TupleVal([]cty.Value{
				cty.StringVal("a"),
				cty.True,
			}),
			[]json.Token{
				json.Delim('['),
				"a",
				true,
				json.Delim(']'),
			},
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
			}),
			[]json.Token{
				json.Delim('['),
				"a",
				"b",
				json.Delim(']'),
			},
		},
		{
			cty.SetVal([]cty.Value{
				// NOTE: Using strings intentionally because sets of string
				// have a well-defined iteration order for a stable result.
				cty.StringVal("a"),
				cty.StringVal("b"),
			}),
			[]json.Token{
				json.Delim('['),
				"a",
				"b",
				json.Delim(']'),
			},
		},
		{
			cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal("a"),
				"bool":   cty.True,
			}),
			[]json.Token{
				json.Delim('{'),
				"bool",
				true,
				"string",
				"a",
				json.Delim('}'),
			},
		},
		{
			cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("a-val"),
				"b": cty.StringVal("b-val"),
			}),
			[]json.Token{
				json.Delim('{'),
				"a",
				"a-val",
				"b",
				"b-val",
				json.Delim('}'),
			},
		},
	}

	var transformer valueToJSONTransformer
	for _, test := range tests {
		t.Run(test.input.GoString(), func(t *testing.T) {
			tokens, err := TransformValue[Fallible[[]json.Token]](test.input, transformer).Result()
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(test.want, tokens); diff != "" {
				t.Errorf("unexpected result\n%s", diff)
			}
		})
	}
}

type valueToJSONTransformer struct{}

var _ ValueTransformer[Fallible[[]json.Token]] = valueToJSONTransformer{}

func (valueToJSONTransformer) EnterValue(input cty.Value, path cty.Path) cty.Value {
	// No inward transform required
	return input
}

func (valueToJSONTransformer) ExitLeafValue(v cty.Value, path cty.Path) Fallible[[]json.Token] {
	ty := v.Type()
	switch {
	case !v.IsKnown():
		return jsonTokenError(path.NewErrorf("cannot encode an unknown value"))
	case v.IsMarked():
		return jsonTokenError(path.NewErrorf("cannot encode a marked value"))
	case v.IsNull():
		return Success([]json.Token{nil})
	case ty == cty.String:
		return Success([]json.Token{v.AsString()})
	case ty == cty.Number:
		return Success([]json.Token{json.Number(v.AsBigFloat().Text('f', -1))})
	case ty == cty.Bool:
		if v.True() {
			return Success([]json.Token{true})
		} else {
			return Success([]json.Token{false})
		}
	case ty.IsCapsuleType():
		return jsonTokenError(path.NewErrorf("cannot encode a capsule type"))
	default:
		return jsonTokenError(path.NewErrorf("don't know what to do with %#v", v))
	}
}

func (valueToJSONTransformer) ExitSequenceValue(input []Fallible[[]json.Token], orig cty.Value, path cty.Path) Fallible[[]json.Token] {
	var ret []json.Token
	ret = append(ret, json.Delim('['))
	for _, item := range input {
		tokens, err := item.Result()
		if err != nil {
			// Just bubble up the first error we encounter
			return jsonTokenError(err)
		}
		ret = append(ret, tokens...)
	}
	ret = append(ret, json.Delim(']'))
	return Success(ret)
}

func (valueToJSONTransformer) ExitMappingValue(input map[string]Fallible[[]json.Token], orig cty.Value, path cty.Path) Fallible[[]json.Token] {
	var ret []json.Token
	ret = append(ret, json.Delim('{'))

	keys := make([]string, 0, len(input))
	for k := range input {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		item := input[k]
		tokens, err := item.Result()
		if err != nil {
			// Just bubble up the first error we encounter
			return jsonTokenError(err)
		}
		ret = append(ret, k)
		ret = append(ret, tokens...)
	}

	ret = append(ret, json.Delim('}'))
	return Success(ret)
}

func jsonTokenError(err error) Fallible[[]json.Token] {
	return Error[[]json.Token](err)
}
