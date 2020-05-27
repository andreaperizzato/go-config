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
	return NewSSMSourceWithConfig(SSMSourceConfig{
		Service: svc,
	})
}

type ssmSource struct {
	svc  ssmiface.SSMAPI
	subs map[string]string
}

var _ssmSourceIfaceCheck Source = &ssmSource{}

func (s *ssmSource) Tag() string {
	return SSMTag
}

func (s *ssmSource) Get(key string) (string, error) {
	var err error
	if s.subs != nil {
		key, err = getParamName(key, s.subs)
		if err != nil {
			return "", err
		}
	}
	out, err := s.svc.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(key),
	})
	if err != nil {
		return "", err
	}
	return *out.Parameter.Value, nil
}

// NewSSMSourceWithSubstitutions creates a source for values stored in SSM with a map of substitutions.
func NewSSMSourceWithSubstitutions(subs map[string]string) Source {
	return NewSSMSourceWithConfig(SSMSourceConfig{
		Service:       ssm.New(session.New()),
		Substitutions: subs,
	})
}

// SSMSourceConfig is the configuration for the creation of an SSMSource.
type SSMSourceConfig struct {
	Service       ssmiface.SSMAPI
	Substitutions map[string]string
}

// NewSSMSourceWithConfig create a Source for a values stored in SSM specifying custom configuration.
func NewSSMSourceWithConfig(cfg SSMSourceConfig) Source {
	return &ssmSource{
		svc:  cfg.Service,
		subs: cfg.Substitutions,
	}
}

func getParamName(source string, subs map[string]string) (result string, err error) {
	paramsCount := strings.Count(source, "$")
	re := regexp.MustCompile("\\$\\w+")
	params := re.FindAll([]byte(source), paramsCount)
	result = source
	for _, param := range params {
		name := string(param[1:]) // remove $
		val, found := subs[name]
		if !found {
			return "", fmt.Errorf("could not find substitution for parameter '%s'", name)
		}
		result = strings.Replace(result, string(param), val, 1)
	}
	return
}
