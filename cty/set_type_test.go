package cty

import (
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSetOperations(t *testing.T) {
	// This test is for the mechanisms that allow a calling application to
	// implement set operations using the underlying set.Set type. This is
	// not expected to be a common case but is useful, for example, for
	// implementing the set-related functions in function/stdlib .

	s1 := SetVal([]Value{
		StringVal("a"),
		StringVal("b"),
		StringVal("c"),
	})
	s2 := SetVal([]Value{
		StringVal("c"),
		StringVal("d"),
		StringVal("e"),
	})

	s1r := s1.AsValueSet()
	s2r := s2.AsValueSet()
	s3r := s1r.Union(s2r)

	s3 := SetValFromValueSet(s3r)

	if got, want := s3.LengthInt(), 5; got != want {
		t.Errorf("wrong length %d; want %d", got, want)
	}

	for _, wantStr := range []string{"a", "b", "c", "d", "e"} {
		if got, want := s3.HasElement(StringVal(wantStr)), True; got != want {
			t.Errorf("missing element %q", wantStr)
		}
	}
}

func TestSetOfCapsuleType(t *testing.T) {
	type capsuleTypeForSetTests struct {
		name string
	}

	encapsulatedNames := func(vals []Value) []string {
		if len(vals) == 0 {
			return nil
		}
		ret := make([]string, len(vals))
		for i, v := range vals {
			ret[i] = v.EncapsulatedValue().(*capsuleTypeForSetTests).name
		}
		sort.Strings(ret)
		return ret
	}

	typeWithHash := CapsuleWithOps("with hash function", reflect.TypeOf(capsuleTypeForSetTests{}), &CapsuleOps{
		RawEquals: func(a, b interface{}) bool {
			return a.(*capsuleTypeForSetTests).name == b.(*capsuleTypeForSetTests).name
		},
		HashKey: func(v interface{}) string {
			return v.(*capsuleTypeForSetTests).name
		},
	})
	typeWithoutHash := CapsuleWithOps("without hash function", reflect.TypeOf(capsuleTypeForSetTests{}), &CapsuleOps{
		RawEquals: func(a, b interface{}) bool {
			return a.(*capsuleTypeForSetTests).name == b.(*capsuleTypeForSetTests).name
		},
	})
	typeWithoutEquals := Capsule("without hash function", reflect.TypeOf(capsuleTypeForSetTests{}))

	t.Run("with hash", func(t *testing.T) {
		// When we provide a hashing function the set implementation can
		// optimize its internal storage by spreading values over multiple
		// smaller buckets.
		v := SetVal([]Value{
			CapsuleVal(typeWithHash, &capsuleTypeForSetTests{"a"}),
			CapsuleVal(typeWithHash, &capsuleTypeForSetTests{"b"}),
			CapsuleVal(typeWithHash, &capsuleTypeForSetTests{"a"}),
			CapsuleVal(typeWithHash, &capsuleTypeForSetTests{"c"}),
		})
		got := encapsulatedNames(v.AsValueSlice())
		want := []string{"a", "b", "c"}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("wrong element names\n%s", diff)
		}
	})
	t.Run("without hash", func(t *testing.T) {
		// When we don't provide a hashing function the outward behavior
		// should still be identical but the internal storage won't be
		// so efficient, due to everything living in one big bucket and
		// so we have to scan over all values to test if a particular
		// element is present.
		v := SetVal([]Value{
			CapsuleVal(typeWithoutHash, &capsuleTypeForSetTests{"a"}),
			CapsuleVal(typeWithoutHash, &capsuleTypeForSetTests{"b"}),
			CapsuleVal(typeWithoutHash, &capsuleTypeForSetTests{"a"}),
			CapsuleVal(typeWithoutHash, &capsuleTypeForSetTests{"c"}),
		})
		got := encapsulatedNames(v.AsValueSlice())
		want := []string{"a", "b", "c"}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("wrong element names\n%s", diff)
		}
	})
	t.Run("without equals", func(t *testing.T) {
		// When we don't even have an equals function we can still store
		// values in the set but we will use the default capsule type
		// behavior of comparing by pointer equality. That means that
		// the name field doesn't coalesce anymore, but two instances
		// of this same d should.
		d := &capsuleTypeForSetTests{"d"}
		v := SetVal([]Value{
			CapsuleVal(typeWithoutEquals, &capsuleTypeForSetTests{"a"}),
			CapsuleVal(typeWithoutEquals, &capsuleTypeForSetTests{"b"}),
			CapsuleVal(typeWithoutEquals, d),
			CapsuleVal(typeWithoutEquals, &capsuleTypeForSetTests{"a"}),
			CapsuleVal(typeWithoutEquals, &capsuleTypeForSetTests{"c"}),
			CapsuleVal(typeWithoutEquals, d),
		})
		got := encapsulatedNames(v.AsValueSlice())
		want := []string{"a", "a", "b", "c", "d"}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("wrong element names\n%s", diff)
		}
	})

}
