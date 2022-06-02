package set

// testRules is a rules implementation that is used for testing. It only
// accepts ints as values, and it has a Hash function that just returns the
// given value modulo 16 so that we can easily and dependably test the
// situation where two non-equivalent values have the same hash value.
type testRules struct{}

func newTestRules() Rules[int] {
	return testRules{}
}

func (r testRules) Hash(val int) int {
	return val % 16
}

func (r testRules) Equivalent(val1 int, val2 int) bool {
	return val1 == val2
}

func (r testRules) SameRules(other Rules[int]) bool {
	// All testRules values are equal, so type-checking is enough.
	_, ok := other.(testRules)
	return ok
}
