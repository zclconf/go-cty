package convert

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestUnify(t *testing.T) {
	tests := []struct {
		Input           []cty.Type
		WantType        cty.Type
		WantConversions []bool
	}{
		{
			[]cty.Type{},
			cty.NilType,
			nil,
		},
		{
			[]cty.Type{cty.String},
			cty.String,
			[]bool{false},
		},
		{
			[]cty.Type{cty.Number},
			cty.Number,
			[]bool{false},
		},
		{
			[]cty.Type{cty.Number, cty.Number},
			cty.Number,
			[]bool{false, false},
		},
		{
			[]cty.Type{cty.Number, cty.String},
			cty.String,
			[]bool{true, false},
		},
		{
			[]cty.Type{cty.String, cty.Number},
			cty.String,
			[]bool{false, true},
		},
		{
			[]cty.Type{cty.Bool, cty.String, cty.Number},
			cty.String,
			[]bool{true, false, true},
		},
		{
			[]cty.Type{cty.Bool, cty.Number},
			cty.NilType,
			nil,
		},
		{
			[]cty.Type{
				cty.Object(map[string]cty.Type{"foo": cty.String}),
				cty.Object(map[string]cty.Type{"foo": cty.String}),
			},
			cty.Object(map[string]cty.Type{"foo": cty.String}),
			[]bool{false, false},
		},
		{
			[]cty.Type{
				cty.Object(map[string]cty.Type{"foo": cty.String}),
				cty.Object(map[string]cty.Type{"foo": cty.Number}),
			},
			cty.Object(map[string]cty.Type{"foo": cty.String}),
			[]bool{false, true},
		},
		{
			[]cty.Type{
				cty.Object(map[string]cty.Type{"foo": cty.String}),
				cty.Object(map[string]cty.Type{"bar": cty.Number}),
			},
			cty.Map(cty.String),
			[]bool{true, true},
		},
		{
			[]cty.Type{
				cty.Object(map[string]cty.Type{"foo": cty.String}),
				cty.EmptyObject,
			},
			cty.Map(cty.String),
			[]bool{true, true},
		},
		{
			[]cty.Type{
				cty.Object(map[string]cty.Type{"foo": cty.Bool}),
				cty.Object(map[string]cty.Type{"bar": cty.Number}),
			},
			cty.NilType,
			nil,
		},
		{
			[]cty.Type{
				cty.Object(map[string]cty.Type{"foo": cty.Bool}),
				cty.Object(map[string]cty.Type{"foo": cty.Number}),
			},
			cty.NilType,
			nil,
		},
		{
			[]cty.Type{
				cty.Tuple([]cty.Type{cty.String}),
				cty.Tuple([]cty.Type{cty.String}),
			},
			cty.Tuple([]cty.Type{cty.String}),
			[]bool{false, false},
		},
		{
			[]cty.Type{
				cty.Tuple([]cty.Type{cty.String}),
				cty.Tuple([]cty.Type{cty.Number}),
			},
			cty.Tuple([]cty.Type{cty.String}),
			[]bool{false, true},
		},
		{
			[]cty.Type{
				cty.Tuple([]cty.Type{cty.String}),
				cty.Tuple([]cty.Type{cty.String, cty.Number}),
			},
			cty.List(cty.String),
			[]bool{true, true},
		},
		{
			[]cty.Type{
				cty.Tuple([]cty.Type{cty.String}),
				cty.EmptyTuple,
			},
			cty.List(cty.String),
			[]bool{true, true},
		},
		{
			[]cty.Type{
				cty.Tuple([]cty.Type{cty.Bool}),
				cty.Tuple([]cty.Type{cty.Number}),
			},
			cty.NilType,
			nil,
		},
		{
			// objects can unify as map(string) within the tuples
			[]cty.Type{
				cty.Tuple([]cty.Type{
					cty.Object(map[string]cty.Type{
						"a": cty.String,
					}),
					cty.Object(map[string]cty.Type{
						"a": cty.String,
					}),
				}),
				cty.Tuple([]cty.Type{
					cty.Object(map[string]cty.Type{
						"a": cty.String,
						"b": cty.String,
					}),
				}),
			},
			cty.List(cty.Map(cty.String)),
			[]bool{true, true},
		},
		{
			// unifies to the same result as above, since the only difference
			// is the addition of a list
			[]cty.Type{
				cty.List(cty.Object(map[string]cty.Type{
					"a": cty.String,
				})),
				cty.Tuple([]cty.Type{
					cty.Object(map[string]cty.Type{
						"a": cty.String,
						"b": cty.String,
					}),
				}),
				cty.Tuple([]cty.Type{
					cty.Object(map[string]cty.Type{
						"a": cty.String,
						"b": cty.String,
					}),
					cty.Object(map[string]cty.Type{
						"c": cty.String,
						"d": cty.String,
					}),
				}),
			},
			cty.List(cty.Map(cty.String)),
			[]bool{true, true, true},
		},
		{
			// Ensure the map does not change the unification process
			[]cty.Type{
				cty.List(cty.Object(map[string]cty.Type{
					"a": cty.String,
				})),
				cty.List(cty.Map(cty.String)),
				cty.Tuple([]cty.Type{
					cty.Map(cty.String),
					cty.Object(map[string]cty.Type{
						"a": cty.String,
						"b": cty.String,
					}),
				}),
			},
			cty.List(cty.Map(cty.String)),
			[]bool{true, false, true},
		},
		{
			// different tuple lengths unify as a list, and the objects can
			// unify as maps
			[]cty.Type{
				cty.Tuple([]cty.Type{
					cty.Object(map[string]cty.Type{
						"a": cty.String,
						"b": cty.Number,
					}),
					cty.Object(map[string]cty.Type{
						"a": cty.String,
						"b": cty.Number,
					}),
				}),
				cty.Tuple([]cty.Type{
					cty.Object(map[string]cty.Type{
						"a": cty.String,
					}),
				}),
			},
			cty.List(cty.Map(cty.String)),
			[]bool{true, true},
		},
		{
			// the equivalent tuple lengths still unify as a tuple, though the
			// objects are unified as a map
			[]cty.Type{
				cty.Tuple([]cty.Type{
					cty.Object(map[string]cty.Type{
						"a": cty.String,
						"b": cty.Number,
					}),
				}),
				cty.Tuple([]cty.Type{
					cty.Object(map[string]cty.Type{
						"a": cty.String,
					}),
				}),
			},
			cty.Tuple([]cty.Type{cty.Map(cty.String)}),
			[]bool{true, true},
		},
		{
			// This should unify to like the tuple above
			[]cty.Type{
				cty.List(
					cty.Object(map[string]cty.Type{
						"a": cty.Number,
						"b": cty.String,
					}),
				),
				cty.Tuple([]cty.Type{
					cty.Object(map[string]cty.Type{
						"a": cty.String,
					}),
				}),
			},
			cty.List(cty.Map(cty.String)),
			[]bool{true, true},
		},
		{
			// This should also unify like the previous 2 examples
			[]cty.Type{
				cty.List(
					cty.Object(map[string]cty.Type{
						"a": cty.Number,
						"b": cty.String,
					}),
				),
				cty.List(cty.Object(map[string]cty.Type{
					"a": cty.String,
				})),
			},
			cty.List(cty.Map(cty.String)),
			[]bool{true, true},
		},
		{
			// Objects and maps should unify along with the surrounding lists
			// and tuples.
			[]cty.Type{
				cty.List(cty.Object(map[string]cty.Type{
					"a": cty.Object(map[string]cty.Type{
						"a": cty.String,
					}),
					"b": cty.Object(map[string]cty.Type{
						"a": cty.String,
						"b": cty.String,
					}),
				})),
				cty.List(cty.Map(
					cty.Object(map[string]cty.Type{
						"a": cty.String,
						"b": cty.String,
					}),
				)),
			},
			cty.List(cty.Map(cty.Map(cty.String))),
			[]bool{true, true},
		},
		{
			// objects can unify as maps within objects
			[]cty.Type{
				cty.Object(map[string]cty.Type{
					"a": cty.Object(map[string]cty.Type{
						"a": cty.String,
					}),
				}),
				cty.Object(map[string]cty.Type{
					"a": cty.Object(map[string]cty.Type{
						"a": cty.String,
						"b": cty.String,
					}),
				}),
			},
			cty.Object(map[string]cty.Type{
				"a": cty.Map(cty.String),
			}),
			[]bool{true, true},
		},
		{
			// nested objects can unify as maps
			[]cty.Type{
				cty.Object(map[string]cty.Type{
					"a": cty.Object(map[string]cty.Type{
						"a": cty.String,
					}),
					"b": cty.Object(map[string]cty.Type{
						"a": cty.String,
						"b": cty.String,
					}),
				}),
				cty.Map(
					cty.Object(map[string]cty.Type{
						"a": cty.String,
						"b": cty.String,
					}),
				),
			},
			cty.Map(cty.Map(cty.String)),
			[]bool{true, true},
		},
		{
			// nested tuples and lists can unify along with the surrounding
			// objects and maps
			[]cty.Type{
				cty.Object(map[string]cty.Type{
					"a": cty.Object(map[string]cty.Type{
						"a": cty.List(cty.String),
					}),
					"b": cty.Object(map[string]cty.Type{
						"a": cty.Tuple([]cty.Type{
							cty.String,
						}),
						"b": cty.List(cty.String),
					}),
				}),
				cty.Map(
					cty.Object(map[string]cty.Type{
						"a": cty.List(cty.String),
						"b": cty.List(cty.String),
					}),
				),
			},
			cty.Map(cty.Map(cty.List(cty.String))),
			[]bool{true, true},
		},
		{
			// objects can unify as maps containing objects when all attributes
			// match
			[]cty.Type{
				cty.Object(map[string]cty.Type{
					"a": cty.Object(map[string]cty.Type{
						"a": cty.String,
					}),
					"b": cty.Object(map[string]cty.Type{
						"a": cty.String,
					}),
				}),
				cty.Map(
					cty.Object(map[string]cty.Type{
						"a": cty.String,
					}),
				),
			},
			cty.Map(
				cty.Object(map[string]cty.Type{
					"a": cty.String,
				}),
			),
			[]bool{true, false},
		},
		{
			// objects can unify as maps with dynamic types
			[]cty.Type{
				cty.Object(map[string]cty.Type{
					"a": cty.Object(map[string]cty.Type{
						"a": cty.String,
					}),
					"b": cty.Object(map[string]cty.Type{
						"a": cty.String,
					}),
				}),
				cty.Map(cty.DynamicPseudoType),
				cty.Map(
					cty.Object(map[string]cty.Type{
						"a": cty.String,
					}),
				),
			},
			cty.Map(cty.DynamicPseudoType),
			[]bool{true, false, true},
		},
		{
			// deeply nested objects and maps can unify
			[]cty.Type{
				cty.Object(map[string]cty.Type{
					"a": cty.Object(map[string]cty.Type{
						"a": cty.Object(map[string]cty.Type{
							"a": cty.String,
						}),
					}),
					"b": cty.Object(map[string]cty.Type{
						"c": cty.Object(map[string]cty.Type{
							"d": cty.String,
						}),
					}),
				}),
				cty.Map(cty.Map(cty.Map(cty.String))),
			},
			cty.Map(cty.Map(cty.Map(cty.String))),
			[]bool{true, false},
		},
		{
			// deeply nested objects with maps can unify as maps
			[]cty.Type{
				cty.Map(cty.Map(cty.Map(cty.String))),
				cty.Object(map[string]cty.Type{
					"a": cty.Object(map[string]cty.Type{
						"a": cty.Object(map[string]cty.Type{
							"a": cty.String,
						}),
						"b": cty.Map(cty.String),
					}),
					"b": cty.Map(cty.Map(cty.String)),
				}),
			},
			cty.Map(cty.Map(cty.Map(cty.String))),
			[]bool{false, true},
		},
		{
			[]cty.Type{
				cty.DynamicPseudoType,
				cty.Tuple([]cty.Type{cty.Number}),
			},
			cty.DynamicPseudoType,
			[]bool{true, true},
		},
		{
			[]cty.Type{
				cty.DynamicPseudoType,
				cty.Object(map[string]cty.Type{"num": cty.Number}),
			},
			cty.DynamicPseudoType,
			[]bool{true, true},
		},
		{
			[]cty.Type{
				cty.Tuple([]cty.Type{cty.Number}),
				cty.DynamicPseudoType,
				cty.Object(map[string]cty.Type{"num": cty.Number}),
			},
			cty.NilType,
			nil,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.Input), func(t *testing.T) {
			gotType, gotConvs := Unify(test.Input)
			if gotType == cty.NilType && test.WantType == cty.NilType {
				// okay!
			} else if ((gotType == cty.NilType) != (test.WantType == cty.NilType)) || !test.WantType.Equals(gotType) {
				t.Errorf("wrong result type\ngot:  %#v\nwant: %#v", gotType, test.WantType)
			}

			gotConvsNil := gotConvs == nil
			wantConvsNil := test.WantConversions == nil
			if gotConvsNil && wantConvsNil {
				// Success!
				return
			}

			if gotConvsNil != wantConvsNil {
				if gotConvsNil {
					t.Fatalf("got nil conversions; want %#v", test.WantConversions)
				} else {
					t.Fatalf("got conversions; want nil")
				}
			}

			gotConvsBool := make([]bool, len(gotConvs))
			for i, f := range gotConvs {
				gotConvsBool[i] = f != nil
			}

			if !reflect.DeepEqual(gotConvsBool, test.WantConversions) {
				t.Fatalf(
					"wrong conversions\ngot:  %#v\nwant: %#v",
					gotConvsBool, test.WantConversions,
				)
			}
		})
	}
}
