package set

// Add inserts the given value into the recieving Set.
//
// This mutates the set in-place. This operation is not thread-safe.
func (s Set) Add(val interface{}) {
	hv := s.rules.Hash(val)
	if _, ok := s.vals[hv]; !ok {
		s.vals[hv] = make([]interface{}, 0, 1)
	}
	bucket := s.vals[hv]

	// See if an equivalent value is already present
	for _, ev := range bucket {
		if s.rules.Equivalent(val, ev) {
			return
		}
	}

	s.vals[hv] = append(bucket, val)
}
