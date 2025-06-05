//go:build go1.23

package cty_test

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
)

func ExampleGoIter() {
	// Test that iterating over a list works
	listVal := cty.ListVal([]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(2)})
	for key, val := range cty.GoIter(listVal) {
		fmt.Println(key.GoString(), val.GoString())
	}
	keyVal := cty.MapVal(map[string]cty.Value{
		"a": cty.NumberIntVal(1),
		"b": cty.NumberIntVal(2),
	})
	for key, val := range cty.GoIter(keyVal) {
		fmt.Println(key.GoString(), val.GoString())
	}
	// Output:
	// cty.NumberIntVal(0) cty.NumberIntVal(1)
	// cty.NumberIntVal(1) cty.NumberIntVal(2)
	// cty.StringVal("a") cty.NumberIntVal(1)
	// cty.StringVal("b") cty.NumberIntVal(2)
}
