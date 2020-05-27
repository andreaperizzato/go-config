package config

import "strings"

type tagValue struct {
	name     string
	optional bool
	flags    map[flag]struct{}
}

type flag string

var flagOptional flag = "optional"

func newTagValue(tag, fieldName string) tagValue {
	bits := strings.Split(tag, ",")
	if len(bits) == 0 || bits[0] == "" {
		return tagValue{name: fieldName}
	}
	t := tagValue{name: bits[0]}
	if len(bits) > 1 {
		t.flags = make(map[flag]struct{})
		for _, k := range bits[1:] {
			t.flags[flag(k)] = struct{}{}
		}
	}
	return t
}
