package cty

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type capsuleTestType1Native struct {
	name string
}

type capsuleTestType2Native struct {
	name string
}

var capsuleTestType1 = Capsule(
	"capsule test type 1",
	reflect.TypeOf(capsuleTestType1Native{}),
)

var capsuleTestType2 = Capsule(
	"capsule test type 2",
	reflect.TypeOf(capsuleTestType2Native{}),
)

func TestCapsuleWithOps(t *testing.T) {
	var i = 0
	var i2 = 0
	var i3 = 1
	t.Run("with ops", func(t *testing.T) {
		ty := CapsuleWithOps("with ops", reflect.TypeOf(0), &CapsuleOps{
			GoString: func(v interface{}) string {
				iPtr := v.(*int)
				return fmt.Sprintf("test.WithOpsVal(%#v)", *iPtr)
			},
			TypeGoString: func(ty reflect.Type) string {
				return fmt.Sprintf("test.WithOps(%s)", ty)
			},
			Equals: func(a, b interface{}) Value {
				aPtr := a.(*int)
				bPtr := b.(*int)
				return BoolVal(*aPtr == *bPtr)
			},
			RawEquals: func(a, b interface{}) bool {
				aPtr := a.(*int)
				bPtr := b.(*int)
				return *aPtr == *bPtr
			},
		})
		v := CapsuleVal(ty, &i)
		v2 := CapsuleVal(ty, &i2)
		v3 := CapsuleVal(ty, &i3)

		got := map[string]interface{}{}
		got["GoString"] = v.GoString()
		got["TypeGoString"] = ty.GoString()
		got["Equals.Yes"] = v.Equals(v2)
		got["Equals.No"] = v.Equals(v3)

		want := map[string]interface{}{
			"GoString":     "test.WithOpsVal(0)",
			"TypeGoString": "test.WithOps(int)",
			"Equals.Yes":   True,
			"Equals.No":    False,
		}

		valCmp := cmp.Comparer(Value.RawEquals)
		if diff := cmp.Diff(want, got, valCmp); diff != "" {
			t.Errorf("wrong results\n%s", diff)
		}
	})
	t.Run("without ops", func(t *testing.T) {
		ty := Capsule("without ops", reflect.TypeOf(0))
		v := CapsuleVal(ty, &i)
		v2 := CapsuleVal(ty, &i2)

		got := map[string]interface{}{}
		got["GoString"] = v.GoString()
		got["TypeGoString"] = ty.GoString()
		got["Equals"] = v.Equals(v2)
		got["RawEquals"] = v.RawEquals(v2)

		want := map[string]interface{}{
			"GoString":     fmt.Sprintf(`cty.CapsuleVal(cty.Capsule("without ops", reflect.TypeOf(0)), (*int)(0x%x))`, &i),
			"TypeGoString": `cty.Capsule("without ops", reflect.TypeOf(0))`,
			"Equals":       False,
			"RawEquals":    false,
		}

		valCmp := cmp.Comparer(Value.RawEquals)
		if diff := cmp.Diff(want, got, valCmp); diff != "" {
			t.Errorf("wrong results\n%s", diff)
		}
	})

}

func TestCapsuleExtensionData(t *testing.T) {
	ty := CapsuleWithOps("with extension data", reflect.TypeOf(0), &CapsuleOps{
		ExtensionData: func(key interface{}) interface{} {
			switch key {
			// Note that this is a bad example of a key, just using a plain
			// string for easier testing. Real-world extension keys should
			// be named types belonging to a package in the application that
			// is defining them.
			case "hello":
				return "world"
			default:
				return nil
			}
		},
	})

	got := ty.CapsuleExtensionData("hello")
	want := interface{}("world")
	if got != want {
		t.Errorf("wrong result for 'hello'\ngot:  %#v\nwant: %#v", got, want)
	}

	got = ty.CapsuleExtensionData("nonexistent")
	want = nil
	if got != want {
		t.Errorf("wrong result for 'nonexistent'\ngot:  %#v\nwant: %#v", got, want)
	}

	ty2 := Capsule("without extension data", reflect.TypeOf(0))
	got = ty2.CapsuleExtensionData("hello")
	want = nil
	if got != want {
		t.Errorf("wrong result for 'hello' without extension data\ngot:  %#v\nwant: %#v", got, want)
	}

}
