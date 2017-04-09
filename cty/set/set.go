package set

// Set is an implementation of the concept of a set: a collection where all
// values are conceptually either in or out of the set, but the members are
// not ordered.
//
// This type primarily exists to be the internal type of sets in cty, but
// it is considered to be at the same level of abstraction as Go's built in
// slice and map collection types, and so should make no cty-specific
// assumptions.
//
// Set operations are not thread safe. It is the caller's responsibility to
// provide mutex guarantees where necessary.
//
// Set operations are not optimized to minimize memory pressure. Mutating
// a set will generally create garbage and so should perhaps be avoided in
// tight loops where memory pressure is a concern.
type Set struct {
	vals  map[int][]interface{}
	rules Rules
}

// NewSet returns an empty set with the membership rules given.
func NewSet(rules Rules) Set {
	return Set{
		vals:  map[int][]interface{}{},
		rules: rules,
	}
}
