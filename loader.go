package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// Loader loads values using multiple sources.
type Loader struct {
	sources []Source
}

// NewLoader creates a new Loader that uses
// all the sources.
func NewLoader(scs ...Source) *Loader {
	return &Loader{
		sources: scs,
	}
}

// Load reads the.
func (c *Loader) Load(v interface{}) error {
	rv := reflect.ValueOf(v)
	rv, err := getWritableValue(rv)
	if err != nil {
		return err
	}
	rt := reflect.TypeOf(v).Elem()
	for i := 0; i < rv.NumField(); i++ {
		ft := rt.Field(i)
		fv := rv.Field(i)
		set, err := getFieldSetter(fv, ft)
		if err != nil {
			return err
		}

		val, err := loadFieldValue(ft, c.sources)
		if err != nil {
			return err
		}

		if val == "" {
			continue
		}

		err = set(fv, val)
		if err != nil {
			return err
		}
	}
	return nil
}

func getFieldSetter(fv reflect.Value, ft reflect.StructField) (fieldSetter, error) {
	if !fv.CanSet() {
		return nil, fmt.Errorf("config: field %s can't be set", ft.Name)
	}
	set, _ := setters[fv.Kind()]
	if set == nil {
		return nil, fmt.Errorf("config: field type %s is not supported", fv.Type().Name())
	}
	return set, nil
}

func getWritableValue(rv reflect.Value) (wv reflect.Value, err error) {
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		err = errors.New("config: v is not a pointer")
		return
	}
	wv = rv.Elem()
	if wv.Kind() != reflect.Struct {
		err = errors.New("config: v is not a struct")
		return
	}
	return
}

func loadFieldValue(ft reflect.StructField, sources []Source) (value string, err error) {
	hasDeprecatedOptionalFlag := false
	matchedTags := 0
	for _, s := range sources {
		tagValue, found := ft.Tag.Lookup(s.Tag())
		if !found {
			continue
		}
		matchedTags++
		tag := newTagValue(tagValue, ft.Name)
		newValue, err := s.Get(tag)
		if err != nil {
			return "", fmt.Errorf("config: error loading field %v for tag %s: %v", ft.Name, s.Tag(), err)
		}
		if newValue != "" {
			value = newValue
		}
		hasDeprecatedOptionalFlag = tag.HasFlag("optional")
	}
	if matchedTags == 0 {
		return
	}

	if value != "" {
		return
	}

	hasDefault := false
	if value == "" {
		value, hasDefault = ft.Tag.Lookup("default")
	}

	// Previous version of this package supported an optional flag: env:"VAR,optional"
	// which would prevent the loader from failing when the field is not set.
	// This was supported only for single tags and has now been replaced with the default tag.
	// The following condition explicitly checks for that case and handles it in order to be
	// retro-compatible.
	if value == "" && !hasDefault && matchedTags == 1 && hasDeprecatedOptionalFlag {
		return "", nil
	}

	if value == "" && !hasDefault {
		return "", missingValueError(ft.Name)
	}
	return
}

func missingValueError(fieldName string) error {
	return fmt.Errorf("config: missing value for field '%s'", fieldName)
}

type fieldSetter func(fv reflect.Value, val string) error

var setters map[reflect.Kind]fieldSetter = map[reflect.Kind]fieldSetter{
	reflect.Int64:  intSetter(64),
	reflect.Int32:  intSetter(32),
	reflect.Int16:  intSetter(16),
	reflect.Int8:   intSetter(8),
	reflect.Int:    intSetter(0),
	reflect.Uint64: uintSetter(64),
	reflect.Uint32: uintSetter(32),
	reflect.Uint16: uintSetter(16),
	reflect.Uint8:  uintSetter(8),
	reflect.Uint:   uintSetter(0),

	reflect.Bool: func(fv reflect.Value, v string) error {
		fv.SetBool(v == "true" || v == "1")
		return nil
	},
	reflect.String: func(fv reflect.Value, v string) error {
		fv.SetString(v)
		return nil
	},
}

func intSetter(bitSize int) fieldSetter {
	return func(fv reflect.Value, v string) error {
		n, err := strconv.ParseInt(v, 10, bitSize)
		if err != nil {
			return err
		}
		fv.SetInt(n)
		return nil
	}
}

func uintSetter(bitSize int) fieldSetter {
	return func(fv reflect.Value, v string) error {
		n, err := strconv.ParseUint(v, 10, bitSize)
		if err != nil {
			return err
		}
		fv.SetUint(n)
		return nil
	}
}
