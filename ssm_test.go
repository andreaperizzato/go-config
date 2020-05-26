package config

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

func TestSSMSource(t *testing.T) {
	testParameterName := "test/parameter/name"
	testParameterValue := "parameter_value"
	client := mockSSM{
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
	s := NewSSMSourceWithClient(&client)

	l := NewLoader(s)
	sett := struct {
		TestParameter string `ssm:"test/parameter/name"`
	}{}

	err := l.Load(&sett)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if sett.TestParameter != testParameterValue {
		t.Fatalf("expected value to be %s, got %s", testParameterValue, sett.TestParameter)
	}
}

func TestSSMSourceWithSubstitutions(t *testing.T) {
	testParameterValue := "parameter_value"
	client := mockSSM{
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
	s := NewSSMSourceWithSubstitutionsWithClient(&client, subs)

	l := NewLoader(s)
	sett := struct {
		TestParameter string `ssm:"project/$stage/parameter"`
	}{}

	err := l.Load(&sett)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if sett.TestParameter != testParameterValue {
		t.Fatalf("expected value to be %s, got %s", testParameterValue, sett.TestParameter)
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
