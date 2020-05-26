package config

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

// SSMTag is the name of the tag to load variables from SSM.
const SSMTag = "ssm"

// NewSSMSource create a source for values stored in SSM.
func NewSSMSource() Source {
	svc := ssm.New(session.New())
	return NewSSMSourceWithClient(svc)
}

// NewSSMSourceWithClient create a Source for a values stored in SSM.
func NewSSMSourceWithClient(svc ssmiface.SSMAPI) Source {
	return &source{
		tag: SSMTag,
		get: getFromSSM(svc),
	}
}

func getFromSSM(svc ssmiface.SSMAPI) func(key string) (string, error) {
	return func(key string) (string, error) {
		out, err := svc.GetParameter(&ssm.GetParameterInput{
			Name: aws.String(key),
		})
		if err != nil {
			return "", err
		}
		return *out.Parameter.Value, nil
	}
}
