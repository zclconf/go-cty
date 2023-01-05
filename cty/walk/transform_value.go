package walk

import (
	"github.com/zclconf/go-cty/cty"
)

// ValueTransformer is the interface that drives the value transforming
// functions like [TransformValue], describing how to produce values of any
// arbitrary result type for a given input value.
//
// To transform from values to values, implement ValueTransformer[cty.Value].
type ValueTransformer[Result any] interface {
	// EnterValue can transform a value from the input before any traversal into
	// its children (if any).
	//
	// The transformation process cannot preserve marks on values so if the
	// input might include marked values then [EnterValue] must strip them
	// off, and probably also remember them to re-apply them in the
	// corresponding Exit method.
	//
	// If pre-processing before traversal is not needed then just return the
	// input value verbatim.
	//
	// Implementations must treat [path] as read-only and must not retain
	// it or any sub-slice of it after [EnterValue] returns, because the
	// caller will re-use the underlying buffer for other calls.
	EnterValue(input cty.Value, path cty.Path) cty.Value

	// ExitLeafValue, ExitSequenceValue, and ExitMappingValue together deal
	// with the three different situations that can arise while unwinding
	// outwards from the leaf values back up to the root.
	//
	// - ExitSequenceValue is used for known and non-null values of list,
	//   set, and tuple types.
	// - ExitMappingValue is used for known and non-null values of map and
	//   object types.
	// - ExitLeafValue is used for all other values.
	//
	// Implementations must treat [path] as read-only and must not retain
	// it or any sub-slice of it after each function returns, because the
	// caller will re-use the underlying buffer for other calls.
	//
	// In all cases the given [cty.Value] is the result of the inward call
	// to [EnterValue]. Exit calls always happen in the reverse order to Enter
	// calls with symmetry, so a transformer can maintain a stack of other
	// contextual data in its own fields if needed.
	ExitLeafValue(input cty.Value, path cty.Path) Result
	ExitSequenceValue(input []Result, orig cty.Value, path cty.Path) Result
	ExitMappingValue(input map[string]Result, orig cty.Value, path cty.Path) Result
}

func TransformValue[Result any](v cty.Value, transformer ValueTransformer[Result]) Result {
	// We'll initially allocate a path with capacity to go three levels deep,
	// so that data structures shallower than that will not need to reallocate
	// the path buffer.
	path := make(cty.Path, 0, 3)
	return transformValue(v, path, transformer)
}

func transformValue[Result any](v cty.Value, path cty.Path, transformer ValueTransformer[Result]) Result {
	v = transformer.EnterValue(v, path)
	ty := v.Type()

	switch {
	case ty.IsListType() || ty.IsSetType() || ty.IsTupleType():
		var results []Result
		for it := v.ElementIterator(); it.Next(); {
			k, ev := it.Element()
			subPath := append(path, cty.IndexStep{Key: k})
			result := transformValue(ev, subPath, transformer)
			results = append(results, result)
		}
		return transformer.ExitSequenceValue(results, v, path)
	case ty.IsMapType() || ty.IsObjectType():
		results := make(map[string]Result)
		for it := v.ElementIterator(); it.Next(); {
			kVal, ev := it.Element()
			k := kVal.AsString()
			var subPath cty.Path
			if ty.IsMapType() {
				subPath = append(path, cty.IndexStep{Key: kVal})
			} else { // must be object type, by elimination
				subPath = append(path, cty.GetAttrStep{Name: k})
			}
			result := transformValue(ev, subPath, transformer)
			results[k] = result
		}
		return transformer.ExitMappingValue(results, v, path)
	default:
		return transformer.ExitLeafValue(v, path)
	}
}
