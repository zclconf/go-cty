package convert

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		Value     cty.Value
		Type      cty.Type
		Want      cty.Value
		WantError bool
	}{
		{
			Value: cty.StringVal("hello"),
			Type:  cty.String,
			Want:  cty.StringVal("hello"),
		},
		{
			Value: cty.StringVal("1"),
			Type:  cty.Number,
			Want:  cty.NumberIntVal(1),
		},
		{
			Value: cty.StringVal("1.5"),
			Type:  cty.Number,
			Want:  cty.NumberFloatVal(1.5),
		},
		{
			Value:     cty.StringVal("hello"),
			Type:      cty.Number,
			WantError: true,
		},
		{
			Value: cty.StringVal("true"),
			Type:  cty.Bool,
			Want:  cty.True,
		},
		{
			Value: cty.StringVal("1"),
			Type:  cty.Bool,
			Want:  cty.True,
		},
		{
			Value: cty.StringVal("false"),
			Type:  cty.Bool,
			Want:  cty.False,
		},
		{
			Value: cty.StringVal("0"),
			Type:  cty.Bool,
			Want:  cty.False,
		},
		{
			Value:     cty.StringVal("hello"),
			Type:      cty.Bool,
			WantError: true,
		},
		{
			Value: cty.NumberIntVal(4),
			Type:  cty.String,
			Want:  cty.StringVal("4"),
		},
		{
			Value: cty.NumberFloatVal(3.14159265359),
			Type:  cty.String,
			Want:  cty.StringVal("3.14159265359"),
		},
		{
			Value: cty.True,
			Type:  cty.String,
			Want:  cty.StringVal("true"),
		},
		{
			Value: cty.False,
			Type:  cty.String,
			Want:  cty.StringVal("false"),
		},
		{
			Value: cty.UnknownVal(cty.String),
			Type:  cty.Number,
			Want:  cty.UnknownVal(cty.Number),
		},
		{
			Value: cty.UnknownVal(cty.Number),
			Type:  cty.String,
			Want:  cty.UnknownVal(cty.String),
		},
		{
			Value: cty.DynamicVal,
			Type:  cty.String,
			Want:  cty.UnknownVal(cty.String),
		},
		{
			Value: cty.StringVal("hello"),
			Type:  cty.DynamicPseudoType,
			Want:  cty.StringVal("hello"),
		},
		{
			Value: cty.ListVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.NumberIntVal(10),
			}),
			Type: cty.List(cty.String),
			Want: cty.ListVal([]cty.Value{
				cty.StringVal("5"),
				cty.StringVal("10"),
			}),
		},
		{
			Value: cty.ListVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.NumberIntVal(10),
			}),
			Type: cty.List(cty.DynamicPseudoType),
			Want: cty.ListVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.NumberIntVal(10),
			}),
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"type":        cty.StringVal("ingress"),
					"from_port":   cty.NumberIntVal(-1),
					"to_port":     cty.NumberIntVal(-1),
					"protocol":    cty.StringVal("icmp"),
					"description": cty.StringVal("ICMP in"),
					"cidr":        cty.TupleVal([]cty.Value{cty.StringVal("0.0.0.0/0")}),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"type":        cty.StringVal("ingress"),
					"from_port":   cty.NumberIntVal(22),
					"to_port":     cty.NumberIntVal(22),
					"protocol":    cty.StringVal("tcp"),
					"description": cty.StringVal("SSH from Bastion"),
					"source_sg":   cty.StringVal("sg-abc123"),
				}),
			}),
			Type:      cty.List(cty.DynamicPseudoType),
			WantError: true, // there is no type that both tuple elements can unify to for conversion to list
		},
		{
			Value: cty.SetVal([]cty.Value{
				cty.StringVal("5"),
				cty.UnknownVal(cty.String),
			}),
			Type: cty.Set(cty.Number),
			Want: cty.SetVal([]cty.Value{cty.NumberIntVal(5), cty.UnknownVal(cty.Number)}),
		},
		{
			Value: cty.SetVal([]cty.Value{
				cty.StringVal("5"),
				cty.StringVal("10"),
			}),
			Type: cty.List(cty.String),
			Want: cty.ListVal([]cty.Value{
				// NOTE: This results depends on the traversal order of the
				// set, which may change if the set implementation changes.
				cty.StringVal("10"),
				cty.StringVal("5"),
			}),
		},
		{
			Value: cty.SetVal([]cty.Value{
				cty.StringVal("5"),
				cty.StringVal("10"),
			}),
			Type: cty.List(cty.DynamicPseudoType),
			Want: cty.ListVal([]cty.Value{
				// NOTE: This results depends on the traversal order of the
				// set, which may change if the set implementation changes.
				cty.StringVal("10"),
				cty.StringVal("5"),
			}),
		},
		{
			Value: cty.SetVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.NumberIntVal(10),
			}),
			Type: cty.List(cty.String),
			Want: cty.ListVal([]cty.Value{
				// NOTE: This results depends on the traversal order of the
				// set, which may change if the set implementation changes.
				cty.StringVal("5"),
				cty.StringVal("10"),
			}),
		},
		{
			Value: cty.SetVal([]cty.Value{
				cty.StringVal("5"),
				cty.UnknownVal(cty.String),
			}),
			Type: cty.List(cty.String),
			Want: cty.UnknownVal(cty.List(cty.String)),
		},
		{
			Value: cty.SetVal([]cty.Value{
				cty.UnknownVal(cty.String),
			}),
			Type: cty.List(cty.String),
			// We get a known list value this time because even though we
			// don't know the single value that's in the list, we _do_ know
			// that there are no other values in the set for it to coalesce
			// with.
			Want: cty.ListVal([]cty.Value{
				cty.UnknownVal(cty.String),
			}),
		},
		{
			Value: cty.ListVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.NumberIntVal(10),
				cty.NumberIntVal(10),
			}),
			Type: cty.Set(cty.String),
			Want: cty.SetVal([]cty.Value{
				cty.StringVal("5"),
				cty.StringVal("10"),
			}),
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.StringVal("hello"),
			}),
			Type: cty.List(cty.String),
			Want: cty.ListVal([]cty.Value{
				cty.StringVal("5"),
				cty.StringVal("hello"),
			}),
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.StringVal("12"),
			}),
			Type: cty.List(cty.Number),
			Want: cty.ListVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.NumberIntVal(12),
			}),
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.NumberIntVal(10),
			}),
			Type: cty.List(cty.DynamicPseudoType),
			Want: cty.ListVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.NumberIntVal(10),
			}),
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.StringVal("hello"),
			}),
			Type: cty.List(cty.DynamicPseudoType),
			Want: cty.ListVal([]cty.Value{
				cty.StringVal("5"),
				cty.StringVal("hello"),
			}),
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.StringVal("hello"),
			}),
			Type: cty.Set(cty.DynamicPseudoType),
			Want: cty.SetVal([]cty.Value{
				cty.StringVal("5"),
				cty.StringVal("hello"),
			}),
		},
		{
			Value: cty.ListValEmpty(cty.String),
			Type:  cty.Set(cty.DynamicPseudoType),
			Want:  cty.SetValEmpty(cty.String),
		},
		{
			Value: cty.SetValEmpty(cty.String),
			Type:  cty.List(cty.DynamicPseudoType),
			Want:  cty.ListValEmpty(cty.String),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"num": cty.NumberIntVal(5),
				"str": cty.StringVal("hello"),
			}),
			Type: cty.Map(cty.String),
			Want: cty.MapVal(map[string]cty.Value{
				"num": cty.StringVal("5"),
				"str": cty.StringVal("hello"),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"num": cty.NumberIntVal(5),
				"str": cty.StringVal("12"),
			}),
			Type: cty.Map(cty.Number),
			Want: cty.MapVal(map[string]cty.Value{
				"num": cty.NumberIntVal(5),
				"str": cty.NumberIntVal(12),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"num1": cty.NumberIntVal(5),
				"num2": cty.NumberIntVal(10),
			}),
			Type: cty.Map(cty.DynamicPseudoType),
			Want: cty.MapVal(map[string]cty.Value{
				"num1": cty.NumberIntVal(5),
				"num2": cty.NumberIntVal(10),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"num": cty.NumberIntVal(5),
				"str": cty.StringVal("hello"),
			}),
			Type: cty.Map(cty.DynamicPseudoType),
			Want: cty.MapVal(map[string]cty.Value{
				"num": cty.StringVal("5"),
				"str": cty.StringVal("hello"),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"list":  cty.ListValEmpty(cty.Bool),
				"tuple": cty.EmptyTupleVal,
			}),
			Type: cty.Map(cty.DynamicPseudoType),
			Want: cty.MapVal(map[string]cty.Value{
				"list":  cty.ListValEmpty(cty.Bool),
				"tuple": cty.ListValEmpty(cty.Bool),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"map": cty.MapValEmpty(cty.String),
				"obj": cty.EmptyObjectVal,
			}),
			Type: cty.Map(cty.DynamicPseudoType),
			Want: cty.MapVal(map[string]cty.Value{
				"map": cty.MapValEmpty(cty.String),
				"obj": cty.MapValEmpty(cty.String),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"num":  cty.NumberIntVal(5),
				"bool": cty.True,
			}),
			Type:      cty.Map(cty.DynamicPseudoType),
			WantError: true, // no common base type to unify to
		},
		{
			Value: cty.MapVal(map[string]cty.Value{
				"greeting": cty.StringVal("Hello"),
				"name":     cty.StringVal("John"),
			}),
			Type: cty.Map(cty.DynamicPseudoType),
			Want: cty.MapVal(map[string]cty.Value{
				"greeting": cty.StringVal("Hello"),
				"name":     cty.StringVal("John"),
			}),
		},
		{
			Value: cty.MapVal(map[string]cty.Value{
				"greeting": cty.StringVal("Hello"),
				"name":     cty.StringVal("John"),
			}),
			Type: cty.Object(map[string]cty.Type{
				"greeting": cty.String,
				"name":     cty.String,
			}),
			Want: cty.ObjectVal(map[string]cty.Value{
				"greeting": cty.StringVal("Hello"),
				"name":     cty.StringVal("John"),
			}),
		},
		{
			Value: cty.MapVal(map[string]cty.Value{
				"greeting": cty.StringVal("Hello"),
				"name":     cty.StringVal("John"),
			}),
			Type: cty.Object(map[string]cty.Type{
				"greeting": cty.List(cty.String),
				"name":     cty.String,
			}),
			WantError: true, // "greeting" cannot be converted
		},
		{
			Value: cty.MapVal(map[string]cty.Value{
				"greeting": cty.StringVal("Hello"),
				"name":     cty.StringVal("John"),
			}),
			Type: cty.Object(map[string]cty.Type{
				"name": cty.String,
			}),
			Want: cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal("John"),
			}),
		},
		{
			Value: cty.MapVal(map[string]cty.Value{
				"name": cty.StringVal("John"),
			}),
			Type: cty.Object(map[string]cty.Type{
				"name":     cty.String,
				"greeting": cty.String,
			}),
			WantError: true, // map has no element for required attribute "greeting"
		},
		{
			Value: cty.MapVal(map[string]cty.Value{
				"name": cty.StringVal("John"),
			}),
			Type: cty.ObjectWithOptionalAttrs(
				map[string]cty.Type{
					"name":     cty.String,
					"greeting": cty.String,
				},
				[]string{"greeting"},
			),
			Want: cty.ObjectVal(map[string]cty.Value{
				"greeting": cty.NullVal(cty.String),
				"name":     cty.StringVal("John"),
			}),
		},
		{
			Value: cty.MapVal(map[string]cty.Value{
				"a": cty.NumberIntVal(2),
				"b": cty.NumberIntVal(5),
			}),
			Type: cty.Map(cty.String),
			Want: cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("2"),
				"b": cty.StringVal("5"),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("foo value"),
				"bar": cty.StringVal("bar value"),
			}),
			Type: cty.Object(map[string]cty.Type{
				"foo": cty.String,
			}),
			Want: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("foo value"),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.True,
			}),
			Type: cty.Object(map[string]cty.Type{
				"foo": cty.String,
			}),
			Want: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("true"),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.DynamicVal,
			}),
			Type: cty.Object(map[string]cty.Type{
				"foo": cty.String,
			}),
			Want: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.UnknownVal(cty.String),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.NullVal(cty.String),
			}),
			Type: cty.Object(map[string]cty.Type{
				"foo": cty.String,
			}),
			Want: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.NullVal(cty.String),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.True,
			}),
			Type: cty.Object(map[string]cty.Type{
				"foo": cty.DynamicPseudoType,
			}),
			Want: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.True,
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"bar": cty.StringVal("bar value"),
			}),
			Type: cty.Object(map[string]cty.Type{
				"foo": cty.String,
			}),
			WantError: true, // given value must have superset object type
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"bar": cty.StringVal("bar value"),
			}),
			Type: cty.Object(map[string]cty.Type{
				"foo": cty.String,
				"baz": cty.String,
			}),
			WantError: true, // given value must have superset object type
		},
		{
			Value: cty.EmptyObjectVal,
			Type: cty.Object(map[string]cty.Type{
				"foo": cty.String,
				"bar": cty.String,
				"baz": cty.String,
			}),
			WantError: true, // given value must have superset object type
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"bar": cty.StringVal("bar value"),
			}),
			Type: cty.ObjectWithOptionalAttrs(
				map[string]cty.Type{
					"foo": cty.String,
					"bar": cty.String,
				},
				[]string{"foo"},
			),
			Want: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.NullVal(cty.String),
				"bar": cty.StringVal("bar value"),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("foo value"),
				"bar": cty.StringVal("bar value"),
			}),
			Type: cty.ObjectWithOptionalAttrs(
				map[string]cty.Type{
					"foo": cty.String,
					"bar": cty.String,
				},
				[]string{"foo"},
			),
			Want: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("foo value"),
				"bar": cty.StringVal("bar value"),
			}),
		},
		{
			Value: cty.EmptyObjectVal,
			Type: cty.ObjectWithOptionalAttrs(
				map[string]cty.Type{
					"foo": cty.String,
					"bar": cty.String,
				},
				[]string{"foo"},
			),
			WantError: true, // Attribute "bar" is required
		},
		{
			Value: cty.NullVal(cty.DynamicPseudoType),
			Type: cty.ObjectWithOptionalAttrs(
				map[string]cty.Type{
					"foo": cty.String,
					"bar": cty.String,
				},
				[]string{"foo"},
			),
			Want: cty.NullVal(cty.Object(map[string]cty.Type{
				"foo": cty.String,
				"bar": cty.String,
			})),
		},
		{
			Value: cty.ListVal([]cty.Value{
				cty.NullVal(cty.DynamicPseudoType),
				cty.ObjectVal(map[string]cty.Value{
					"bar": cty.StringVal("bar value"),
				}),
			}),
			Type: cty.List(cty.ObjectWithOptionalAttrs(
				map[string]cty.Type{
					"foo": cty.String,
					"bar": cty.String,
				},
				[]string{"foo"},
			)),
			Want: cty.ListVal([]cty.Value{
				cty.NullVal(cty.Object(map[string]cty.Type{
					"foo": cty.String,
					"bar": cty.String,
				})),
				cty.ObjectVal(map[string]cty.Value{
					"foo": cty.NullVal(cty.String),
					"bar": cty.StringVal("bar value"),
				}),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.True,
			}),
			Type: cty.Object(map[string]cty.Type{
				"foo": cty.Number,
			}),
			WantError: true, // recursive conversion from bool to number is impossible
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.UnknownVal(cty.Bool),
			}),
			Type: cty.Object(map[string]cty.Type{
				"foo": cty.Number,
			}),
			WantError: true, // recursive conversion from bool to number is impossible
		},
		{
			Value: cty.NullVal(cty.String),
			Type:  cty.DynamicPseudoType,
			Want:  cty.NullVal(cty.String),
		},
		{
			Value: cty.UnknownVal(cty.String),
			Type:  cty.DynamicPseudoType,
			Want:  cty.UnknownVal(cty.String),
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			Type: cty.Tuple([]cty.Type{
				cty.String,
			}),
			Want: cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
			}),
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.True,
			}),
			Type: cty.Tuple([]cty.Type{
				cty.String,
			}),
			Want: cty.TupleVal([]cty.Value{
				cty.StringVal("true"),
			}),
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.True,
			}),
			Type:      cty.EmptyTuple,
			WantError: true,
		},
		{
			Value: cty.EmptyTupleVal,
			Type: cty.Tuple([]cty.Type{
				cty.String,
			}),
			WantError: true,
		},
		{
			Value: cty.EmptyTupleVal,
			Type:  cty.Set(cty.String),
			Want:  cty.SetValEmpty(cty.String),
		},

		// Marks on values should propagate, even deeply.
		{
			Value: cty.StringVal("hello").Mark(1),
			Type:  cty.String,
			Want:  cty.StringVal("hello").Mark(1),
		},
		{
			Value: cty.StringVal("true").Mark(1),
			Type:  cty.Bool,
			Want:  cty.True.Mark(1),
		},
		{
			Value: cty.TupleVal([]cty.Value{cty.StringVal("hello").Mark(1)}),
			Type:  cty.List(cty.String),
			Want:  cty.ListVal([]cty.Value{cty.StringVal("hello").Mark(1)}),
		},
		{
			Value: cty.SetVal([]cty.Value{
				cty.StringVal("hello").Mark(1),
				cty.StringVal("hello").Mark(2),
			}),
			Type: cty.Set(cty.String),
			Want: cty.SetVal([]cty.Value{cty.StringVal("hello")}).WithMarks(cty.NewValueMarks(1, 2)),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{"foo": cty.StringVal("hello").Mark(1)}),
			Type:  cty.Map(cty.String),
			Want:  cty.MapVal(map[string]cty.Value{"foo": cty.StringVal("hello").Mark(1)}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("hello").Mark(1),
				"bar": cty.StringVal("world").Mark(1),
			}),
			Type: cty.Object(map[string]cty.Type{"foo": cty.String}),
			Want: cty.ObjectVal(map[string]cty.Value{"foo": cty.StringVal("hello").Mark(1)}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("hello"),
				"bar": cty.StringVal("world").Mark(1),
			}),
			Type: cty.Object(map[string]cty.Type{"foo": cty.String}),
			Want: cty.ObjectVal(map[string]cty.Value{"foo": cty.StringVal("hello")}),
		},
		// reduction of https://github.com/hashicorp/terraform/issues/23804
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"a": cty.ObjectVal(map[string]cty.Value{
					"x": cty.TupleVal([]cty.Value{cty.StringVal("foo")}),
				}),
				"b": cty.ObjectVal(map[string]cty.Value{
					"x": cty.TupleVal([]cty.Value{cty.StringVal("bar")}),
				}),
				"c": cty.ObjectVal(map[string]cty.Value{
					"x": cty.TupleVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar")}),
				}),
			}),
			Type: cty.Map(cty.Map(cty.DynamicPseudoType)),
			Want: cty.MapVal(map[string]cty.Value{
				"a": cty.MapVal(map[string]cty.Value{
					"x": cty.ListVal([]cty.Value{cty.StringVal("foo")}),
				}),
				"b": cty.MapVal(map[string]cty.Value{
					"x": cty.ListVal([]cty.Value{cty.StringVal("bar")}),
				}),
				"c": cty.MapVal(map[string]cty.Value{
					"x": cty.ListVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar")}),
				}),
			}),
		},
		// reduction of https://github.com/hashicorp/terraform/issues/24167
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"a": cty.ObjectVal(map[string]cty.Value{
					"x": cty.NullVal(cty.DynamicPseudoType),
				}),
				"b": cty.ObjectVal(map[string]cty.Value{
					"x": cty.ObjectVal(map[string]cty.Value{
						"c": cty.NumberIntVal(1),
						"d": cty.NumberIntVal(2),
					}),
				}),
			}),
			Type:      cty.Map(cty.Map(cty.Object(map[string]cty.Type{"x": cty.Map(cty.DynamicPseudoType)}))),
			WantError: true,
		},
		// reduction of https://github.com/hashicorp/terraform/issues/23431
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"a": cty.ObjectVal(map[string]cty.Value{
					"x": cty.StringVal("foo"),
				}),
				"b": cty.MapValEmpty(cty.DynamicPseudoType),
			}),
			Type: cty.Map(cty.Map(cty.DynamicPseudoType)),
			Want: cty.MapVal(map[string]cty.Value{
				"a": cty.MapVal(map[string]cty.Value{
					"x": cty.StringVal("foo"),
				}),
				"b": cty.MapValEmpty(cty.String),
			}),
		},
		// reduction of https://github.com/hashicorp/terraform/issues/27269
		{
			Value: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.NullVal(cty.DynamicPseudoType),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"b": cty.ListVal([]cty.Value{
							cty.ObjectVal(map[string]cty.Value{
								"c": cty.StringVal("d"),
							}),
						}),
					}),
				}),
			}),
			Type: cty.List(cty.Object(map[string]cty.Type{
				"a": cty.Object(map[string]cty.Type{
					"b": cty.List(cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"c": cty.String,
						"d": cty.String,
					}, []string{"d"})),
				}),
			})),
			Want: cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.NullVal(cty.Object(map[string]cty.Type{
						"b": cty.List(cty.Object(map[string]cty.Type{
							"c": cty.String,
							"d": cty.String,
						})),
					})),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"b": cty.ListVal([]cty.Value{
							cty.ObjectVal(map[string]cty.Value{
								"c": cty.StringVal("d"),
								"d": cty.NullVal(cty.String),
							}),
						}),
					}),
				}),
			}),
		},
		// When converting null values into nested types which include objects
		// with optional attributes, we expect the resulting value to be of a
		// recursively concretized type.
		{
			Value: cty.NullVal(cty.DynamicPseudoType),
			Type: cty.Object(
				map[string]cty.Type{
					"foo": cty.ObjectWithOptionalAttrs(
						map[string]cty.Type{
							"bar": cty.String,
						},
						[]string{"bar"},
					),
				},
			),
			Want: cty.NullVal(cty.Object(map[string]cty.Type{
				"foo": cty.Object(map[string]cty.Type{
					"bar": cty.String,
				}),
			})),
		},
		// The same nested optional attributes flattening should happen for
		// unknown values, too.
		{
			Value: cty.UnknownVal(cty.DynamicPseudoType),
			Type: cty.Object(
				map[string]cty.Type{
					"foo": cty.ObjectWithOptionalAttrs(
						map[string]cty.Type{
							"bar": cty.String,
						},
						[]string{"bar"},
					),
				},
			),
			Want: cty.UnknownVal(cty.Object(map[string]cty.Type{
				"foo": cty.Object(map[string]cty.Type{
					"bar": cty.String,
				}),
			})),
		},
		// https://github.com/hashicorp/terraform/issues/21588:
		{
			Value: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.EmptyObjectVal,
					"b": cty.NumberIntVal(2),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{"var1": cty.StringVal("val1")}),
					"b": cty.StringVal("2"),
				}),
			}),
			Type: cty.List(cty.Object(map[string]cty.Type{
				"a": cty.DynamicPseudoType,
				"b": cty.String,
			})),
			Want: cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.MapValEmpty(cty.String),
					"b": cty.StringVal("2"),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.MapVal(map[string]cty.Value{"var1": cty.StringVal("val1")}),
					"b": cty.StringVal("2"),
				}),
			}),
			WantError: false,
		},
		// https://github.com/hashicorp/terraform/issues/24377:
		{
			Value: cty.TupleVal([]cty.Value{
				cty.ListVal([]cty.Value{cty.StringVal("a")}),
				cty.StringVal("b"),
				cty.NullVal(cty.DynamicPseudoType),
			}),
			Type:      cty.Set(cty.DynamicPseudoType),
			WantError: true,
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.ListVal([]cty.Value{cty.StringVal("a")}),
				cty.StringVal("b"),
				cty.NullVal(cty.DynamicPseudoType),
			}),
			Type:      cty.List(cty.DynamicPseudoType),
			WantError: true,
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.ListVal([]cty.Value{cty.StringVal("a")}),
				cty.StringVal("b"),
			}),
			Type:      cty.Set(cty.DynamicPseudoType),
			WantError: true,
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.ListVal([]cty.Value{cty.StringVal("a")}),
				cty.StringVal("b"),
			}),
			Type:      cty.List(cty.DynamicPseudoType),
			WantError: true,
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.StringVal("a"),
				cty.NumberIntVal(9),
				cty.NullVal(cty.DynamicPseudoType),
			}),
			Type: cty.Set(cty.DynamicPseudoType),
			Want: cty.SetVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("9"),
				cty.NullVal(cty.DynamicPseudoType),
			}),
			WantError: false,
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.StringVal("a"),
				cty.NumberIntVal(9),
				cty.NullVal(cty.DynamicPseudoType),
			}),
			Type: cty.List(cty.DynamicPseudoType),
			Want: cty.ListVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("9"),
				cty.NullVal(cty.DynamicPseudoType),
			}),
			WantError: false,
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.NullVal(cty.DynamicPseudoType),
				cty.NullVal(cty.DynamicPseudoType),
				cty.NullVal(cty.DynamicPseudoType),
			}),
			Type: cty.Set(cty.DynamicPseudoType),
			Want: cty.SetVal([]cty.Value{
				cty.NullVal(cty.DynamicPseudoType),
			}),
			WantError: false,
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.NullVal(cty.DynamicPseudoType),
				cty.NullVal(cty.DynamicPseudoType),
				cty.NullVal(cty.DynamicPseudoType),
			}),
			Type: cty.List(cty.DynamicPseudoType),
			Want: cty.ListVal([]cty.Value{
				cty.NullVal(cty.DynamicPseudoType),
				cty.NullVal(cty.DynamicPseudoType),
				cty.NullVal(cty.DynamicPseudoType),
			}),
			WantError: false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v to %#v", test.Value, test.Type), func(t *testing.T) {
			got, err := Convert(test.Value, test.Type)

			switch {
			case test.WantError:
				if err == nil {
					t.Errorf("conversion succeeded with %#v; want error", got)
				}
			default:
				if err != nil {
					t.Fatalf("conversion failed: %s", err)
				}

				if !got.RawEquals(test.Want) {
					t.Errorf(
						"wrong result\nvalue: %#v\ntype:  %#v\ngot:   %#v\nwant:  %#v",
						test.Value, test.Type,
						got, test.Want,
					)
				}
			}
		})
	}
}
