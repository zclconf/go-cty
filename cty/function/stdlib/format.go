package stdlib

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/apparentlymart/go-textseg/textseg"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/json"
)

//go:generate ragel -Z format_fsm.rl
//go:generate gofmt -w format_fsm.go

var FormatFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "format",
			Type: cty.String,
		},
	},
	VarParam: &function.Parameter{
		Name:      "args",
		Type:      cty.DynamicPseudoType,
		AllowNull: true,
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		str, err := formatFSM(args[0].AsString(), args[1:])
		return cty.StringVal(str), err
	},
})

// Format produces a string representation of zero or more values using a
// format string similar to the "printf" function in C.
//
// It supports the following "verbs":
//
//     %%      Literal percent sign, consuming no value
//     %v      A default formatting of the value based on type, as described below.
//     %#v     JSON serialization of the value
//     %t      Converts to boolean and then produces "true" or "false"
//     %b      Converts to number, requires integer, produces binary representation
//     %d      Converts to number, requires integer, produces decimal representation
//     %o      Converts to number, requires integer, produces octal representation
//     %x      Converts to number, requires integer, produces hexadecimal representation
//             with lowercase letters
//     %X      Like %x but with uppercase letters
//     %e      Converts to number, produces scientific notation like -1.234456e+78
//     %E      Like %e but with an uppercase "E" representing the exponent
//     %f      Converts to number, produces decimal representation with fractional
//             part but no exponent, like 123.456
//     %g      %e for large exponents or %f otherwise
//     %G      %E for large exponents or %f otherwise
//     %s      Converts to string and produces the string's characters
//     %q      Converts to string and produces JSON-quoted string representation,
//             like %v.
//
// The default format selections made by %v are:
//
//     string  %s
//     number  %g
//     bool    %t
//     other   %#v
//
// Null values produce the literal keyword "null" for %v and %#v, and produce
// an error otherwise.
//
// Width is specified by an optional decimal number immediately preceding the
// verb letter. If absent, the width is whatever is necessary to represent the
// value. Precision is specified after the (optional) width by a period
// followed by a decimal number. If no period is present, a default precision
// is used. A period with no following number is invalid.
// For examples:
//
//     %f     default width, default precision
//     %9f    width 9, default precision
//     %.2f   default width, precision 2
//     %9.2f  width 9, precision 2
//
// Width and precision are measured in unicode characters (grapheme clusters).
//
// For most values, width is the minimum number of characters to output,
// padding the formatted form with spaces if necessary.
//
// For strings, precision limits the length of the input to be formatted (not
// the size of the output), truncating if necessary.
//
// For numbers, width sets the minimum width of the field and precision sets
// the number of places after the decimal, if appropriate, except that for
// %g/%G precision sets the total number of significant digits.
//
// The following additional symbols can be used immediately after the percent
// introducer as flags:
//
//           (a space) leave a space where the sign would be if number is positive
//     +     Include a sign for a number even if it is positive (numeric only)
//     -     Pad with spaces on the left rather than the right
//     0     Pad with zeros rather than spaces.
//
// Flag characters are ignored for verbs that do not support them.
//
// By default, % sequences consume successive arguments starting with the first.
// Introducing a [n] sequence immediately before the verb letter, where n is a
// decimal integer, explicitly chooses a particular value argument by its
// one-based index. Subsequent calls without an explicit index will then
// proceed with n+1, n+2, etc.
//
// An error is produced if the format string calls for an impossible conversion
// or accesses more values than are given. An error is produced also for
// an unsupported format verb.
func Format(format cty.Value, vals ...cty.Value) (cty.Value, error) {
	args := make([]cty.Value, 0, len(vals)+1)
	args = append(args, format)
	args = append(args, vals...)
	return FormatFunc.Call(args)
}

type formatVerb struct {
	Raw    string
	Offset int

	ArgNum int
	Mode   rune

	Zero  bool
	Sharp bool
	Plus  bool
	Minus bool
	Space bool

	HasPrec bool
	Prec    int

	HasWidth bool
	Width    int
}

// formatAppend is called by formatFSM (generated by format_fsm.rl) for each
// formatting sequence that is encountered.
func formatAppend(verb *formatVerb, buf *bytes.Buffer, args []cty.Value) error {
	argIdx := verb.ArgNum - 1
	if argIdx >= len(args) {
		return fmt.Errorf(
			"not enough arguments for %q at %d: need index %d but have %d total",
			verb.Raw, verb.Offset,
			verb.ArgNum, len(args),
		)
	}
	arg := args[argIdx]

	if verb.Mode != 'v' && arg.IsNull() {
		return fmt.Errorf("unsupported value for %q at %d: null value cannot be formatted", verb.Raw, verb.Offset)
	}

	// Normalize to make some things easier for downstream formatters
	if !verb.HasWidth {
		verb.Width = -1
	}
	if !verb.HasPrec {
		verb.Prec = -1
	}

	// For our first pass we'll ensure the verb is supported and then fan
	// out to other functions based on what conversion is needed.
	switch verb.Mode {

	case 'v':
		return formatAppendAsIs(verb, buf, arg)

	case 't':
		return formatAppendBool(verb, buf, arg)

	case 'b', 'd', 'o', 'x', 'X', 'e', 'E', 'f', 'g', 'G':
		return formatAppendNumber(verb, buf, arg)

	case 's', 'q':
		return formatAppendString(verb, buf, arg)

	default:
		return fmt.Errorf("unsupported format verb %q in %q at offset %d", verb.Mode, verb.Raw, verb.Offset)
	}
}

func formatAppendAsIs(verb *formatVerb, buf *bytes.Buffer, arg cty.Value) error {

	if !verb.Sharp && !arg.IsNull() {
		// Unless the caller overrode it with the sharp flag, we'll try some
		// specialized formats before we fall back on JSON.
		switch arg.Type() {
		case cty.String:
			fmted := arg.AsString()
			fmted = formatPadWidth(verb, fmted)
			buf.WriteString(fmted)
			return nil
		case cty.Number:
			bf := arg.AsBigFloat()
			fmted := bf.Text('g', -1)
			fmted = formatPadWidth(verb, fmted)
			buf.WriteString(fmted)
			return nil
		}
	}

	jb, err := json.Marshal(arg, arg.Type())
	if err != nil {
		return fmt.Errorf("unsupported value for %q at %d: %s", verb.Raw, verb.Offset, err)
	}
	fmted := formatPadWidth(verb, string(jb))
	buf.WriteString(fmted)

	return nil
}

func formatAppendBool(verb *formatVerb, buf *bytes.Buffer, arg cty.Value) error {
	var err error
	arg, err = convert.Convert(arg, cty.Bool)
	if err != nil {
		return fmt.Errorf("unsupported value for %q at %d: %s", verb.Raw, verb.Offset, err)
	}

	if arg.True() {
		buf.WriteString("true")
	} else {
		buf.WriteString("false")
	}
	return nil
}

func formatAppendNumber(verb *formatVerb, buf *bytes.Buffer, arg cty.Value) error {
	var err error
	arg, err = convert.Convert(arg, cty.Number)
	if err != nil {
		return fmt.Errorf("unsupported value for %q at %d: %s", verb.Raw, verb.Offset, err)
	}

	switch verb.Mode {
	case 'b', 'd', 'o', 'x', 'X':
		return formatAppendInteger(verb, buf, arg)
	default:
		bf := arg.AsBigFloat()

		// For floats our format syntax is a subset of Go's, so it's
		// safe for us to just lean on the existing Go implementation.
		fmtstr := formatStripIndexSegment(verb.Raw)
		fmted := fmt.Sprintf(fmtstr, bf)
		buf.WriteString(fmted)
		return nil
	}
}

func formatAppendInteger(verb *formatVerb, buf *bytes.Buffer, arg cty.Value) error {
	bf := arg.AsBigFloat()
	bi, acc := bf.Int(nil)
	if acc != big.Exact {
		return fmt.Errorf("unsupported value for %q at %d: an integer is required", verb.Raw, verb.Offset)
	}

	// For integers our format syntax is a subset of Go's, so it's
	// safe for us to just lean on the existing Go implementation.
	fmtstr := formatStripIndexSegment(verb.Raw)
	fmted := fmt.Sprintf(fmtstr, bi)
	buf.WriteString(fmted)
	return nil
}

func formatAppendString(verb *formatVerb, buf *bytes.Buffer, arg cty.Value) error {
	var err error
	arg, err = convert.Convert(arg, cty.String)
	if err != nil {
		return fmt.Errorf("unsupported value for %q at %d: %s", verb.Raw, verb.Offset, err)
	}

	// We _cannot_ directly use the Go fmt.Sprintf implementation for strings
	// because it measures widths and precisions in runes rather than grapheme
	// clusters.

	str := arg.AsString()
	if verb.Prec > 0 {
		strB := []byte(str)
		pos := 0
		wanted := verb.Prec
		for i := 0; i < wanted; i++ {
			next := strB[pos:]
			if len(next) == 0 {
				// ran out of characters before we hit our max width
				break
			}
			d, _, _ := textseg.ScanGraphemeClusters(strB[pos:], true)
			pos += d
		}
		str = str[:pos]
	}

	switch verb.Mode {
	case 's':
		fmted := formatPadWidth(verb, str)
		buf.WriteString(fmted)
	case 'q':
		jb, err := json.Marshal(cty.StringVal(str), cty.String)
		if err != nil {
			// Should never happen, since we know this is a known, non-null string
			panic(fmt.Errorf("failed to marshal %#v as JSON: %s", arg, err))
		}
		fmted := formatPadWidth(verb, string(jb))
		buf.WriteString(fmted)
	default:
		// Should never happen because formatAppend should've already validated
		panic(fmt.Errorf("invalid string formatting mode %q", verb.Mode))
	}
	return nil
}

func formatPadWidth(verb *formatVerb, fmted string) string {
	if verb.Width < 0 {
		return fmted
	}

	// Safe to ignore errors because ScanGraphemeClusters cannot produce errors
	givenLen, _ := textseg.TokenCount([]byte(fmted), textseg.ScanGraphemeClusters)
	wantLen := verb.Width
	if givenLen >= wantLen {
		return fmted
	}

	padLen := wantLen - givenLen
	padChar := " "
	if verb.Zero {
		padChar = "0"
	}
	pads := strings.Repeat(padChar, padLen)

	if verb.Minus {
		return fmted + pads
	}
	return pads + fmted
}

// formatStripIndexSegment strips out any [nnn] segment present in a verb
// string so that we can pass it through to Go's fmt.Sprintf with a single
// argument. This is used in cases where we're just leaning on Go's formatter
// because it's a superset of ours.
func formatStripIndexSegment(rawVerb string) string {
	// We assume the string has already been validated here, since we should
	// only be using this function with strings that were accepted by our
	// scanner in formatFSM.
	start := strings.Index(rawVerb, "[")
	end := strings.Index(rawVerb, "]")
	if start == -1 || end == -1 {
		return rawVerb
	}

	return rawVerb[:start] + rawVerb[end+1:]
}
