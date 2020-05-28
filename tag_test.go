package config

import "testing"

func TestTagValue(t *testing.T) {
	tag := newTagValue("testParameter,optional, secure", "")
	if tag.Name != "testParameter" {
		t.Fatalf("expected tag name to be 'test parameter', but got %s", tag.Name)
	}
	if found := tag.HasFlag("optional"); !found {
		t.Fatalf("expected to find the 'optional' flag")
	}
	if found := tag.HasFlag("secure"); !found {
		t.Fatalf("expected to find the 'secure' flag")
	}
	if found := tag.HasFlag("delicious"); found {
		t.Fatalf("did not expected to find the 'delicious' flag")
	}
}
