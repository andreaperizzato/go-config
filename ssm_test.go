package config

import (
	"errors"
	"testing"

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

	// Test error with substitutions
	svc = mockSSM{
		getParameter: func(in *ssm.GetParameterInput) (out *ssm.GetParameterOutput, err error) {
			t.Fatal("test should not reach this point")
			return
		},
	}
	subs = map[string]string{}

	src = NewSSMSourceWithConfig(SSMSourceConfig{
		Service:       svc,
		Substitutions: subs,
	})
	l = NewLoader(src)
	err = l.Load(&sett)
	if err == nil {
		t.Fatal("expected to get an error, got nil")
	}
	if err.Error() != "could not find substitution for parameter 'stage'" {
		t.Fatalf("expected to get error message 'failed', got: '%s'", err.Error())
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
		name    string
		source  string
		subs    map[string]string
		want    string
		wantErr bool
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
		{
			name:   "returns an error if it can't find the value for a parameter",
			source: "project/$stage/$feature",
			subs: map[string]string{
				"stage": "prod",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getParamName(tt.source, tt.subs)
			if (err != nil) != tt.wantErr {
				t.Errorf("getParamName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getParamName() = %v, want %v", got, tt.want)
			}
		})
	}
}