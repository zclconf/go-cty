package ctymarks

// ProvenanceMark is embedded inside any mark type that should be treated by
// cty as a "provenance mark".
//
// This special kind of mark is tracked more conservatively by cty, and so
// applications which use provenance marks should be written to accommodate
// some additional constraints:
//
//   - Any operation involving unknown values where there is a known value that
//     could potentially cause the result to have a provenance mark is expected
//     to produce that mark on its result, even though there might be other
//     known values that would not produce that mark.
//   - If there are any operations where the previous rule doesn't hold then
//     that's likely to be considered a bug and fixed in a later release without
//     considering it a breaking change, and so callers should use provenance
//     marks only in situations where them being reported more accurately in
//     future would be considered an improvement rather than a regression.
//   - Whenever at least one provenance mark is in scope it becomes possible
//     for there to be unknown values of unknown type that are not considered by
//     Go's "==" operator to be equal to cty.DynamicVal, because they are
//     tracking nested marks for use in downstream operations. (Conversely,
//     applications that DO NOT use provenance marks can assume that
//     [cty.DynamicVal] is the only representation of an unknown value of
//     unknown type, once any shallow marks have been removed.)
//
// This special kind of mark exists as a pragmatic concession to allow
// introducing some more conservative tracking of marks without breaking
// applications that were relying on cty's earlier "best effort" approach to
// marks in operations involving unknown values, where therefore sometimes
// operations involving unknown values would not produce a marked result even
// though a subsequent re-evaluation with known values would produce a marked
// result.
//
// For example, [OpenTofu] uses marks to represent its concepts of "sensitive"
// values and accepts that sometimes an unknown value that was not considered
// sensitive can be replaced by a sensitive value once it becomes known in
// a later phase. Since OpenTofu treats sensitive values in certain locations
// as an error, OpenTofu generally prefers the more "best-effort" handling of
// that mark instead of conservatively reporting it whenever there's any
// possibility of a sensitive value being present after all values are known.
// Therefore OpenTofu should not consider its "sensitive" mark to be a
// provenance mark.
//
// However, OpenTofu also uses marks to track dynamic dependencies between
// operations by noticing when a value used to configure one object is derived
// from a result from another object. In that case, failing to report a mark
// that could potentially appear in the final known result would cause a
// correctness problem and so OpenTofu should treat the marks used in that case
// as provenance marks, expecting that any gaps in the tracking of those marks
// discovered in future will be fixed in a future version of cty.
//
// [OpenTofu]: https://opentofu.org/
type ProvenanceMark struct{}

var _ provenanceMark = ProvenanceMark{}

func (ProvenanceMark) isProvenanceMark() {}

// IsProvenanceMark returns true if the type of the given mark has
// [ProvenanceMark] embedded in it, which causes some different treatment of
// the mark as described in that type's documentation.
func IsProvenanceMark(mark any) bool {
	_, ok := mark.(provenanceMark)
	return ok
}

type provenanceMark interface {
	isProvenanceMark()
}
