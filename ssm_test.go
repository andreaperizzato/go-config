package config

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/stretchr/testify/assert"
)

func TestSSMSource(t *testing.T) {
	client := mockSSM{
		getParameter: func(in *ssm.GetParameterInput) (out *ssm.GetParameterOutput, err error) {
			assert.Equal(t, "dunderMifflin/michaelScott/catchPhrase", *in.Name)

			out = &ssm.GetParameterOutput{
				Parameter: &ssm.Parameter{
					Value: aws.String("That's what she said"),
				},
			}
			return
		},
	}
	s := NewSSMSourceWithClient(&client)

	l := NewLoader(s)
	sett := struct {
		TestParameter string `ssm:"dunderMifflin/michaelScott/catchPhrase"`
	}{}

	err := l.Load(&sett)
	assert.NoError(t, err)
	assert.Equal(t, "That's what she said", sett.TestParameter)
}

type mockSSM struct {
	ssmiface.SSMAPI
	getParameter func(*ssm.GetParameterInput) (out *ssm.GetParameterOutput, err error)
}

func (ssm mockSSM) GetParameter(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	return ssm.getParameter(in)
}
