package config

import (
	"os"
	"testing"
)

func TestEnvSource(t *testing.T) {
	s := NewEnvSource()
	if s.Tag() != "env" {
		t.Errorf("expected tag to be '%s' but was '%s'", "env", s.Tag())
	}

	// Value not set in the enviornment.
	v, err := s.Get("unset-env")
	if err != nil {
		t.Fatalf("unexpected error getting value for key '%s': %v", "unset-env", err)
	}
	if v != "" {
		t.Errorf("expected value to be '%s' but was '%s'", "", v)
	}

	// Value set in the enviornment.
	err = os.Setenv("testenv", "testvalue")
	if err != nil {
		t.Fatalf("unexpected error setting env variable: %v", err)
	}
	v, err = s.Get("testenv")
	if err != nil {
		t.Fatalf("unexpected error getting value for key '%s': %v", "testenv", err)
	}
	if v != "testvalue" {
		t.Errorf("expected value to be '%s' but was '%s'", "testvalue", v)
	}
}
