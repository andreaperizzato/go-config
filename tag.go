package config

import "strings"

// TagValue is the parsed version of a struct field's tag
type TagValue struct {
	Name  string
	flags map[string]struct{}
}

// HasFlag returns a boolean indicating if a TagValue has a particular flag
func (t TagValue) HasFlag(f string) bool {
	_, found := t.flags[f]
	return found
}

func newTagValue(tag, fieldName string) TagValue {
	bits := strings.Split(tag, ",")
	if len(bits) == 0 || bits[0] == "" {
		return TagValue{Name: fieldName}
	}
	t := TagValue{Name: bits[0]}
	if len(bits) > 1 {
		t.flags = make(map[string]struct{})
		for _, k := range bits[1:] {
			t.flags[strings.TrimSpace(k)] = struct{}{}
		}
	}
	return t
}
