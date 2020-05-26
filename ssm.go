package config

import (
	"fmt"
	"regexp"
	"strings"

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

// NewSSMSourceWithSubstitutions does stuff
// TODO temporary API
func NewSSMSourceWithSubstitutions(subs map[string]string) Source {
	svc := ssm.New(session.New())
	return NewSSMSourceWithSubstitutionsWithClient(svc, subs)
}

// NewSSMSourceWithSubstitutionsWithClient does stuff
// TODO temporary API
// I don't like the name of this function. Maybe we could make the svc part of the source struct and make getFromSSMWithSubstitutions a method?
func NewSSMSourceWithSubstitutionsWithClient(svc ssmiface.SSMAPI, subs map[string]string) Source {
	return &source{
		tag: SSMTag,
		get: getFromSSMWithSubstitutions(svc, subs),
	}
}

func getFromSSMWithSubstitutions(svc ssmiface.SSMAPI, subs map[string]string) func(key string) (string, error) {
	return func(key string) (string, error) {
		name, err := getParamName(key, subs)
		if err != nil {
			return "", err
		}
		out, err := svc.GetParameter(&ssm.GetParameterInput{
			Name: aws.String(name),
		})
		if err != nil {
			return "", err
		}
		return *out.Parameter.Value, nil
	}
}

func getParamName(source string, subs map[string]string) (result string, err error) {
	paramsCount := strings.Count(source, "$")
	re := regexp.MustCompile("\\$\\w+")
	params := re.FindAll([]byte(source), paramsCount)
	result = source
	for _, param := range params {
		name := string(param[1:])
		val, found := subs[name]
		if !found {
			return "", fmt.Errorf("failed")
		}
		result = strings.Replace(result, string(param), val, 1)
	}
	return
}
