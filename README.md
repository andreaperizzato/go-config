# Config

**Config** provides a simple way to load configuration values from different sources using tags. 

```go
package main

import (
	"fmt"
	"log"

	"github.com/andreaperizzato/go-config"
)

type Settings struct {
	BaseURL        string `env:"BASE_URL"`
	LogLevel       string `env:"LOG_LEVEL,optional"`
	MaxConnections int8   `env:"MAX_CONNECTIONS" default:"2"`
	DisableSSL     bool   `env:"DISABLE_SSL,optional"`
	JWTSecret      string `ssm:"jwt_secret"`
}

func main() {
	l := config.NewLoader(
		// Load from the environment (tag "env").
		config.NewEnvSource(),
		// Load from AWS SSM (tag "ssm").
		config.NewSSMSource(),
	)

	var s Settings
	err := l.Load(&s)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(s)
}
```

## Features

- Load values from the environment
- Load values from AWS SSM
- Supports `string`, `int8`, `int16`, `int32`, `int64`
- Supports `bool` setting `true` when the value is `"true"` or `"1"`
- Fields can be optional
- Fields can have default values
- Add your own sources

## Available sources

### Environment

```go
s := config.NewEnvSource()
```

creates a new `Source` that loads values from the environment, using `os.Getenv`.

Tag with `env` to load values from the environment.

### AWS SSM (Amazon Simple Systems Manager)

```go
// Using the default session and SSM client.
s := config.NewSSMSource()

// Using a custom client and session.
svc := // create your SSM client
s = config.NewSSMSourceWithClient(svc)
```

creates a new `Source` that loads values from SSM using [GetParameter](https://docs.aws.amazon.com/sdk-for-go/api/service/ssm/#SSM.GetParameter). Note that you must have permissions to get the parameters from SSM.

Tag with `ssm` to load values from SSM.

## Custom sources

A `Source` is an interface that loads values from a location:

```go
type Source interface {
	Tag() string
	Get(key string) (string, error)
}
```

where:

```go 
func Tag() string
```

returns the name of the tag linked to this source. For instance, [EnvSource](./env.go) returns `"env"` so that fields tagged with `"env:NAME"` will be loaded using this source.

```go 
func Get(key string) (string, error)
```

loads the value for a key.

## Contributing

Thank you for considering contributing! Please use GitHub issues and Pull Requests for contributing.

## License

The MIT License (MIT). Please see License File for more information.
