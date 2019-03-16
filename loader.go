package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// Loader loads.
type Loader struct {
	sources map[string]Getter
}

// NewLoader creates a new Loader that uses
// all the sources.
func NewLoader(scs ...Source) *Loader {
	l := Loader{
		sources: make(map[string]Getter, len(scs)),
	}
	for _, s := range scs {
		l.sources[s.Tag()] = s.Get
	}
	return &l
}

// Load reads the.
func (c *Loader) Load(v interface{}) error {
	rv := reflect.ValueOf(v)
	rv, err := getWritableValue(rv)
	if err != nil {
		return err
	}
	rt := reflect.TypeOf(v).Elem()
	for t, getter := range c.sources {
		err := loadValues(v, rv, rt, t, getter)
		if err != nil {
			return err
		}
	}
	return nil
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

func loadValues(v interface{}, rv reflect.Value, rt reflect.Type, tagName string, getter Getter) error {
	for i := 0; i < rv.NumField(); i++ {
		fv := rv.Field(i)
		ft := rt.Field(i)
		val, found := ft.Tag.Lookup(tagName)
		if !found {
			continue
		}
		tag := newTagValue(val, ft.Name)
		if !fv.CanSet() {
			return fmt.Errorf("config: field %s can't be set", ft.Name)
		}
		switch fv.Kind() {
		case reflect.Bool:
			b, err := loadBool(tag, getter)
			if err != nil {
				return err
			}
			fv.SetBool(b)

		case reflect.String:
			s, err := loadString(tag, getter)
			if err != nil {
				return err
			}
			fv.SetString(s)

		case reflect.Int64:
			i, err := loadInt(64, tag, getter)
			if err != nil {
				return err
			}
			fv.SetInt(i)

		case reflect.Int32:
			i, err := loadInt(32, tag, getter)
			if err != nil {
				return err
			}
			fv.SetInt(i)

		case reflect.Int16:
			i, err := loadInt(16, tag, getter)
			if err != nil {
				return err
			}
			fv.SetInt(i)

		case reflect.Int8:
			i, err := loadInt(8, tag, getter)
			if err != nil {
				return err
			}
			fv.SetInt(i)
		}
	}
	return nil
}

func missingValue(val string, tag tagValue) bool {
	return val == "" && !tag.optional
}

func missingValueError(key string) error {
	return fmt.Errorf("config: missing value for key '%s'", key)
}

func loadBool(tag tagValue, getter Getter) (bool, error) {
	val, err := getter(tag.name)
	if err != nil {
		return false, err
	}
	if missingValue(val, tag) {
		return false, missingValueError(tag.name)
	}
	return val == "true" || val == "1", nil
}

func loadString(tag tagValue, getter Getter) (string, error) {
	val, err := getter(tag.name)
	if err != nil {
		return "", err
	}
	if missingValue(val, tag) {
		return "", missingValueError(tag.name)
	}
	return val, nil
}

func loadInt(bitSize int, tag tagValue, getter Getter) (int64, error) {
	val, err := getter(tag.name)
	if err != nil {
		return 0, err
	}
	if missingValue(val, tag) {
		return 0, missingValueError(tag.name)
	}
	return strconv.ParseInt(val, 10, bitSize)
}
