package diff

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestDiff_Apply(t *testing.T) {
	tests := []struct {
		name   string
		diff   Diff
		source cty.Value
		want   cty.Value
	}{
		// Replace
		{
			"ReplaceString",
			Diff{
				ReplaceChange{
					Path:     nil,
					OldValue: cty.StringVal("A"),
					NewValue: cty.StringVal("B"),
				},
			},
			cty.StringVal("A"),
			cty.StringVal("B"),
		},
		{
			"ReplaceObject",
			Diff{
				ReplaceChange{
					Path:     cty.GetAttrPath("a"),
					OldValue: cty.StringVal("A"),
					NewValue: cty.StringVal("B"),
				},
			},
			cty.ObjectVal(map[string]cty.Value{"a": cty.StringVal("A")}),
			cty.ObjectVal(map[string]cty.Value{"a": cty.StringVal("B")}),
		},
		{
			"ReplaceMapAdd",
			Diff{
				ReplaceChange{
					Path:     cty.IndexPath(cty.StringVal("a")),
					OldValue: cty.NullVal(cty.String),
					NewValue: cty.StringVal("A"),
				},
			},
			cty.MapValEmpty(cty.String),
			cty.MapVal(map[string]cty.Value{"a": cty.StringVal("A")}),
		},
		{
			"ReplaceMapUpdate",
			Diff{
				ReplaceChange{
					Path:     cty.IndexPath(cty.StringVal("a")),
					OldValue: cty.StringVal("A"),
					NewValue: cty.StringVal("B"),
				},
			},
			cty.MapVal(map[string]cty.Value{"a": cty.StringVal("A")}),
			cty.MapVal(map[string]cty.Value{"a": cty.StringVal("B")}),
		},
		{
			"ReplaceList",
			Diff{
				ReplaceChange{
					Path:     cty.IndexPath(cty.NumberIntVal(1)),
					OldValue: cty.StringVal("B"),
					NewValue: cty.StringVal("X"),
				},
			},
			cty.ListVal([]cty.Value{cty.StringVal("A"), cty.StringVal("B"), cty.StringVal("C")}),
			cty.ListVal([]cty.Value{cty.StringVal("A"), cty.StringVal("X"), cty.StringVal("C")}),
		},
		{
			"ReplaceTuple",
			Diff{
				ReplaceChange{
					Path:     cty.IndexPath(cty.NumberIntVal(1)),
					OldValue: cty.StringVal("B"),
					NewValue: cty.StringVal("X"),
				},
			},
			cty.TupleVal([]cty.Value{cty.StringVal("A"), cty.StringVal("B")}),
			cty.TupleVal([]cty.Value{cty.StringVal("A"), cty.StringVal("X")}),
		},

		// Delete
		{
			"DeleteObject",
			Diff{
				DeleteChange{
					Path:     cty.GetAttrPath("a"),
					OldValue: cty.StringVal("A"),
				},
			},
			cty.ObjectVal(map[string]cty.Value{"a": cty.StringVal("A"), "b": cty.StringVal("B")}),
			cty.ObjectVal(map[string]cty.Value{"b": cty.StringVal("B")}),
		},
		{
			"DeleteMap",
			Diff{
				DeleteChange{
					Path:     cty.IndexPath(cty.StringVal("a")),
					OldValue: cty.StringVal("A"),
				},
			},
			cty.MapVal(map[string]cty.Value{"a": cty.StringVal("A"), "b": cty.StringVal("B")}),
			cty.MapVal(map[string]cty.Value{"b": cty.StringVal("B")}),
		},
		{
			"DeleteList",
			Diff{
				DeleteChange{
					Path:     cty.IndexPath(cty.NumberIntVal(1)),
					OldValue: cty.StringVal("B"),
				},
			},
			cty.ListVal([]cty.Value{cty.StringVal("A"), cty.StringVal("B"), cty.StringVal("C")}),
			cty.ListVal([]cty.Value{cty.StringVal("A"), cty.StringVal("C")}),
		},
		{
			"DeleteTuple",
			Diff{
				DeleteChange{
					Path:     cty.IndexPath(cty.NumberIntVal(0)),
					OldValue: cty.StringVal("A"),
				},
			},
			cty.TupleVal([]cty.Value{cty.StringVal("A"), cty.StringVal("B")}),
			cty.TupleVal([]cty.Value{cty.StringVal("B")}),
		},

		// Insert
		{
			"InsertListEmpty",
			Diff{
				InsertChange{
					Path:        nil,
					NewValue:    cty.StringVal("a"),
					BeforeValue: cty.NullVal(cty.String),
				},
			},
			cty.ListValEmpty(cty.String),
			cty.ListVal([]cty.Value{cty.StringVal("a")}),
		},
		{
			"InsertList",
			Diff{
				InsertChange{
					Path:        nil,
					NewValue:    cty.StringVal("x"),
					BeforeValue: cty.StringVal("a"),
				},
			},
			cty.ListVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b"), cty.StringVal("a")}),
			cty.ListVal([]cty.Value{cty.StringVal("x"), cty.StringVal("a"), cty.StringVal("b"), cty.StringVal("a")}),
		},
		{
			"InsertTupleEmpty",
			Diff{
				InsertChange{
					Path:        nil,
					NewValue:    cty.StringVal("a"),
					BeforeValue: cty.NullVal(cty.String),
				},
			},
			cty.TupleVal(nil),
			cty.TupleVal([]cty.Value{cty.StringVal("a")}),
		},
		{
			"InsertList",
			Diff{
				InsertChange{
					Path:        nil,
					NewValue:    cty.StringVal("x"),
					BeforeValue: cty.StringVal("b"),
				},
			},
			cty.TupleVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b")}),
			cty.TupleVal([]cty.Value{cty.StringVal("a"), cty.StringVal("x"), cty.StringVal("b")}),
		},

		// Add
		{
			"AddSet",
			Diff{
				AddChange{
					Path:     nil,
					NewValue: cty.StringVal("a"),
				},
			},
			cty.SetValEmpty(cty.String),
			cty.SetVal([]cty.Value{cty.StringVal("a")}),
		},

		// Remove
		{
			"RemoveSet",
			Diff{
				RemoveChange{
					Path:     nil,
					OldValue: cty.StringVal("a"),
				},
			},
			cty.SetVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b")}),
			cty.SetVal([]cty.Value{cty.StringVal("b")}),
		},

		// Context
		{
			"Context",
			Diff{
				Context{
					Path:      cty.GetAttrPath("a"),
					WantValue: cty.StringVal("A"),
				},
			},
			cty.ObjectVal(map[string]cty.Value{"a": cty.StringVal("A")}),
			cty.ObjectVal(map[string]cty.Value{"a": cty.StringVal("A")}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.diff.Apply(tt.source)
			if err != nil {
				t.Fatalf("Apply() err = %v", err)
			}
			if !got.RawEquals(tt.want) {
				t.Errorf("Apply\nGot\n%#v\nWant\n%#v", got, tt.want)
			}
		})
	}
}
