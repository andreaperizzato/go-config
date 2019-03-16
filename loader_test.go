package config

import (
	"fmt"
	"reflect"
	"testing"
)

type testSource struct {
	values map[string]string
}

func (ts *testSource) Tag() string {
	return "test"
}
func (ts *testSource) Get(key string) (string, error) {
	v, found := ts.values[key]
	if !found {
		return "", fmt.Errorf("error getting key %s", key)
	}
	return v, nil
}

func TestA(t *testing.T) {
	testCases := []struct {
		desc   string
		v      interface{}
		values map[string]string
		out    interface{}
		err    string
	}{
		{
			desc: "v is not writable",
			v:    struct{ T string }{},
			err:  "config: v is not a pointer",
		},
		{
			desc: "field is not exported",
			v: &struct {
				t string `test:"a"`
			}{},
			err: "config: field t can't be set",
		},
		{
			desc: "v is not a pointer to a struct",
			v: func() *string {
				s := "test"
				return &s
			}(),
			err: "config: v is not a struct",
		},
		{
			desc: "missing required field",
			v: &struct {
				T string `test:"ttt"`
			}{},
			values: map[string]string{
				"ttt": "",
			},
			err: "config: missing value for key 'ttt'",
		},
		{
			desc: "does not fail when optional fields are not set",
			v: &struct {
				T string `test:"ttt,optional"`
			}{},
			values: map[string]string{
				"ttt": "",
			},
			out: &struct {
				T string `test:"ttt,optional"`
			}{},
		},
		{
			desc: "ignores fields without a known tag",
			v: &struct {
				T string `json:"val"`
			}{},
			out: &struct {
				T string `json:"val"`
			}{},
		},
		{
			desc: "string field",
			v: &struct {
				T string `test:"field"`
			}{},
			values: map[string]string{
				"field": "hello",
			},
			out: &struct {
				T string `test:"field"`
			}{
				T: "hello",
			},
		},
		{
			desc: "bool field",
			v: &struct {
				T1 bool `test:"field1"`
				T2 bool `test:"field2"`
				T3 bool `test:"field3"`
				T4 bool `test:"field4"`
				T5 bool `test:"field5"`
			}{},
			values: map[string]string{
				"field1": "true",
				"field2": "1",
				"field3": "other",
				"field4": "0",
				"field5": "false",
			},
			out: &struct {
				T1 bool `test:"field1"`
				T2 bool `test:"field2"`
				T3 bool `test:"field3"`
				T4 bool `test:"field4"`
				T5 bool `test:"field5"`
			}{
				T1: true,
				T2: true,
				T3: false,
				T4: false,
				T5: false,
			},
		},
		{
			desc: "int field is not a number",
			v: &struct {
				T int64 `test:"field"`
			}{},
			values: map[string]string{
				"field": "not-a-number",
			},
			err: `strconv.ParseInt: parsing "not-a-number": invalid syntax`,
		},
		{
			desc: "int field overflow",
			v: &struct {
				T int8 `test:"field"`
			}{},
			values: map[string]string{
				"field": "512",
			},
			err: `strconv.ParseInt: parsing "512": value out of range`,
		},
		{
			desc: "int field",
			v: &struct {
				T1 int64 `test:"field1"`
				T2 int32 `test:"field2"`
				T3 int16 `test:"field3"`
				T4 int8  `test:"field4"`
			}{},
			values: map[string]string{
				"field1": "11",
				"field2": "22",
				"field3": "33",
				"field4": "44",
			},
			out: &struct {
				T1 int64 `test:"field1"`
				T2 int32 `test:"field2"`
				T3 int16 `test:"field3"`
				T4 int8  `test:"field4"`
			}{
				T1: 11,
				T2: 22,
				T3: 33,
				T4: 44,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			l := NewLoader(&testSource{
				values: tC.values,
			})
			err := l.Load(tC.v)
			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != tC.err {
				t.Fatalf("expected error to be '%s' but was '%s'", tC.err, errMsg)
			}
			if tC.err == "" && !reflect.DeepEqual(tC.out, tC.v) {
				t.Errorf("expected output to be %v but was %v", tC.out, tC.v)
			}
		})
	}
}

// func Test(t *testing.T) {
// 	type data struct {
// 		A string `tt:"va"`
// 		B string `tt:"vb,optional"`
// 	}
// 	l := NewLoader(&testSource{
// 		values: map[string]string{
// 			"va": "test-a",
// 			"vb": "",
// 		},
// 	})
// 	var d data
// 	err := l.Load(&d)
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}
// 	if d.A != "test-a" {
// 		t.Errorf("expected `a` to be '%s' but was '%s'", "test-a", d.A)
// 	}
// }
