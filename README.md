<center>
<img src="./.github/logo.svg" width="100px">
<h1>postdog - GO mailing toolkit</h1>
</center>

<p>
  <a href="https://pkg.go.dev/github.com/bounoable/postdog">
    <img alt="GoDoc" src="https://img.shields.io/badge/godoc-reference-purple">
  </a>
  <img alt="Version" src="https://img.shields.io/github/v/tag/bounoable/postdog" />
  <a href="#" target="_blank">
    <img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-blue.svg" />
  </a>
</p>

`postdog` is a mailing toolkit for GO applications, providing:

- Simple mail writing
- Sending mails with support for different transports (SMTP, Gmail etc.)
- Middleware
- Hooks
- Configuration via YAML
- Plugin support

## Getting Started

> Visit the docs at [**go.dev**](https://pkg.go.dev/github.com/bounoable/postdog) for more examples.

### Installation

```sh
go get github.com/bounoable/postdog
```

### Main packages

- [letter](./letter) provides the `Letter` type and write helpers
- [office](./office) provides the `Office` type for queued sending and support for logging, middlewares, hooks & plugins
- [autowire](./autowire) provides automatic `Office` configuration through a YAML config file

### Configuration

You can configure your transports in a YAML config file or [configure them manually](#manual).

#### YAML

```yaml
# /app/configs/postdog.yml

default: test # Set the default transport

transports:
  test: # Specify a unique name
    provider: smtp # Set the transport provider
    config: # Provider configuration
      host: smtp.mailtrap.io
      port: 587
      username: abcdef123456
      password: 123456abcdef
  
  production:
    provider: gmail
    config:
      serviceAccount: ${SERVICE_ACCOUNT_PATH} # Use environment variable
```

```go
import (
  "github.com/bounoable/postdog/autowire"
  "github.com/bounoable/postdog/transport/smtp"
  "github.com/bounoable/postdog/transport/gmail"
)

func main() {
  // Load the configuration
  cfg, err := autowire.File(
    "/app/configs/postdog.yml",
    // Register the SMTP and Gmail providers.
    smtp.Register,
    gmail.Register,
  )
  if err != nil {
    panic(err)
  }

  po, err := cfg.Office(context.Background())
  if err != nil {
    panic(err)
  }

  // Send via the default transport ("test")
  if err = po.Send(context.Background(), letter.Write(
    letter.From("Bob", "bob@example.com"),
    letter.To("Linda", "linda@example.com"),
    // ...
  )); err != nil {
    panic(err)
  }

  // Send via a specific transport ("production")
  if err = po.SendWith(context.Background(), "production", letter.Write(
    letter.From("Bob", "bob@example.com"),
    letter.To("Linda", "linda@example.com"),
    // ...
  )); err != nil {
    panic(err)
  }
}
```

#### Manual

You can also configure the transports manually or use them directly:

```go
func main() {
  test := smtp.NewTransport("smtp.mailtrap.io", 587, "abcdef123456", "123456abcdef")
  prod, err := gmail.NewTransport(context.Background(), gmail.CredentialsFile("/path/to/google/service/account.json"))

  if err != nil {
    panic(err)
  }

  po := office.New()
  po.ConfigureTransport("test", test)
  po.ConfigureTransport("production", prod, office.DefaultTransport()) // make it the default transport

  err = po.Send(context.Background(), letter.Write())
}
```

## Plugins

