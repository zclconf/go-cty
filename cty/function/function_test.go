package function

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestReturnTypeForValues(t *testing.T) {
	tests := []struct {
		Spec     *Spec
		Args     []cty.Value
		WantType cty.Type
		WantErr  bool
	}{
		{
			Spec: &Spec{
				Params: []Parameter{},
				Type:   StaticReturnType(cty.Number),
				Impl:   stubImpl,
			},
			Args:     []cty.Value{},
			WantType: cty.Number,
		},
		{
			Spec: &Spec{
				Params: []Parameter{},
				Type:   StaticReturnType(cty.Number),
				Impl:   stubImpl,
			},
			Args:    []cty.Value{cty.NumberIntVal(2)},
			WantErr: true,
		},
		{
			Spec: &Spec{
				Params: []Parameter{},
				Type:   StaticReturnType(cty.Number),
				Impl:   stubImpl,
			},
			Args:    []cty.Value{cty.UnknownVal(cty.Number)},
			WantErr: true,
		},
		{
			Spec: &Spec{
				Params: []Parameter{
					{
						Type: cty.Number,
					},
				},
				Type: StaticReturnType(cty.Number),
				Impl: stubImpl,
			},
			Args:     []cty.Value{cty.NumberIntVal(2)},
			WantType: cty.Number,
		},
		{
			Spec: &Spec{
				Params: []Parameter{
					{
						Type: cty.Number,
					},
				},
				Type: StaticReturnType(cty.Number),
				Impl: stubImpl,
			},
			Args:     []cty.Value{cty.UnknownVal(cty.Number)},
			WantType: cty.Number,
		},
		{
			Spec: &Spec{
				Params: []Parameter{
					{
						Type: cty.Number,
					},
				},
				Type: StaticReturnType(cty.Number),
				Impl: stubImpl,
			},
			Args:     []cty.Value{cty.DynamicVal},
			WantType: cty.DynamicPseudoType,
		},
		{
			Spec: &Spec{
				Params: []Parameter{
					{
						Type:             cty.Number,
						AllowDynamicType: true,
					},
				},
				Type: StaticReturnType(cty.Number),
				Impl: stubImpl,
			},
			Args:     []cty.Value{cty.DynamicVal},
			WantType: cty.Number,
		},
		{
			Spec: &Spec{
				Params: []Parameter{
					{
						Type:             cty.Number,
						AllowDynamicType: true,
					},
				},
				Type: StaticReturnType(cty.Number),
				Impl: stubImpl,
			},
			Args:    []cty.Value{cty.UnknownVal(cty.String)},
			WantErr: true,
		},
		{
			Spec: &Spec{
				Params: []Parameter{
					{
						Type:             cty.Number,
						AllowDynamicType: true,
					},
				},
				Type: StaticReturnType(cty.Number),
				Impl: stubImpl,
			},
			Args:    []cty.Value{cty.StringVal("hello")},
			WantErr: true,
		},
		{
			Spec: &Spec{
				Params: []Parameter{
					{
						Type: cty.List(cty.DynamicPseudoType),
					},
				},
				Type: func(args []cty.Value) (cty.Type, error) {
					ty := cty.Number
					for i, arg := range args {
						if arg.ContainsMarked() {
							return ty, fmt.Errorf("arg %d %#v contains marks", i, arg)
						}
					}
					return ty, nil
				},
				Impl: stubImpl,
			},
			Args: []cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("ok").Mark("marked"),
				}),
			},
			WantType: cty.Number,
		},
		{
			Spec: &Spec{
				Params: []Parameter{
					{
						Type: cty.List(cty.String),
					},
				},
				VarParam: &Parameter{
					Type: cty.List(cty.String),
				},
				Type: func(args []cty.Value) (cty.Type, error) {
					ty := cty.Number
					for i, arg := range args {
						if arg.ContainsMarked() {
							return ty, fmt.Errorf("arg %d %#v contains marks", i, arg)
						}
					}
					return ty, nil
				},
				Impl: stubImpl,
			},
			Args: []cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("one"),
				}),
				cty.ListVal([]cty.Value{
					cty.StringVal("two").Mark("marked"),
				}),
			},
			WantType: cty.Number,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			f := New(test.Spec)
			gotType, gotErr := f.ReturnTypeForValues(test.Args)

			if test.WantErr {
				if gotErr == nil {
					t.Errorf("succeeded with %#v; want error", gotType)
				}
			} else {
				if gotErr != nil {
					t.Fatalf("unexpected error\nspec: %#v\nargs: %#v\nerr:  %s\nwant: %#v", test.Spec, test.Args, gotErr, test.WantType)
				}

				if gotType == cty.NilType {
					t.Fatalf("returned type is invalid")
				}

				if !gotType.Equals(test.WantType) {
					t.Errorf("wrong return type\nspec: %#v\nargs: %#v\ngot:  %#v\nwant: %#v", test.Spec, test.Args, gotType, test.WantType)
				}
			}
		})
	}
}

func TestFunctionWithNewDescriptions(t *testing.T) {
	t.Run("no params", func(t *testing.T) {
		f1 := New(&Spec{
			Description: "old func",
			Params:      []Parameter{},
			Type:        stubType,
			Impl:        stubImpl,
		})
		f2 := f1.WithNewDescriptions(
			"new func",
			nil,
		)

		if got, want := f1.Description(), "old func"; got != want {
			t.Errorf("wrong original func description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f2.Description(), "new func"; got != want {
			t.Errorf("wrong updated func description\ngot:  %s\nwant: %s", got, want)
		}
	})
	t.Run("one pos param", func(t *testing.T) {
		f1 := New(&Spec{
			Description: "old func",
			Params: []Parameter{
				{
					Name:        "a",
					Description: "old a",
				},
			},
			Type: stubType,
			Impl: stubImpl,
		})
		f2 := f1.WithNewDescriptions(
			"new func",
			[]string{"new a"},
		)

		if got, want := f1.Description(), "old func"; got != want {
			t.Errorf("wrong original func description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f2.Description(), "new func"; got != want {
			t.Errorf("wrong updated func description\ngot:  %s\nwant: %s", got, want)
		}

		if got, want := len(f1.Params()), 1; got != want {
			t.Fatalf("wrong original param count\ngot:  %d\nwant: %d", got, want)
		}
		if got, want := len(f2.Params()), 1; got != want {
			t.Fatalf("wrong updated param count\ngot:  %d\nwant: %d", got, want)
		}
		if got, want := f1.Params()[0].Description, "old a"; got != want {
			t.Errorf("wrong original param a description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f2.Params()[0].Description, "new a"; got != want {
			t.Errorf("wrong updated param a description\ngot:  %s\nwant: %s", got, want)
		}
	})
	t.Run("two pos params", func(t *testing.T) {
		f1 := New(&Spec{
			Description: "old func",
			Params: []Parameter{
				{
					Name:        "a",
					Description: "old a",
				},
				{
					Name:        "b",
					Description: "old b",
				},
			},
			Type: stubType,
			Impl: stubImpl,
		})
		f2 := f1.WithNewDescriptions(
			"new func",
			[]string{"new a", "new b"},
		)

		if got, want := f1.Description(), "old func"; got != want {
			t.Errorf("wrong original func description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f2.Description(), "new func"; got != want {
			t.Errorf("wrong updated func description\ngot:  %s\nwant: %s", got, want)
		}

		if got, want := len(f1.Params()), 2; got != want {
			t.Fatalf("wrong original param count\ngot:  %d\nwant: %d", got, want)
		}
		if got, want := len(f2.Params()), 2; got != want {
			t.Fatalf("wrong updated param count\ngot:  %d\nwant: %d", got, want)
		}
		if got, want := f1.Params()[0].Description, "old a"; got != want {
			t.Errorf("wrong original param a description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f2.Params()[0].Description, "new a"; got != want {
			t.Errorf("wrong updated param a description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f1.Params()[1].Description, "old b"; got != want {
			t.Errorf("wrong original param b description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f2.Params()[1].Description, "new b"; got != want {
			t.Errorf("wrong updated param b description\ngot:  %s\nwant: %s", got, want)
		}
	})
	t.Run("varparam overridden", func(t *testing.T) {
		f1 := New(&Spec{
			Description: "old func",
			Params: []Parameter{
				{
					Name:        "a",
					Description: "old a",
				},
			},
			VarParam: &Parameter{
				Name:        "b",
				Description: "old b",
			},
			Type: stubType,
			Impl: stubImpl,
		})
		f2 := f1.WithNewDescriptions(
			"new func",
			[]string{"new a", "new b"},
		)

		if got, want := f1.Description(), "old func"; got != want {
			t.Errorf("wrong original func description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f2.Description(), "new func"; got != want {
			t.Errorf("wrong updated func description\ngot:  %s\nwant: %s", got, want)
		}

		if got, want := len(f1.Params()), 1; got != want {
			t.Fatalf("wrong original param count\ngot:  %d\nwant: %d", got, want)
		}
		if got, want := len(f2.Params()), 1; got != want {
			t.Fatalf("wrong updated param count\ngot:  %d\nwant: %d", got, want)
		}
		if got, want := f1.Params()[0].Description, "old a"; got != want {
			t.Errorf("wrong original param a description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f2.Params()[0].Description, "new a"; got != want {
			t.Errorf("wrong updated param a description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f1.VarParam().Description, "old b"; got != want {
			t.Errorf("wrong original param b description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f2.VarParam().Description, "new b"; got != want {
			t.Errorf("wrong updated param b description\ngot:  %s\nwant: %s", got, want)
		}
	})
	t.Run("varparam not overridden", func(t *testing.T) {
		f1 := New(&Spec{
			Description: "old func",
			Params: []Parameter{
				{
					Name:        "a",
					Description: "old a",
				},
			},
			VarParam: &Parameter{
				Name:        "b",
				Description: "old b",
			},
			Type: stubType,
			Impl: stubImpl,
		})
		f2 := f1.WithNewDescriptions(
			"new func",
			[]string{"new a"},
		)

		if got, want := f1.Description(), "old func"; got != want {
			t.Errorf("wrong original func description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f2.Description(), "new func"; got != want {
			t.Errorf("wrong updated func description\ngot:  %s\nwant: %s", got, want)
		}

		if got, want := len(f1.Params()), 1; got != want {
			t.Fatalf("wrong original param count\ngot:  %d\nwant: %d", got, want)
		}
		if got, want := len(f2.Params()), 1; got != want {
			t.Fatalf("wrong updated param count\ngot:  %d\nwant: %d", got, want)
		}
		if got, want := f1.Params()[0].Description, "old a"; got != want {
			t.Errorf("wrong original param a description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f2.Params()[0].Description, "new a"; got != want {
			t.Errorf("wrong updated param a description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f1.VarParam().Description, "old b"; got != want {
			t.Errorf("wrong original param b description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f2.VarParam().Description, "old b"; got != want {
			// This is the one case where we allow the caller to leave one of
			// the param descriptions unchanged, because we want to allow
			// a function to grow a variadic parameter later without it being
			// a breaking change for existing callers that might be overriding
			// descriptions.
			t.Errorf("wrong updated param b description\ngot:  %s\nwant: %s", got, want)
		}
	})
	t.Run("solo varparam overridden", func(t *testing.T) {
		f1 := New(&Spec{
			Description: "old func",
			VarParam: &Parameter{
				Name:        "a",
				Description: "old a",
			},
			Type: stubType,
			Impl: stubImpl,
		})
		f2 := f1.WithNewDescriptions(
			"new func",
			[]string{"new a"},
		)

		if got, want := f1.Description(), "old func"; got != want {
			t.Errorf("wrong original func description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f2.Description(), "new func"; got != want {
			t.Errorf("wrong updated func description\ngot:  %s\nwant: %s", got, want)
		}

		if got, want := len(f1.Params()), 0; got != want {
			t.Fatalf("wrong original param count\ngot:  %d\nwant: %d", got, want)
		}
		if got, want := len(f2.Params()), 0; got != want {
			t.Fatalf("wrong updated param count\ngot:  %d\nwant: %d", got, want)
		}
		if got, want := f1.VarParam().Description, "old a"; got != want {
			t.Errorf("wrong original param b description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f2.VarParam().Description, "new a"; got != want {
			t.Errorf("wrong updated param b description\ngot:  %s\nwant: %s", got, want)
		}
	})
	t.Run("solo varparam not overridden", func(t *testing.T) {
		f1 := New(&Spec{
			Description: "old func",
			VarParam: &Parameter{
				Name:        "a",
				Description: "old a",
			},
			Type: stubType,
			Impl: stubImpl,
		})
		f2 := f1.WithNewDescriptions(
			"new func",
			nil,
		)

		if got, want := f1.Description(), "old func"; got != want {
			t.Errorf("wrong original func description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f2.Description(), "new func"; got != want {
			t.Errorf("wrong updated func description\ngot:  %s\nwant: %s", got, want)
		}

		if got, want := len(f1.Params()), 0; got != want {
			t.Fatalf("wrong original param count\ngot:  %d\nwant: %d", got, want)
		}
		if got, want := len(f2.Params()), 0; got != want {
			t.Fatalf("wrong updated param count\ngot:  %d\nwant: %d", got, want)
		}
		if got, want := f1.VarParam().Description, "old a"; got != want {
			t.Errorf("wrong original param b description\ngot:  %s\nwant: %s", got, want)
		}
		if got, want := f2.VarParam().Description, "old a"; got != want {
			// This is the one case where we allow the caller to leave one of
			// the param descriptions unchanged, because we want to allow
			// a function to grow a variadic parameter later without it being
			// a breaking change for existing callers that might be overriding
			// descriptions.
			t.Errorf("wrong updated param b description\ngot:  %s\nwant: %s", got, want)
		}
	})
}

func stubType([]cty.Value) (cty.Type, error) {
	return cty.NilType, fmt.Errorf("should not be called")
}

func stubImpl([]cty.Value, cty.Type) (cty.Value, error) {
	return cty.NilVal, fmt.Errorf("should not be called")
}
