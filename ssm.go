package config

import (
	"fmt"
	"strings"

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

func (s *ssmSource) Get(tag tagValue) (string, error) {
	name := tag.name
	_, isSecure := tag.flags[flagSecure]
	if s.subs != nil {
		name = getParamName(tag.name, s.subs)
	}
	out, err := s.svc.GetParameter(&ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: &isSecure,
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

func getParamName(source string, subs map[string]string) (result string) {
	result = source
	for k, v := range subs {
		sub := fmt.Sprintf("$%s", k)
		result = strings.ReplaceAll(result, sub, v)
	}
	return
}
