package config

// Getter gets a value for a key.
type Getter func(tag TagValue) (string, error)

// Source is a source of values.
type Source interface {
	// Tag is the tag used to trigger this loader.
	Tag() string
	// Get returns the value for the key.
	Get(tag TagValue) (string, error)
}

type source struct {
	tag string
	get Getter
}

func (s *source) Tag() string {
	return s.tag
}

func (s *source) Get(tag TagValue) (string, error) {
	return s.get(tag)
}
