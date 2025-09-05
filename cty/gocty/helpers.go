package gocty

import (
	"math/big"
	"reflect"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/set"
)

var valueType = reflect.TypeOf(cty.Value{})
var typeType = reflect.TypeOf(cty.Type{})

var setType = reflect.TypeOf(set.Set[any]{})

var bigFloatType = reflect.TypeOf(big.Float{})
var bigIntType = reflect.TypeOf(big.Int{})

var emptyInterfaceType = reflect.TypeOf(any(nil))

var stringType = reflect.TypeOf("")

type tagInfo struct {
	index    int
	optional bool
}

// tagOptions is the string following a comma in a struct field's "cty"
// tag, or the empty string. It does not include the leading comma.
type tagOptions string

func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1:])
	}
	return tag, tagOptions("")
}

// Contains reports whether a comma-separated list of options
// contains a particular substr flag. substr must be surrounded by a
// string boundary or commas.
func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}

// structTagInfo interrogates the fields of the given type (which must
// be a struct type, or we'll panic) and returns a map from the cty
// attribute names declared via struct tags to the tagInfo
//
// This function will panic if two fields within the struct are tagged with
// the same cty attribute name.
func structTagInfo(st reflect.Type) map[string]tagInfo {
	ct := st.NumField()
	ret := make(map[string]tagInfo, ct)

	for i := 0; i < ct; i++ {
		field := st.Field(i)
		attrName, opt := parseTag(field.Tag.Get("cty"))
		if attrName != "" {
			ret[attrName] = tagInfo{
				index:    i,
				optional: opt.Contains("optional"),
			}
		}
	}

	return ret
}
