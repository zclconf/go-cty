package ctystrings

import (
	"testing"
)

func TestSafeKnownPrefix(t *testing.T) {
	tests := []struct {
		Input, Want string
	}{
		// NOTE: Under future improvements to SafeKnownPrefix the "Want"
		// results for all of these tests can safely get longer, thereby
		// describing a more precise constraint, but we should avoid making
		// them shorter because that will weaken existing constraints from
		// older versions.
		// (We might make exceptions for behaviors that are found to be
		// clearly wrong, but consider the consequences carefully.)

		{
			"",
			"",
		},
		{
			"a",
			"", // The "a" is discarded because it might combine with diacritics to follow
		},
		{
			"boo",
			"bo", // The final o is discarded because it might combine with diacritics to follow
		},
		{
			"boop\r",
			"boop", // The final \r is discarded because it could combine with \r\n to produce a single grapheme cluster
		},
		{
			"hello 가",
			"hello ", // Hangul syllables can combine arbitrarily, so we must trim of trailing ones
		},
		{
			"hello 🤷🏽‍♂️",
			"hello ", // We conservatively trim the whole emoji sequence because other emoji modifiers might come in later unicode specs
		},
		{
			"hello 🤷🏽‍♂️ ",
			"hello 🤷🏽‍♂️ ", // A subsequent character avoids the need to trim
		},
		{
			"hello 🤷",
			"hello ", // "Person Shrugging" can potentially combine with subsequent skin tone modifiers or ZWJ followed by gender presentation modifiers
		},
		{
			"hello 🤷 ",
			"hello 🤷 ", // A subsequent character avoids the need to trim
		},
		{
			"hello 🤷\u200d", // U+200D is "zero width joiner"
			"hello ",        // The "Person Shrugging" followed by zero with joiner anticipates a subsequent modifier to join with
		},
		{
			"hello \U0001f1e6", // This is the beginning of a "regional indicator symbol", which are supposed to appear in pairs but we only have one here
			"hello ",           // The symbol was discarded because we can't know what character it represents until we have both parts
		},
		{
			"hello \U0001f1e6\U0001f1e6", // This is a regional indicator symbol "AA", which happens to be Aruba but it's not important exactly which country we're encoding
			"hello ",                     // The text segmentation spec allows any number of consecutive regional indicators, so we must always discard any number of them at the end.
		},
		{
			"hello \U0001f1e6\U0001f1e6 ",
			"hello \U0001f1e6\U0001f1e6 ", // A subsequent character avoids the need to trim
		},

		// The following all rely on our additional heuristic about certain
		// commonly-used delimiters that we know can never be the beginning
		// of a combined grapheme cluster sequence. We make these exceptions
		// because cty tends to be used more often for constructing strings
		// for use by machines than for constructing text for human consumption.
		{
			"ami-", // e.g. prefix of an Amazon EC2 object identifier
			"ami-",
		},
		{
			"foo_", // e.g. prefix of a variable name
			"foo_",
		},
		{
			`{"foo":`, // e.g. prefix of a JSON object
			`{"foo":`,
		},
		{
			`beep();`, // e.g. prefix of a program in a C-like language?
			`beep();`,
		},
		{
			`https://`, // e.g. prefix of a URL with a known scheme
			`https://`,
		},
		{
			`c:\`, // e.g. windows filesystem path with a known drive letter
			`c:\`,
		},
		{
			`["foo",`, // e.g. prefix of a JSON document that includes a partially-known array
			`["foo",`,
		},
		{
			`foo.bar.`, // e.g. prefix of a traversal through attributes
			`foo.bar.`,
		},
		{
			`beep(`, // e.g. prefix of a program in a C-like language?
			`beep(`,
		},
		{
			`beep()`, // e.g. prefix of a program in a C-like language?
			`beep()`,
		},
		{
			`{`, // e.g. prefix of a JSON object
			`{`,
		},
		{
			`[{}`, // e.g. fragment of JSON
			`[{}`,
		},
		{
			`[`, // e.g. prefix of a JSON array
			`[`,
		},
		{
			`[[]`, // e.g. fragment of JSON
			`[[]`,
		},
		{
			`whatever |`, // e.g. partial Unix-style command line
			`whatever |`,
		},
		{
			`https://example.com/foo?`, // e.g. prefix of a URL with a query string
			`https://example.com/foo?`,
		},
		{
			`boop!`, // dunno but seems weird to have ? without !
			`boop!`,
		},
		{
			`ls ~`, // A reference to somebody's home directory
			`ls ~`,
		},
		{
			`a `, // A space always disambiguates whether our suffix is safe
			`a `,
		},
		{
			"a\t", // A tab always disambiguates whether our suffix is safe
			"a\t",
		},
		{
			`username@`, // e.g. incomplete email address
			`username@`,
		},
		{
			`#`, // e.g. start of a single-linecomment in some machine languages, or a "hashtag"
			`#`,
		},
		{
			`print $`, // e.g. start of a reference to a Perl scalar
			`print $`,
		},
		{
			`print %`, // e.g. start of a reference to a Perl hash
			`print %`,
		},
		{
			`^`, // e.g. start of a pessimistic version constraint in some version constraint syntaxes
			`^`,
		},
		{
			`foo(&`, // e.g. the "address of" operator in some programming languages
			`foo(&`,
		},
		{
			`foo *`, // e.g. multiplying by something
			`foo *`,
		},
		{
			`foo +`, // e.g. addition
			`foo +`,
		},
		{
			`["`, // e.g. we know it's a JSON string but we don't know the content yet
			`["`,
		},
		{
			`['`, // e.g. a string in a JSON-like language that also supports single quotes!
			`['`,
		},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			got := SafeKnownPrefix(test.Input)

			if got != test.Want {
				t.Errorf("wrong result\ninput: %q\ngot:   %q\nwant:  %q", test.Input, got, test.Want)
			}
		})
	}
}
