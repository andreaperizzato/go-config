package config

import "os"

// EnvTag is the name of the tag to load variables from the environment.
const EnvTag = "env"

// NewEnvSource creates a new Source for the current environment.
func NewEnvSource() Source {
	return &source{
		tag: EnvTag,
		get: loadFromEnv,
	}
}

func loadFromEnv(key string) (string, error) {
	return os.Getenv(key), nil
}
