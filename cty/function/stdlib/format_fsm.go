// line 1 "format_fsm.rl"
// This file is generated from format_fsm.rl. DO NOT EDIT.

// line 5 "format_fsm.rl"

package stdlib

import (
	"bytes"
	"fmt"

	"github.com/zclconf/go-cty/cty"
)

// line 19 "format_fsm.go"
var _formatfsm_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 2, 1, 4,
	1, 5, 1, 6, 1, 7, 1, 8,
	1, 9, 1, 10, 1, 11, 1, 12,
	1, 13, 1, 14, 1, 16, 1, 17,
	1, 18, 2, 3, 4, 2, 12, 18,
	2, 15, 18,
}

var _formatfsm_key_offsets []byte = []byte{
	0, 0, 15, 29, 34, 41, 45, 51,
	58, 60, 63, 71,
}

var _formatfsm_trans_keys []byte = []byte{
	32, 35, 37, 42, 43, 45, 46, 48,
	91, 49, 57, 65, 90, 97, 122, 32,
	35, 42, 43, 45, 46, 48, 91, 49,
	57, 65, 90, 97, 122, 46, 65, 90,
	97, 122, 42, 48, 57, 65, 90, 97,
	122, 65, 90, 97, 122, 48, 57, 65,
	90, 97, 122, 46, 48, 57, 65, 90,
	97, 122, 49, 57, 93, 48, 57, 42,
	46, 49, 57, 65, 90, 97, 122, 37,
}

var _formatfsm_single_lengths []byte = []byte{
	0, 9, 8, 1, 1, 0, 0, 1,
	0, 1, 2, 1,
}

var _formatfsm_range_lengths []byte = []byte{
	0, 3, 3, 2, 3, 2, 3, 3,
	1, 1, 3, 0,
}

var _formatfsm_index_offsets []byte = []byte{
	0, 0, 13, 25, 29, 34, 37, 41,
	46, 48, 51, 57,
}

var _formatfsm_indicies []byte = []byte{
	1, 2, 3, 4, 5, 6, 7, 8,
	11, 9, 10, 10, 0, 1, 2, 4,
	5, 6, 7, 8, 11, 9, 10, 10,
	0, 12, 13, 13, 0, 14, 15, 16,
	16, 0, 16, 16, 0, 15, 16, 16,
	0, 12, 9, 13, 13, 0, 17, 0,
	18, 17, 0, 4, 7, 9, 10, 10,
	0, 19, 3,
}

var _formatfsm_trans_targs []byte = []byte{
	0, 2, 2, 11, 3, 2, 2, 4,
	2, 7, 11, 8, 4, 11, 5, 6,
	11, 9, 10, 1,
}

var _formatfsm_trans_actions []byte = []byte{
	7, 17, 9, 3, 27, 15, 13, 0,
	11, 25, 33, 19, 23, 38, 31, 29,
	41, 21, 0, 1,
}

var _formatfsm_eof_actions []byte = []byte{
	0, 35, 35, 35, 35, 35, 35, 35,
	35, 35, 35, 5,
}

const formatfsm_start int = 11
const formatfsm_first_final int = 11
const formatfsm_error int = 0

const formatfsm_en_main int = 11

// line 18 "format_fsm.rl"

func formatFSM(format string, a []cty.Value) (string, error) {
	var buf bytes.Buffer
	data := format
	nextArg := 1 // arg numbers are 1-based
	var verb formatVerb

	// line 149 "format_fsm.rl"

	// Ragel state
	p := 0          // "Pointer" into data
	pe := len(data) // End-of-data "pointer"
	cs := 0         // current state (will be initialized by ragel-generated code)
	ts := 0
	te := 0
	eof := pe

	// Keep Go compiler happy even if generated code doesn't use these
	_ = ts
	_ = te
	_ = eof

	// line 124 "format_fsm.go"
	{
		cs = formatfsm_start
	}

	// line 129 "format_fsm.go"
	{
		var _klen int
		var _trans int
		var _acts int
		var _nacts uint
		var _keys int
		if p == pe {
			goto _test_eof
		}
		if cs == 0 {
			goto _out
		}
	_resume:
		_keys = int(_formatfsm_key_offsets[cs])
		_trans = int(_formatfsm_index_offsets[cs])

		_klen = int(_formatfsm_single_lengths[cs])
		if _klen > 0 {
			_lower := int(_keys)
			var _mid int
			_upper := int(_keys + _klen - 1)
			for {
				if _upper < _lower {
					break
				}

				_mid = _lower + ((_upper - _lower) >> 1)
				switch {
				case data[p] < _formatfsm_trans_keys[_mid]:
					_upper = _mid - 1
				case data[p] > _formatfsm_trans_keys[_mid]:
					_lower = _mid + 1
				default:
					_trans += int(_mid - int(_keys))
					goto _match
				}
			}
			_keys += _klen
			_trans += _klen
		}

		_klen = int(_formatfsm_range_lengths[cs])
		if _klen > 0 {
			_lower := int(_keys)
			var _mid int
			_upper := int(_keys + (_klen << 1) - 2)
			for {
				if _upper < _lower {
					break
				}

				_mid = _lower + (((_upper - _lower) >> 1) & ^1)
				switch {
				case data[p] < _formatfsm_trans_keys[_mid]:
					_upper = _mid - 2
				case data[p] > _formatfsm_trans_keys[_mid+1]:
					_lower = _mid + 2
				default:
					_trans += int((_mid - int(_keys)) >> 1)
					goto _match
				}
			}
			_trans += _klen
		}

	_match:
		_trans = int(_formatfsm_indicies[_trans])
		cs = int(_formatfsm_trans_targs[_trans])

		if _formatfsm_trans_actions[_trans] == 0 {
			goto _again
		}

		_acts = int(_formatfsm_trans_actions[_trans])
		_nacts = uint(_formatfsm_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _formatfsm_actions[_acts-1] {
			case 0:
				// line 28 "format_fsm.rl"

				verb = formatVerb{
					ArgNum: nextArg,
					Prec:   -1,
					Width:  -1,
				}
				ts = p

			case 1:
				// line 37 "format_fsm.rl"

				buf.WriteByte(data[p])

			case 4:
				// line 48 "format_fsm.rl"

				return buf.String(), fmt.Errorf("unrecognized format character %q at offset %d", data[p], p)

			case 5:
				// line 52 "format_fsm.rl"

				verb.Sharp = true

			case 6:
				// line 55 "format_fsm.rl"

				verb.Zero = true

			case 7:
				// line 58 "format_fsm.rl"

				verb.Minus = true

			case 8:
				// line 61 "format_fsm.rl"

				verb.Plus = true

			case 9:
				// line 64 "format_fsm.rl"

				verb.Space = true

			case 10:
				// line 68 "format_fsm.rl"

				verb.ArgNum = 0

			case 11:
				// line 71 "format_fsm.rl"

				verb.ArgNum = (10 * verb.ArgNum) + (int(data[p]) - '0')

			case 12:
				// line 75 "format_fsm.rl"

				verb.HasWidth = true

			case 13:
				// line 78 "format_fsm.rl"

				verb.Width = (10 * verb.Width) + (int(data[p]) - '0')

			case 14:
				// line 81 "format_fsm.rl"

				verb.WidthFromArg = true

			case 15:
				// line 85 "format_fsm.rl"

				verb.HasPrec = true

			case 16:
				// line 88 "format_fsm.rl"

				verb.Prec = (10 * verb.Prec) + (int(data[p]) - '0')

			case 17:
				// line 91 "format_fsm.rl"

				verb.PrecFromArg = true

			case 18:
				// line 95 "format_fsm.rl"

				verb.Mode = rune(data[p])
				te = p + 1
				verb.Raw = data[ts:te]
				verb.Offset = ts

				newNext, err := formatAppend(&verb, &buf, a)
				if err != nil {
					return buf.String(), err
				}
				nextArg = newNext

				// line 324 "format_fsm.go"
			}
		}

	_again:
		if cs == 0 {
			goto _out
		}
		p++
		if p != pe {
			goto _resume
		}
	_test_eof:
		{
		}
		if p == eof {
			__acts := _formatfsm_eof_actions[cs]
			__nacts := uint(_formatfsm_actions[__acts])
			__acts++
			for ; __nacts > 0; __nacts-- {
				__acts++
				switch _formatfsm_actions[__acts-1] {
				case 2:
					// line 41 "format_fsm.rl"

				case 3:
					// line 44 "format_fsm.rl"

					return buf.String(), fmt.Errorf("invalid format string starting at offset %d", p)

				case 4:
					// line 48 "format_fsm.rl"

					return buf.String(), fmt.Errorf("unrecognized format character %q at offset %d", data[p], p)

					// line 360 "format_fsm.go"
				}
			}
		}

	_out:
		{
		}
	}

	// line 167 "format_fsm.rl"

	// If we fall out here without being in a final state then we've
	// encountered something that the scanner can't match, which should
	// be impossible (the scanner matches all bytes _somehow_) but we'll
	// flag it anyway rather than just losing data from the end.
	if cs < formatfsm_first_final {
		return buf.String(), fmt.Errorf("extraneous characters beginning at offset %i", p)
	}

	return buf.String(), nil
}
