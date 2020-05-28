package config

import "strings"

// TagValue is the parsed version of a struct field's tag
type TagValue struct {
	Name  string
	Flags map[string]struct{}
}

func newTagValue(tag, fieldName string) TagValue {
	bits := strings.Split(tag, ",")
	if len(bits) == 0 || bits[0] == "" {
		return TagValue{Name: fieldName}
	}
	t := TagValue{Name: bits[0]}
	if len(bits) > 1 {
		t.Flags = make(map[string]struct{})
		for _, k := range bits[1:] {
			t.Flags[k] = struct{}{}
		}
	}
	return t
}
