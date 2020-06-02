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
	LogLevel       string `env:"LOG_LEVEL"`
	MaxConnections int8   `env:"MAX_CONNECTIONS" default:"2"`
	DisableSSL     bool   `env:"DISABLE_SSL" default:""`
	JWTSecret      string `ssm:"jwt_secret" env:"JWT_SECRET"`
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

## Reading values

You can specify multiple sources for the same variable:

```go
type Settings struct {
	A string 	`env:"VALUE_A" ssm:"value_a"`
	B int 		`ssm:"value_b" env:"VALUE_B" default:""`
	C string 	`env:"VALUE_C" default:"hello"`
}
```

The order in which the parameters will be loaded is defined by how you setup the sources. For instance:

```go
l := config.NewLoader(config.NewSSMSource(), config.NewEnvSource())
```

will load the values from SSM first and then use the environment. If you consider the type above:

- `A`: will load the parameter `value_a` from SSM, then `VALUE_A` from the environment. If `VALUE_A` is set,
  it will replace the initial value set from SSM. If both `value_a` and `VALUE_A` are not set, the loader will fail since there is no default.
- `B`: same as before (not that the order of the tags is irrelevant). The difference is that if both `value_b` and `VALUE_B` are not set, `B` will be
  set to the zero value (0 in this case) and the loader won't fail. This is how you can define something as optional.
- `C`: will get `VALUE_C` from the environment and if not set, will use `hello` as default.

### Deprecated optional flag

Earlier version of this package supported an `optional` flag to denote that a source was not required. This flag is not deprecated and should be replaced with `default:""`:

```go
type OldSettings struct {
	Key string `env:"VALUE,optional"`
}

type NewSettings struct {
	Key string `env:"VALUE" default:""`
}
```

This is because it now supports multiple sources and the following would be undefined:

```go
type Settings struct {
	Key string `env:"VALUE,optional" ssm:"value"`
}
```

In order to be retro-compatible, you can still use the `optional` flag and it will be taken into account only with one source.

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

Tag with `ssm` to load values from SSM. You can use `secure` to load a secure string:

```go
type Settings struct {
	Key string `ssm:"api_key,secure"`
}
```

## Custom sources

A `Source` is an interface that loads values from a location:

```go
type Source interface {
	Tag() string
	Get(TagValue) (string, error)
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
