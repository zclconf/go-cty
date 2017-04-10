package cty

// setRules provides a Rules implementation for the ./set package that
// respects the equality rules for cty values of the given type.
//
// This implementation expects that values added to the set will be
// valid internal values for the given Type, which is to say that wrapping
// the given value in a Value struct along with the ruleset's type should
// produce a valid, working Value.
type setRules struct {
	Type Type
}

func (r setRules) Hash(v interface{}) int {
	// TODO: implement this
	return 0
}

func (r setRules) Equivalent(v1 interface{}, v2 interface{}) bool {
	v1v := Value{
		ty: r.Type,
		v:  v1,
	}
	v2v := Value{
		ty: r.Type,
		v:  v2,
	}

	eqv := v1v.Equals(v2v)

	// By comparing the result to true we ensure that an Unknown result,
	// which will result if either value is unknown, will be considered
	// as non-equivalent. Two unknown values are not equivalent for the
	// sake of set membership.
	return eqv.v == true
}
