package config

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

func TestSSMSource(t *testing.T) {
	testParameterName := "test/parameter/name"
	testParameterValue := "parameter_value"
	sett := struct {
		TestParameter string `ssm:"test/parameter/name"`
	}{}

	svc := mockSSM{
		getParameter: func(in *ssm.GetParameterInput) (out *ssm.GetParameterOutput, err error) {
			if *in.Name != testParameterName {
				t.Fatalf("expected parameter name to be %s, got %s", testParameterName, *in.Name)
			}
			if in.WithDecryption != nil && *in.WithDecryption == true {
				t.Fatalf("expected WithDecryption to be false, got %v", *in.WithDecryption)
			}

			out = &ssm.GetParameterOutput{
				Parameter: &ssm.Parameter{
					Value: &testParameterValue,
				},
			}
			return
		},
	}

	src := NewSSMSourceWithClient(&svc)
	l := NewLoader(src)
	err := l.Load(&sett)
	if err != nil {
		t.Fatalf("failed to load config with err: %v", err)
	}
	if sett.TestParameter != testParameterValue {
		t.Fatalf("expected parameter value to be %s, got %s", testParameterValue, sett.TestParameter)
	}

	// Test when SSM.GetParameter fails
	svc = mockSSM{
		getParameter: func(in *ssm.GetParameterInput) (out *ssm.GetParameterOutput, err error) {
			err = errors.New("failed")
			return
		},
	}
	src = NewSSMSourceWithClient(&svc)
	l = NewLoader(src)
	err = l.Load(&sett)
	if err == nil {
		t.Fatal("expected to get an error, got nil")
	}
	if err.Error() != "failed" {
		t.Fatalf("expected to get error message 'failed', got: '%s'", err.Error())
	}

	// Test when SSM.GetParameter can't find parameter
	// when SSM can't find a parameter, it responds with ssm.ErrCodeParameterNotFound, we should treat this the same as if a variable is not set on the environment, i.e. return "" and no error from the getter so it can get caught by the default handling logic
	svc = mockSSM{
		getParameter: func(in *ssm.GetParameterInput) (out *ssm.GetParameterOutput, err error) {
			err = &ssm.ParameterNotFound{
				Message_: aws.String("failed"),
			}
			return
		},
	}
	src = NewSSMSourceWithClient(&svc)
	l = NewLoader(src)
	err = l.Load(&sett)
	if err == nil {
		t.Fatal("expected to get an error, got nil")
	}
	if err.Error() != "config: missing value for key 'test/parameter/name'" {
		t.Fatalf("unexpected error: '%s'", err.Error())
	}
}

func TestSSMSourceWithSubstitutions(t *testing.T) {
	testParameterValue := "parameter_value"
	sett := struct {
		TestParameter string `ssm:"project/$stage/parameter"`
	}{}
	svc := mockSSM{
		getParameter: func(in *ssm.GetParameterInput) (out *ssm.GetParameterOutput, err error) {
			if *in.Name != "project/prod/parameter" {
				t.Fatalf("expected parameter name to be %s, got %s", "project/prod/parameter", *in.Name)
			}
			if in.WithDecryption != nil && *in.WithDecryption == true {
				t.Fatalf("expected WithDecryption to be false, got %v", *in.WithDecryption)
			}

			out = &ssm.GetParameterOutput{
				Parameter: &ssm.Parameter{
					Value: &testParameterValue,
				},
			}
			return
		},
	}
	subs := map[string]string{
		"stage": "prod",
	}

	src := NewSSMSourceWithConfig(SSMSourceConfig{
		Service:       svc,
		Substitutions: subs,
	})
	l := NewLoader(src)
	err := l.Load(&sett)
	if err != nil {
		t.Fatalf("failed to load config with err: %v", err)
	}
	if sett.TestParameter != testParameterValue {
		t.Fatalf("expected value to be %s, got %s", testParameterValue, sett.TestParameter)
	}
}

func TestSSMSecureStrings(t *testing.T) {
	testParameterName := "test/parameter/name"
	testParameterValue := "parameter_value"
	sett := struct {
		TestParameter string `ssm:"test/parameter/name,secure"`
	}{}

	svc := mockSSM{
		getParameter: func(in *ssm.GetParameterInput) (out *ssm.GetParameterOutput, err error) {
			if *in.Name != testParameterName {
				t.Fatalf("expected parameter name to be %s, got %s", testParameterName, *in.Name)
			}
			if in.WithDecryption == nil {
				t.Fatal("expected input.WithDecryption to be true, got nil")
			}
			if *in.WithDecryption == false {
				t.Fatal("expected input.WithDecryption to be true, got false")
			}
			out = &ssm.GetParameterOutput{
				Parameter: &ssm.Parameter{
					Value: &testParameterValue,
				},
			}
			return
		},
	}

	src := NewSSMSourceWithClient(&svc)
	l := NewLoader(src)
	err := l.Load(&sett)
	if err != nil {
		t.Fatalf("failed to load config with err: %v", err)
	}
	if sett.TestParameter != testParameterValue {
		t.Fatalf("expected parameter value to be %s, got %s", testParameterValue, sett.TestParameter)
	}
}

type mockSSM struct {
	ssmiface.SSMAPI
	getParameter func(*ssm.GetParameterInput) (out *ssm.GetParameterOutput, err error)
}

func (ssm mockSSM) GetParameter(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	return ssm.getParameter(in)
}

func Test_getParamName(t *testing.T) {
	tests := []struct {
		name   string
		source string
		subs   map[string]string
		want   string
	}{
		{
			name:   "returns the passed string if it has no parameters",
			source: "project/prod/ultraSpeed",
			subs:   map[string]string{},
			want:   "project/prod/ultraSpeed",
		},
		{
			name:   "replaces all parameters with their corresponding values",
			source: "project/$stage/$feature",
			subs: map[string]string{
				"stage":   "prod",
				"feature": "ultraSpeed",
			},
			want: "project/prod/ultraSpeed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getParamName(tt.source, tt.subs)
			if got != tt.want {
				t.Errorf("getParamName() = %v, want %v", got, tt.want)
			}
		})
	}
}
