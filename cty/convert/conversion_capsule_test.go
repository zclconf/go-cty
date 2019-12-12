package convert

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestConvertCapsuleType(t *testing.T) {
	capTy := cty.CapsuleWithOps("test thingy", reflect.TypeOf(""), &cty.CapsuleOps{
		GoString: func(rawV interface{}) string {
			vPtr := rawV.(*string)
			return fmt.Sprintf("capTy(%q)", *vPtr)
		},
		TypeGoString: func(ty reflect.Type) string {
			return "capTy"
		},
		RawEquals: func(a, b interface{}) bool {
			aPtr := a.(*string)
			bPtr := b.(*string)
			return *aPtr == *bPtr
		},
		ConversionFrom: func(srcTy cty.Type) func(interface{}, cty.Path) (cty.Value, error) {
			if !srcTy.Equals(cty.String) {
				return nil
			}
			return func(rawV interface{}, path cty.Path) (cty.Value, error) {
				vPtr := rawV.(*string)
				return cty.StringVal(*vPtr), nil
			}
		},
		ConversionTo: func(dstTy cty.Type) func(cty.Value, cty.Path) (interface{}, error) {
			if !dstTy.Equals(cty.String) {
				return nil
			}
			return func(from cty.Value, path cty.Path) (interface{}, error) {
				s := from.AsString()
				return &s, nil
			}
		},
	})

	capVal := func(s string) cty.Value {
		return cty.CapsuleVal(capTy, &s)
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
			WantErr: `test thingy required`,
		},
		{
			From:    capVal("hello"),
			To:      cty.Bool,
			WantErr: `bool required`,
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
			WantErr: `test thingy required`,
		},
		{
			From:    cty.NullVal(cty.Bool),
			To:      capTy,
			WantErr: `test thingy required`,
		},
		{
			From:    cty.UnknownVal(capTy),
			To:      cty.Bool,
			WantErr: `bool required`,
		},
		{
			From:    cty.NullVal(capTy),
			To:      cty.Bool,
			WantErr: `bool required`,
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
