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

type mockSSM struct {
	ssmiface.SSMAPI
	getParameter func(*ssm.GetParameterInput) (out *ssm.GetParameterOutput, err error)
}

func (ssm mockSSM) GetParameter(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	return ssm.getParameter(in)
}
