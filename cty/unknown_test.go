package cty

import (
	"testing"
)

func TestDynamicValComparable(t *testing.T) {
	// For better or worse, the initial definition of cty.DynamicVal was
	// comparable and some callers used expressions like "val == cty.DynamicVal"
	// to test for it, so this is now effectively a contract.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %s", r)
		}
	}()
	if DynamicVal == DynamicVal {
		// ok
	} else {
		t.Error("DynamicVal is not equal to itself, but should be")
	}
	if UnknownVal(String) == DynamicVal {
		t.Error("unknown string is equal to dynamic val, but shouldn't be")
	}
}
