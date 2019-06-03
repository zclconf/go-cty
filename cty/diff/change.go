package diff

import (
	"errors"
	"fmt"

	"github.com/zclconf/go-cty/cty"
)

// Change is an abstract type representing a single change operation as
// part of a Diff.
//
// Change is a closed interface, meaning that the only permitted
// implementations are those within this package.
type Change interface {
	changeSigil() changeImpl
	apply(val cty.Value) (cty.Value, error)
}

// Embed changeImpl into a struct to make it a Change implementation
type changeImpl struct{}

func (c changeImpl) changeSigil() changeImpl {
	return c
}

// ReplaceChange is a Change implementation that represents replacing an
// existing value with an entirely new value.
//
// When adding a new element to a map value, this change type should be used
// with OldValue set to a null value of the appropriate type.
type ReplaceChange struct {
	changeImpl
	Path     cty.Path
	OldValue cty.Value
	NewValue cty.Value
}

func (c ReplaceChange) apply(val cty.Value) (cty.Value, error) {
	if len(c.Path) == 0 {
		// Empty path, replace entire value.
		if !val.RawEquals(c.OldValue) {
			return cty.NilVal, errors.New("existing value does not match")
		}
		return c.NewValue, nil
	}
	parent, _ := c.Path[:len(c.Path)-1].Apply(val)
	if !c.OldValue.IsNull() || !parent.Type().IsMapType() {
		// Compare existing.
		existing, err := c.Path.Apply(val)
		if err != nil {
			return cty.NilVal, c.Path.NewErrorf("path does not exist in value")
		}
		if !existing.RawEquals(c.OldValue) {
			return cty.NilVal, c.Path.NewErrorf("existing value does not match")
		}
	}
	key := c.Path[len(c.Path)-1]
	ty := parent.Type()
	switch {
	case ty.IsObjectType():
		kv := parent.AsValueMap()
		kv[key.(cty.GetAttrStep).Name] = c.NewValue
		return cty.ObjectVal(kv), nil
	case ty.IsMapType():
		kv := parent.AsValueMap()
		if kv == nil {
			kv = make(map[string]cty.Value)
		}
		kv[key.(cty.IndexStep).Key.AsString()] = c.NewValue
		return cty.MapVal(kv), nil
	case ty.IsListType():
		var vv []cty.Value
		idx := key.(cty.IndexStep).Key
		for it := parent.ElementIterator(); it.Next(); {
			i, ev := it.Element()
			if i.RawEquals(idx) {
				vv = append(vv, c.NewValue)
				continue
			}
			vv = append(vv, ev)
		}
		return cty.ListVal(vv), nil
	case ty.IsTupleType():
		var vv []cty.Value
		idx := key.(cty.IndexStep).Key
		for it := parent.ElementIterator(); it.Next(); {
			i, ev := it.Element()
			if i.RawEquals(idx) {
				vv = append(vv, c.NewValue)
				continue
			}
			vv = append(vv, ev)
		}
		return cty.TupleVal(vv), nil
	}
	panic(fmt.Sprintf("Not supported: %s", ty.FriendlyName()))
}

// DeleteChange is a Change implementation that represents removing an
// element from an indexable collection.
//
// For a list type, if the deleted element is not the final element in
// the list then the resulting "gap" is closed by renumbering all subsequent
// items. Therefore a Diff containing a sequence of DeleteChange operations
// on the same list must be careful to consider the new state of the element
// indices after each step, or present the deletions in reverse order to
// avoid such complexity.
type DeleteChange struct {
	changeImpl
	Path     cty.Path
	OldValue cty.Value
}

func (c DeleteChange) apply(val cty.Value) (cty.Value, error) {
	// Compare existing.
	existing, err := c.Path.Apply(val)
	if err != nil {
		return cty.NilVal, c.Path.NewErrorf("path does not exist in value")
	}
	if !existing.RawEquals(c.OldValue) {
		return cty.NilVal, c.Path.NewErrorf("existing value does not match")
	}
	parent, _ := c.Path[:len(c.Path)-1].Apply(val)
	key := c.Path[len(c.Path)-1]
	ty := parent.Type()
	switch {
	case ty.IsObjectType():
		kv := parent.AsValueMap()
		delete(kv, key.(cty.GetAttrStep).Name)
		return cty.ObjectVal(kv), nil
	case ty.IsMapType():
		kv := parent.AsValueMap()
		delete(kv, key.(cty.IndexStep).Key.AsString())
		return cty.MapVal(kv), nil
	case ty.IsListType():
		var vv []cty.Value
		idx := key.(cty.IndexStep).Key
		for it := parent.ElementIterator(); it.Next(); {
			i, ev := it.Element()
			if i.RawEquals(idx) {
				// Skip
				continue
			}
			vv = append(vv, ev)
		}
		return cty.ListVal(vv), nil
	case ty.IsTupleType():
		var vv []cty.Value
		idx := key.(cty.IndexStep).Key
		for it := parent.ElementIterator(); it.Next(); {
			i, ev := it.Element()
			if i.RawEquals(idx) {
				// Skip
				continue
			}
			vv = append(vv, ev)
		}
		return cty.TupleVal(vv), nil
	}
	return cty.NilVal, c.Path.NewErrorf("value is not indexable")
}

// InsertChange is a Change implementation that represents inserting a new
// element into a list.
//
// When appending to a list, the Path should be to the not-yet-existing index
// and BeforeValue should be a null of the appropriate type.
type InsertChange struct {
	changeImpl
	Path        cty.Path
	NewValue    cty.Value
	BeforeValue cty.Value
}

func (c InsertChange) apply(val cty.Value) (cty.Value, error) {
	list, err := c.Path.Apply(val)
	if err != nil {
		return cty.NilVal, c.Path.NewErrorf("path does not exist in value")
	}
	if !list.CanIterateElements() {
		return cty.NilVal, c.Path.NewErrorf("value is not iterable")
	}
	ty := list.Type()
	vals := list.AsValueSlice()
	out := make([]cty.Value, 0, len(vals)+1)
	if len(vals) == 0 && c.BeforeValue.IsNull() {
		if ty.IsListType() {
			if !c.BeforeValue.Type().Equals(ty.ElementType()) {
				return cty.NilVal, c.Path.NewErrorf("before value must be a %s", ty.ElementType())
			}
		}
		out = append(out, c.NewValue)
	} else {
		match := false
		for _, v := range vals {
			if v.RawEquals(c.BeforeValue) && !match {
				out = append(out, c.NewValue)
				match = true
			}
			out = append(out, v)
		}
		if !match {
			return cty.NilVal, c.Path.NewErrorf("before value does not exist")
		}
	}
	switch {
	case ty.IsListType():
		return cty.ListVal(out), nil
	case ty.IsTupleType():
		return cty.TupleVal(out), nil
	}
	panic(fmt.Sprintf("Not supported: %s", ty.FriendlyName()))
}

// AddChange is a Change implementation that represents adding a value to
// a set. The Path is to the set itself, and NewValue is the value to insert.
type AddChange struct {
	changeImpl
	Path     cty.Path
	NewValue cty.Value
}

func (c AddChange) apply(val cty.Value) (cty.Value, error) {
	set, err := c.Path.Apply(val)
	if err != nil {
		return cty.NilVal, c.Path.NewErrorf("path does not exist in value")
	}
	if !set.Type().IsSetType() {
		return cty.NilVal, c.Path.NewErrorf("value is not a set")
	}
	s := set.AsValueSet()
	s.Add(c.NewValue)
	return cty.SetVal(s.Values()), nil
}

// RemoveChange is a Change implementation that represents removing a value
// from a set. The path is to the set itself, and OldValue is the value to
// remove.
//
// Note that it is not possible to remove an unknown value from a set
// because no two unknown values are equal, so a diff whose source value
// had sets with unknown members cannot be applied and is useful only
// for presentation to a user. Generally-speaking one should avoid including
// unknowns in the source value when creating a diff.
type RemoveChange struct {
	changeImpl
	Path     cty.Path
	OldValue cty.Value
}

func (c RemoveChange) apply(val cty.Value) (cty.Value, error) {
	set, err := c.Path.Apply(val)
	if err != nil {
		return cty.NilVal, c.Path.NewErrorf("path does not exist in value")
	}
	if !set.Type().IsSetType() {
		return cty.NilVal, c.Path.NewErrorf("value is not a set")
	}
	s := set.AsValueSet()
	if !s.Has(c.OldValue) {
		return cty.NilVal, c.Path.NewErrorf("old value does not exist")
	}
	s.Remove(c.OldValue)
	return cty.SetVal(s.Values()), nil
}

// NestedDiff is a Change implementation that applies a nested diff to a
// value.
//
// A NestedDiff is similar to a ReplaceChange, except that rather than
// providing a literal new value the replacement value is instead the result
// of applying the diff to the old value.
//
// The Paths in the nested diff are relative to the path of the NestedDiff
// node, so the absolute paths of the affected elements are the concatenation
// of the two.
//
// This is primarily useful for representing updates to set elements. Since
// set elements are addressed by their own value, it convenient to specify
// the value path only once and apply a number of other operations to it.
// However, it's acceptable to use NestedDiff on any value type as long as
// the nested diff is valid for that type.
type NestedDiff struct {
	changeImpl
	Path     cty.Path
	OldValue cty.Value
	Diff     Diff
}

func (c NestedDiff) apply(val cty.Value) (cty.Value, error) {
	return cty.NullVal(val.Type()), errors.New("not yet implemented")
}

// Context is a funny sort of Change implementation that doesn't actually
// change anything but fails if the value at the given path doesn't match
// the given value.
//
// This can be used to add additional context to a diff so that merge
// conflicts can be detected.
type Context struct {
	changeImpl
	Path      cty.Path
	WantValue cty.Value
}

func (c Context) apply(val cty.Value) (cty.Value, error) {
	existing, err := c.Path.Apply(val)
	if err != nil {
		return cty.NilVal, c.Path.NewErrorf("path does not exist in value")
	}
	if !existing.RawEquals(c.WantValue) {
		return cty.NilVal, c.Path.NewErrorf("existing value does not match")
	}
	return val, nil
}
