package cty

import (
	"fmt"
	"testing"
)

func TestPathString(t *testing.T) {
	testCases := []struct {
		Path           Path
		ExpectedOutput string
	}{
		{
			Path{},
			"",
		},
		{
			Path{
				GetAttrStep{Name: "example"},
				GetAttrStep{Name: "subkey"},
			},
			`example.subkey`,
		},
		{
			Path{
				GetAttrStep{Name: "example"},
				GetAttrStep{Name: "subkey"},
				IndexStep{Key: StringVal("strKey")},
			},
			`example.subkey["strKey"]`,
		},
		{
			Path{
				GetAttrStep{Name: "example"},
				GetAttrStep{Name: "a_list"},
				IndexStep{Key: NumberIntVal(5)},
			},
			`example.a_list[5]`,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			given := tc.Path.String()
			if given != tc.ExpectedOutput {
				t.Fatalf("Expected %q, given: %q", tc.ExpectedOutput, given)
			}
		})
	}
}

func TestPathPrepend(t *testing.T) {
	p := Path{
		GetAttrStep{Name: "key"},
	}
	given := p.Prepend(GetAttrStep{Name: "parent"})
	if given.String() != "parent.key" {
		t.Fatalf("Expected: %q, given: %q", "parent.key", given.String())
	}
}

func TestPathAppend(t *testing.T) {
	p := Path{
		GetAttrStep{Name: "key"},
	}
	given := p.Append(GetAttrStep{Name: "subkey"})
	if given.String() != "key.subkey" {
		t.Fatalf("Expected: %q, given: %q", "key.subkey", given.String())
	}
}
