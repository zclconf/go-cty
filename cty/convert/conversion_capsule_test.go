package convert

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestConvertCapsuleType(t *testing.T) {
	capTy := cty.CapsuleWithOps("test thingy", reflect.TypeOf(""), &cty.CapsuleOps{
		GoString: func(rawV any) string {
			vPtr := rawV.(*string)
			return fmt.Sprintf("capTy(%q)", *vPtr)
		},
		TypeGoString: func(ty reflect.Type) string {
			return "capTy"
		},
		RawEquals: func(a, b any) bool {
			aPtr := a.(*string)
			bPtr := b.(*string)
			return *aPtr == *bPtr
		},
		ConversionFrom: func(srcTy cty.Type) func(any, cty.Path) (cty.Value, error) {
			if !srcTy.Equals(cty.String) {
				return nil
			}
			return func(rawV any, path cty.Path) (cty.Value, error) {
				vPtr := rawV.(*string)
				return cty.StringVal(*vPtr), nil
			}
		},
		ConversionTo: func(dstTy cty.Type) func(cty.Value, cty.Path) (any, error) {
			if !dstTy.Equals(cty.String) {
				return nil
			}
			return func(from cty.Value, path cty.Path) (any, error) {
				s := from.AsString()
				return &s, nil
			}
		},
	})

	capVal := func(s string) cty.Value {
		return cty.CapsuleVal(capTy, &s)
	}

	capIntTy := cty.CapsuleWithOps("int test thingy", reflect.TypeOf(0), &cty.CapsuleOps{
		ConversionFrom: func(src cty.Type) func(any, cty.Path) (cty.Value, error) {
			if src.Equals(capTy) {
				return func(v any, p cty.Path) (cty.Value, error) {
					return capVal(fmt.Sprintf("%d", *(v.(*int)))), nil
				}
			}
			return nil
		},
	})
	capIntVal := func(i int) cty.Value {
		return cty.CapsuleVal(capIntTy, &i)
	}

	tests := []struct {
		From    cty.Value
		To      cty.Type
		Want    cty.Value
		WantErr string
	}{
		{
			From: capVal("hello"),
			To:   cty.String,
			Want: cty.StringVal("hello"),
		},
		{
			From: cty.StringVal("hello"),
			To:   capTy,
			Want: capVal("hello"),
		},
		{
			From:    cty.True,
			To:      capTy,
			WantErr: `test thingy required, but have bool`,
		},
		{
			From:    capVal("hello"),
			To:      cty.Bool,
			WantErr: `bool required, but have test thingy`,
		},
		{
			From: cty.UnknownVal(capTy),
			To:   cty.String,
			Want: cty.UnknownVal(cty.String),
		},
		{
			From: cty.NullVal(capTy),
			To:   cty.String,
			Want: cty.NullVal(cty.String),
		},
		{
			From:    cty.UnknownVal(cty.Bool),
			To:      capTy,
			WantErr: `test thingy required, but have bool`,
		},
		{
			From:    cty.NullVal(cty.Bool),
			To:      capTy,
			WantErr: `test thingy required, but have bool`,
		},
		{
			From:    cty.UnknownVal(capTy),
			To:      cty.Bool,
			WantErr: `bool required, but have test thingy`,
		},
		{
			From:    cty.NullVal(capTy),
			To:      cty.Bool,
			WantErr: `bool required, but have test thingy`,
		},
		{
			From: capIntVal(42),
			To:   capTy,
			Want: capVal("42"),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v to %#v", test.From, test.To), func(t *testing.T) {
			got, err := Convert(test.From, test.To)

			if test.WantErr == "" {
				if err != nil {
					t.Fatalf("wrong error\nwant: <no error>\ngot:  %s", err)
				}
				if !test.Want.RawEquals(got) {
					t.Errorf("wrong result\nwant: %#v\ngot:  %#v", got, test.Want)
				}
			} else {
				if err == nil {
					t.Fatalf("wrong error\nwant: %s\ngot:  <no error>", test.WantErr)
				}
				if got, want := err.Error(), test.WantErr; got != want {
					t.Errorf("wrong error\nwant: %s\ngot:  %s", got, want)
				}
			}
		})
	}
}
