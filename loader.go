package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// Loader loads values using multiple sources.
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
		if !fv.CanSet() {
			return fmt.Errorf("config: field %s can't be set", ft.Name)
		}

		val, found, err := getTagValue(ft, tagName, getter)
		if err != nil {
			return err
		}
		if !found {
			continue
		}

		switch fv.Kind() {
		case reflect.Bool:
			fv.SetBool(boolValue(val))

		case reflect.String:
			fv.SetString(val)

		case reflect.Int64:
			i, err := intValue(64, val)
			if err != nil {
				return err
			}
			fv.SetInt(i)

		case reflect.Int32:
			i, err := intValue(32, val)
			if err != nil {
				return err
			}
			fv.SetInt(i)

		case reflect.Int16:
			i, err := intValue(16, val)
			if err != nil {
				return err
			}
			fv.SetInt(i)

		case reflect.Int8:
			i, err := intValue(8, val)
			if err != nil {
				return err
			}
			fv.SetInt(i)

		case reflect.Int:
			i, err := intValue(0, val)
			if err != nil {
				return err
			}
			fv.SetInt(i)

		}
	}
	return nil
}

func getTagValue(ft reflect.StructField, tagName string, getter Getter) (val string, found bool, err error) {
	tagValue, found := ft.Tag.Lookup(tagName)
	if !found {
		return
	}
	tag := newTagValue(tagValue, ft.Name)
	val, err = getter(tag.name)
	if err != nil {
		return
	}
	if val == "" {
		val, _ = ft.Tag.Lookup("default")
	}
	if missingValue(val, tag) {
		err = missingValueError(tag.name)
	}
	return
}

func missingValue(val string, tag tagValue) bool {
	return val == "" && !tag.optional
}

func missingValueError(key string) error {
	return fmt.Errorf("config: missing value for key '%s'", key)
}

func boolValue(v string) bool {
	return v == "true" || v == "1"
}

func intValue(bitSize int, v string) (int64, error) {
	return strconv.ParseInt(v, 10, bitSize)
}
