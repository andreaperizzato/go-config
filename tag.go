package config

import "strings"

type tagValue struct {
	name     string
	optional bool
}

func newTagValue(tag, fieldName string) tagValue {
	bits := strings.Split(tag, ",")
	if len(bits) == 0 || bits[0] == "" {
		return tagValue{name: fieldName}
	}
	t := tagValue{name: bits[0]}
	if len(bits) > 1 {
		t.optional = bits[1] == "optional"
	}
	return t
}
