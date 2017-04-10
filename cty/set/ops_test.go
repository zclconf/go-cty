package set

import (
	"reflect"
	"sort"
	"testing"
)

// TestBasicSetOps tests the fundamental operations, whose implementations operate
// directly on the underlying data structure. The remaining operations are implemented
// in terms of these.
func TestBasicSetOps(t *testing.T) {
	s := NewSet(testRules{})
	want := map[int][]interface{}{}
	if !reflect.DeepEqual(s.vals, want) {
		t.Fatalf("new set has unexpected contents %#v; want %#v", s.vals, want)
	}
	s.Add(1)
	want[1] = []interface{}{1}
	if !reflect.DeepEqual(s.vals, want) {
		t.Fatalf("after s.Add(1) set has unexpected contents %#v; want %#v", s.vals, want)
	}
	if !s.Has(1) {
		t.Fatalf("s.Has(1) returned false; want true")
	}
	s.Add(2)
	want[2] = []interface{}{2}
	if !reflect.DeepEqual(s.vals, want) {
		t.Fatalf("after s.Add(2) set has unexpected contents %#v; want %#v", s.vals, want)
	}
	if !s.Has(2) {
		t.Fatalf("s.Has(2) returned false; want true")
	}

	// Our testRules cause 17 and 33 to return the same hash value as 1, so we can use this
	// to test the situation where multiple values are in a bucket.
	if s.Has(17) {
		t.Fatalf("s.Has(17) returned true; want false")
	}
	s.Add(17)
	s.Add(33)
	want[1] = append(want[1], 17, 33)
	if !reflect.DeepEqual(s.vals, want) {
		t.Fatalf("after s.Add(17) and s.Add(33) set has unexpected contents %#v; want %#v", s.vals, want)
	}
	if !s.Has(17) {
		t.Fatalf("s.Has(17) returned false; want true")
	}
	if !s.Has(33) {
		t.Fatalf("s.Has(33) returned false; want true")
	}

	vals := make([]int, 0)
	s.EachValue(func(v interface{}) {
		vals = append(vals, v.(int))
	})
	sort.Ints(vals)
	if want := []int{1, 2, 17, 33}; !reflect.DeepEqual(vals, want) {
		t.Fatalf("wrong values from EachValue %#v; want %#v", vals, want)
	}

	s.Remove(2)
	delete(want, 2)
	if !reflect.DeepEqual(s.vals, want) {
		t.Fatalf("after s.Remove(2) set has unexpected contents %#v; want %#v", s.vals, want)
	}

	s.Remove(17)
	want[1] = []interface{}{1, 33}
	if !reflect.DeepEqual(s.vals, want) {
		t.Fatalf("after s.Remove(17) set has unexpected contents %#v; want %#v", s.vals, want)
	}

	s.Remove(1)
	want[1] = []interface{}{33}
	if !reflect.DeepEqual(s.vals, want) {
		t.Fatalf("after s.Remove(1) set has unexpected contents %#v; want %#v", s.vals, want)
	}

	s.Remove(33)
	delete(want, 1)
	if !reflect.DeepEqual(s.vals, want) {
		t.Fatalf("after s.Remove(33) set has unexpected contents %#v; want %#v", s.vals, want)
	}

	vals = make([]int, 0)
	s.EachValue(func(v interface{}) {
		vals = append(vals, v.(int))
	})
	if len(vals) > 0 {
		t.Fatalf("s.EachValue produced values %#v; want no calls", vals)
	}
}
