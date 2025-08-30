package cty_test

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/ctymarks"
)

func TestValueWrangleMarksDeep(t *testing.T) {
	tests := map[string]struct {
		input   cty.Value
		funcs   []cty.WrangleFunc
		want    cty.Value
		wantErr string
	}{
		"null with no marks nor funcs": {
			input: cty.NullVal(cty.DynamicPseudoType),
			funcs: nil,
			want:  cty.NullVal(cty.DynamicPseudoType),
		},
		"null with no marks and unused func": {
			input: cty.NullVal(cty.DynamicPseudoType),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					return nil, fmt.Errorf("this error should not be observed")
				},
			},
			want: cty.NullVal(cty.DynamicPseudoType),
		},
		"null with mark but no funcs": {
			input: cty.NullVal(cty.DynamicPseudoType).Mark("irrelevant"),
			funcs: nil,
			want:  cty.NullVal(cty.DynamicPseudoType).Mark("irrelevant"),
		},
		"null with mark that is unaffected by func": {
			input: cty.NullVal(cty.DynamicPseudoType).Mark("irrelevant"),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					if mark != "irrelevant" || len(path) != 0 {
						return nil, fmt.Errorf("this error should not be observed")
					}
					return nil, nil
				},
			},
			want: cty.NullVal(cty.DynamicPseudoType).Mark("irrelevant"),
		},
		"null with mark and func that's blocked by earlier func": {
			input: cty.NullVal(cty.DynamicPseudoType).Mark("maybe bad"),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					return ctymarks.WrangleKeep, nil
				},
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					return nil, fmt.Errorf("this error should not be observed")
				},
			},
			want: cty.NullVal(cty.DynamicPseudoType).Mark("maybe bad"),
		},
		"null with mark and func that's not blocked by earlier func": {
			input: cty.NullVal(cty.DynamicPseudoType).Mark("maybe bad"),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					return nil, nil
				},
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					return ctymarks.WrangleDrop, fmt.Errorf("found mark %q at path %#v", mark, path)
				},
			},
			want:    cty.NullVal(cty.DynamicPseudoType),
			wantErr: `found mark "maybe bad" at path cty.Path(nil)`,
		},
		"null with marks, one of which is dropped": {
			input: cty.NullVal(cty.DynamicPseudoType).Mark("keep").Mark("drop"),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					if mark == "drop" {
						return ctymarks.WrangleDrop, nil
					}
					return nil, nil
				},
			},
			want: cty.NullVal(cty.DynamicPseudoType).Mark("keep"),
		},
		"null with marks, one of which is replaced": {
			input: cty.NullVal(cty.DynamicPseudoType).Mark("keep").Mark("drop"),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					if mark == "drop" {
						return ctymarks.WrangleReplace("replacement"), nil
					}
					return nil, nil
				},
			},
			want: cty.NullVal(cty.DynamicPseudoType).Mark("keep").Mark("replacement"),
		},
		"null with a mark that causes an error": {
			input: cty.NullVal(cty.DynamicPseudoType).Mark("bad").Mark("irrelevant"),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					if mark == "bad" {
						return ctymarks.WrangleDrop, fmt.Errorf("found mark %q at path %#v", mark, path)
					}
					return nil, nil
				},
			},
			want:    cty.NullVal(cty.DynamicPseudoType).Mark("irrelevant"),
			wantErr: `found mark "bad" at path cty.Path(nil)`,
		},

		// Sets are not really any different than primitive values for the
		// sake of this function, because they can't contain any nested values
		// that are individually marked. This single test is therefore here
		// just to check that we don't do anything weird with a set.
		"set with marks, one of which is dropped": {
			input: cty.SetVal([]cty.Value{cty.True}).Mark("drop").Mark("keep"),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					if mark == "drop" {
						return ctymarks.WrangleDrop, nil
					}
					return nil, nil
				},
			},
			want: cty.SetVal([]cty.Value{cty.True}).Mark("keep"),
		},

		"list with no marks and inert wrangle func": {
			input: cty.ListVal([]cty.Value{
				cty.StringVal("unmarked 1"),
				cty.StringVal("unmarked 2"),
				cty.StringVal("unmarked 3"),
			}),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					return nil, fmt.Errorf("this error should not be observed")
				},
			},
			want: cty.ListVal([]cty.Value{
				cty.StringVal("unmarked 1"),
				cty.StringVal("unmarked 2"),
				cty.StringVal("unmarked 3"),
			}),
		},
		"list with nested marks, one of which is dropped": {
			input: cty.ListVal([]cty.Value{
				cty.StringVal("unmarked"),
				cty.StringVal("marked 1").Mark("drop"),
				cty.StringVal("marked 2").Mark("drop").Mark("keep"),
				cty.StringVal("marked 3").Mark("keep"),
			}),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					if mark == "drop" {
						return ctymarks.WrangleDrop, nil
					}
					return nil, nil
				},
			},
			want: cty.ListVal([]cty.Value{
				cty.StringVal("unmarked"),
				cty.StringVal("marked 1"),
				cty.StringVal("marked 2").Mark("keep"),
				cty.StringVal("marked 3").Mark("keep"),
			}),
		},
		"tuple with nested marks, one of which is dropped": {
			input: cty.TupleVal([]cty.Value{
				cty.True,
				cty.StringVal("marked 1").Mark("drop"),
				cty.StringVal("marked 2").Mark("drop").Mark("keep"),
				cty.StringVal("marked 3").Mark("keep"),
			}),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					if mark == "drop" {
						return ctymarks.WrangleDrop, nil
					}
					return nil, nil
				},
			},
			want: cty.TupleVal([]cty.Value{
				cty.True,
				cty.StringVal("marked 1"),
				cty.StringVal("marked 2").Mark("keep"),
				cty.StringVal("marked 3").Mark("keep"),
			}),
		},
		"list with nested marks, one of which is expanded": {
			input: cty.ListVal([]cty.Value{
				cty.StringVal("unmarked"),
				cty.StringVal("marked 1").Mark("expand"),
				cty.StringVal("marked 2").Mark("expand").Mark("keep"),
				cty.StringVal("marked 3").Mark("keep"),
			}),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					if mark == "expand" {
						return ctymarks.WrangleExpand, nil
					}
					return nil, nil
				},
			},
			want: cty.ListVal([]cty.Value{
				cty.StringVal("unmarked"),
				cty.StringVal("marked 1").Mark("expand"),
				cty.StringVal("marked 2").Mark("expand").Mark("keep"),
				cty.StringVal("marked 3").Mark("keep"),
			}).Mark("expand"),
		},
		"list with nested mark that causes error": {
			input: cty.ListVal([]cty.Value{
				cty.StringVal("unmarked 1"),
				cty.StringVal("marked").Mark("bad"),
				cty.StringVal("unmarked 2"),
			}),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					if mark == "bad" {
						return nil, fmt.Errorf("found mark %q at path %#v", mark, path)
					}
					return nil, nil
				},
			},
			want: cty.ListVal([]cty.Value{
				cty.StringVal("unmarked 1"),
				cty.StringVal("marked").Mark("bad"),
				cty.StringVal("unmarked 2"),
			}),
			wantErr: `found mark "bad" at path cty.Path{cty.IndexStep{Key:cty.NumberIntVal(1)}}`,
		},
		"list with nested marks that cause error": {
			input: cty.ListVal([]cty.Value{
				cty.StringVal("unmarked 1"),
				cty.StringVal("marked 1").Mark("bad"),
				cty.StringVal("marked 2").Mark("bad"),
				cty.StringVal("unmarked 2"),
			}),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					if mark == "bad" {
						return nil, fmt.Errorf("found mark %q at path %#v", mark, path)
					}
					return nil, nil
				},
			},
			want: cty.ListVal([]cty.Value{
				cty.StringVal("unmarked 1"),
				cty.StringVal("marked 1").Mark("bad"),
				cty.StringVal("marked 2").Mark("bad"),
				cty.StringVal("unmarked 2"),
			}),
			wantErr: `found mark "bad" at path cty.Path{cty.IndexStep{Key:cty.NumberIntVal(1)}}
found mark "bad" at path cty.Path{cty.IndexStep{Key:cty.NumberIntVal(2)}}`,
		},

		"object with no marks and inert wrangle func": {
			input: cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal("Bob"),
				"age":  cty.NumberIntVal(84),
				"friends": cty.ListVal([]cty.Value{
					cty.StringVal("Harpreet"),
					cty.StringVal("Amanda"),
				}),
			}),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					return nil, fmt.Errorf("this error should not be observed")
				},
			},
			want: cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal("Bob"),
				"age":  cty.NumberIntVal(84),
				"friends": cty.ListVal([]cty.Value{
					cty.StringVal("Harpreet"),
					cty.StringVal("Amanda"),
				}),
			}),
		},
		"object with marks, one of which is dropped": {
			input: cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal("Bob"),
				"age":  cty.NumberIntVal(84).Mark("drop").Mark("keep"),
				"friends": cty.ListVal([]cty.Value{
					cty.StringVal("Harpreet").Mark("drop"),
					cty.StringVal("Amanda").Mark("keep"),
				}),
			}),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					if mark == "drop" {
						return ctymarks.WrangleDrop, nil
					}
					return nil, nil
				},
			},
			want: cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal("Bob"),
				"age":  cty.NumberIntVal(84).Mark("keep"),
				"friends": cty.ListVal([]cty.Value{
					cty.StringVal("Harpreet"),
					cty.StringVal("Amanda").Mark("keep"),
				}),
			}),
		},
		"object with marks, one of which is expanded": {
			input: cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal("Bob"),
				"age":  cty.NumberIntVal(84).Mark("keep"),
				"friends": cty.ListVal([]cty.Value{
					cty.StringVal("Harpreet").Mark("expand"),
					cty.StringVal("Amanda").Mark("keep"),
				}),
			}).Mark("keep"),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					if mark == "expand" {
						return ctymarks.WrangleExpand, nil
					}
					return nil, nil
				},
			},
			want: cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal("Bob"),
				"age":  cty.NumberIntVal(84).Mark("keep"),
				"friends": cty.ListVal([]cty.Value{
					cty.StringVal("Harpreet").Mark("expand"),
					cty.StringVal("Amanda").Mark("keep"),
				}),
			}).Mark("keep").Mark("expand"),
		},
		"map with no marks and inert wrangle func": {
			input: cty.MapVal(map[string]cty.Value{
				"foo": cty.StringVal("bar"),
				"baz": cty.StringVal("beep"),
			}),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					return nil, fmt.Errorf("this error should not be observed")
				},
			},
			want: cty.MapVal(map[string]cty.Value{
				"foo": cty.StringVal("bar"),
				"baz": cty.StringVal("beep"),
			}),
		},
		"map with marks, one of which is dropped": {
			input: cty.MapVal(map[string]cty.Value{
				"unmarked 1": cty.StringVal("unmarked"),
				"marked 1":   cty.StringVal("marked").Mark("keep"),
				"marked 2":   cty.StringVal("marked").Mark("keep").Mark("drop"),
				"marked 3":   cty.StringVal("marked").Mark("drop"),
				"unmarked 2": cty.StringVal("unmarked"),
			}),
			funcs: []cty.WrangleFunc{
				func(mark any, path cty.Path) (ctymarks.WrangleAction, error) {
					if mark == "drop" {
						return ctymarks.WrangleDrop, nil
					}
					return nil, nil
				},
			},
			want: cty.MapVal(map[string]cty.Value{
				"unmarked 1": cty.StringVal("unmarked"),
				"marked 1":   cty.StringVal("marked").Mark("keep"),
				"marked 2":   cty.StringVal("marked").Mark("keep"),
				"marked 3":   cty.StringVal("marked"),
				"unmarked 2": cty.StringVal("unmarked"),
			}),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotErr := test.input.WrangleMarksDeep(test.funcs...)
			if wantErr := test.wantErr; wantErr != "" {
				if gotErr == nil {
					t.Errorf("unexpected success\nwant error: %s", wantErr)
				} else if gotErr := gotErr.Error(); gotErr != wantErr {
					t.Errorf("wrong error\ngot:  %s\nwant: %s", gotErr, wantErr)
				}
			} else if gotErr != nil {
				t.Errorf("unexpected error: %s", gotErr)
			}

			if want := test.want; !want.RawEquals(got) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, want)
			}
		})
	}
}
