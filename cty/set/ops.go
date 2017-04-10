package set

// Add inserts the given value into the receiving Set.
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

// Remove deletes the given value from the receiving set, if indeed it was
// there in the first place. If the value is not present, this is a no-op.
func (s Set) Remove(val interface{}) {
	hv := s.rules.Hash(val)
	bucket, ok := s.vals[hv]
	if !ok {
		return
	}

	for i, ev := range bucket {
		if s.rules.Equivalent(val, ev) {
			newBucket := make([]interface{}, 0, len(bucket)-1)
			newBucket = append(newBucket, bucket[:i]...)
			newBucket = append(newBucket, bucket[i+1:]...)
			if len(newBucket) > 0 {
				s.vals[hv] = newBucket
			} else {
				delete(s.vals, hv)
			}
			return
		}
	}
}

// Has returns true if the given value is in the receiving set, or false if
// it is not.
func (s Set) Has(val interface{}) bool {
	hv := s.rules.Hash(val)
	bucket, ok := s.vals[hv]
	if !ok {
		return false
	}

	for _, ev := range bucket {
		if s.rules.Equivalent(val, ev) {
			return true
		}
	}
	return false
}

// EachValue calls the given callback once for each value in the set, in an
// undefined order that callers should not depend on.
func (s Set) EachValue(cb func(interface{})) {
	for _, bucket := range s.vals {
		for _, val := range bucket {
			cb(val)
		}
	}
}
