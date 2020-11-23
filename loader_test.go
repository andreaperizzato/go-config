package config

import (
	"fmt"
	"io"
	"reflect"
	"testing"
)

type testSource struct {
	tag    string
	values map[string]string
}

func (ts *testSource) Tag() string {
	if ts.tag == "" {
		return "test"
	}
	return ts.tag
}
func (ts *testSource) Get(tag TagValue) (string, error) {
	v, found := ts.values[tag.Name]
	if !found {
		return "", fmt.Errorf("error getting key %s", tag.Name)
	}
	return v, nil
}

func Test_SingleSource(t *testing.T) {
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
			err: "config: missing value for field 'T'",
		},
		{
			desc: "error getting value",
			v: &struct {
				T string `test:"ttt"`
			}{},
			values: map[string]string{},
			err:    "config: error loading field T for tag test: error getting key ttt",
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
			desc: "error for unsupported types",
			v: &struct {
				T io.Reader `test:"reader"`
			}{},
			err: "config: field type Reader is not supported",
		},
		{
			desc: "string field",
			v: &struct {
				T string `test:"field"`
				D string `test:"other" default:"venice"`
				Q string `test:"hello" default:""`
			}{},
			values: map[string]string{
				"field": "hello",
				"other": "",
				"hello": "",
			},
			out: &struct {
				T string `test:"field"`
				D string `test:"other" default:"venice"`
				Q string `test:"hello" default:""`
			}{
				T: "hello",
				D: "venice",
				Q: "",
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
				D1 bool `test:"other1" default:"true"`
				D2 bool `test:"other2" default:"1"`
				D3 bool `test:"other3" default:"0"`
				D4 bool `test:"other4" default:""`
			}{},
			values: map[string]string{
				"field1": "true",
				"field2": "1",
				"field3": "other",
				"field4": "0",
				"field5": "false",
				"other1": "",
				"other2": "",
				"other3": "",
				"other4": "",
			},
			out: &struct {
				T1 bool `test:"field1"`
				T2 bool `test:"field2"`
				T3 bool `test:"field3"`
				T4 bool `test:"field4"`
				T5 bool `test:"field5"`
				D1 bool `test:"other1" default:"true"`
				D2 bool `test:"other2" default:"1"`
				D3 bool `test:"other3" default:"0"`
				D4 bool `test:"other4" default:""`
			}{
				T1: true,
				T2: true,
				T3: false,
				T4: false,
				T5: false,
				D1: true,
				D2: true,
				D3: false,
				D4: false,
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
				T5 int   `test:"field5"`
				D1 int64 `test:"other1" default:"91"`
				D2 int32 `test:"other2" default:"92"`
				D3 int16 `test:"other3" default:"93"`
				D4 int8  `test:"other4" default:"94"`
				D5 int   `test:"other5" default:"95"`
				D6 int   `test:"other5" default:""`
			}{},
			values: map[string]string{
				"field1": "11",
				"field2": "22",
				"field3": "33",
				"field4": "44",
				"field5": "55",
				"other1": "",
				"other2": "",
				"other3": "",
				"other4": "",
				"other5": "",
				"other6": "",
			},
			out: &struct {
				T1 int64 `test:"field1"`
				T2 int32 `test:"field2"`
				T3 int16 `test:"field3"`
				T4 int8  `test:"field4"`
				T5 int   `test:"field5"`
				D1 int64 `test:"other1" default:"91"`
				D2 int32 `test:"other2" default:"92"`
				D3 int16 `test:"other3" default:"93"`
				D4 int8  `test:"other4" default:"94"`
				D5 int   `test:"other5" default:"95"`
				D6 int   `test:"other5" default:""`
			}{
				T1: 11,
				T2: 22,
				T3: 33,
				T4: 44,
				T5: 55,
				D1: 91,
				D2: 92,
				D3: 93,
				D4: 94,
				D5: 95,
				D6: 0,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			l := NewLoader(&testSource{values: tC.values})
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

func Test_MultipleSources(t *testing.T) {
	testCases := []struct {
		desc string
		v    interface{}
		s1   Source
		s2   Source
		out  interface{}
		err  string
	}{
		{
			desc: "error with first source stops processing",
			v: &struct {
				T string `test:"a" happy:"b"`
			}{},
			s1:  &testSource{tag: "test", values: map[string]string{}},
			s2:  &testSource{tag: "happy", values: map[string]string{"b": "hello"}},
			err: "config: error loading field T for tag test: error getting key a",
		},
		{
			desc: "missing value in all sources",
			v: &struct {
				T string `test:"a" happy:"b" json:"x"`
			}{},
			s1:  &testSource{tag: "test", values: map[string]string{"a": ""}},
			s2:  &testSource{tag: "happy", values: map[string]string{"b": ""}},
			err: "config: missing value for field 'T'",
		},
		{
			desc: "sources are evaluated in sequence",
			v: &struct {
				T string `test:"a" happy:"b"`
			}{},
			s1: &testSource{tag: "test", values: map[string]string{"a": "replaced"}},
			s2: &testSource{tag: "happy", values: map[string]string{"b": "hello"}},
			out: &struct {
				T string `test:"a" happy:"b"`
			}{
				T: "hello",
			},
		},
		{
			desc: "default is ignored when one source provides value",
			v: &struct {
				T string `test:"a" happy:"b" default:"ignored"`
			}{},
			s1: &testSource{tag: "test", values: map[string]string{"a": "replaced"}},
			s2: &testSource{tag: "happy", values: map[string]string{"b": "hello"}},
			out: &struct {
				T string `test:"a" happy:"b" default:"ignored"`
			}{
				T: "hello",
			},
		},
		{
			desc: "default is used when no sources provide a value",
			v: &struct {
				T string `test:"a" happy:"b" default:"actual"`
			}{},
			s1: &testSource{tag: "test", values: map[string]string{"a": ""}},
			s2: &testSource{tag: "happy", values: map[string]string{"b": ""}},
			out: &struct {
				T string `test:"a" happy:"b" default:"actual"`
			}{
				T: "actual",
			},
		},
		{
			desc: "empty default makes field optional",
			v: &struct {
				T string `test:"a" happy:"b" default:""`
			}{},
			s1: &testSource{tag: "test", values: map[string]string{"a": ""}},
			s2: &testSource{tag: "happy", values: map[string]string{"b": ""}},
			out: &struct {
				T string `test:"a" happy:"b" default:""`
			}{},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			l := NewLoader(tC.s1, tC.s2)
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

func Test_DeprecatedOptionalFlag(t *testing.T) {
	l := NewLoader(&testSource{
		tag: "test",
		values: map[string]string{
			"ttt": "",
		},
	})

	v := struct {
		T string `test:"ttt,optional"`
	}{}
	out := v

	err := l.Load(&v)
	if err != nil {
		t.Fatalf("expected error to be nil but was '%s'", err.Error())
	}
	if !reflect.DeepEqual(out, v) {
		t.Errorf("expected output to be %v but was %v", out, v)
	}
}
