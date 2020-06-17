<img src="./.github/logo.svg" width="100px">
<h1>postdog - GO mailing toolkit</h1>

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

> Visit the docs at [**go.dev**](https://pkg.go.dev/github.com/bounoable/postdog) for examples.

### Installation

```sh
go get github.com/bounoable/postdog
```

### Main packages

| Package                | Description                                                          |
|:-----------------------|:---------------------------------------------------------------------|
| [postdog](./)          | Queued sending and support for logging, middlewares, hooks & plugins |
| [letter](./letter)     | Provides the `Letter` type and write helpers                         |
| [autowire](./autowire) | Automatic `Office` setup through a YAML config file                  |

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
package main

import (
  "context"

  "github.com/bounoable/postdog/autowire"
  "github.com/bounoable/postdog/letter"
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

  po, err := cfg.Office(
    context.Background(),
    // Office options ... (plugins etc.)
  )
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
package main

import (
  "context"

  "github.com/bounoable/postdog"
  "github.com/bounoable/postdog/letter"
  "github.com/bounoable/postdog/transport/smtp"
  "github.com/bounoable/postdog/transport/gmail"
)

func main() {
  test := smtp.NewTransport("smtp.mailtrap.io", 587, "abcdef123456", "123456abcdef")
  prod, err := gmail.NewTransport(context.Background(), gmail.CredentialsFile("/path/to/google/service/account.json"))

  if err != nil {
    panic(err)
  }

  po := postdog.New()
  po.ConfigureTransport("test", test)
  po.ConfigureTransport("production", prod, postdog.DefaultTransport()) // make it the default transport

  err = po.Send(context.Background(), letter.Write())

  // or use transport directly
  err = prod.Send(context.Background(), letter.Write())
}
```

## Sending mails / letters

```go
package main

import (
  "context"

  "github.com/bounoable/postdog"
  "github.com/bounoable/postdog/autowire"
  "github.com/bounoable/postdog/letter"
  "github.com/bounoable/postdog/transport/smtp"
  "github.com/bounoable/postdog/transport/gmail"
)

func main() {
  // Load config
  cfg, err := autowire.File("/path/to/config.yml", smtp.Register, gmail.Register)
  if err != nil {
    panic(err)
  }

  // Initialize office
  po, err := cfg.Office(context.Background())
  if err != nil {
    panic(err)
  }

  // Send letter directly
  err = po.Send(context.Background(), letter.Write())

  // Queued sending

  // 1. Make sure postdog is running
  go po.Run(context.Background(), postdog.Workers(3))

  // 2. Dispatch the letter
  err = po.Dispatch(context.Background(), letter.Write())
}
```

## Plugins

You can extend `postdog` with plugins that register custom middleware and hooks:

| Plugin                         | Description                      |
|:-------------------------------|:---------------------------------|
| [Markdown](./plugin/markdown) | Markdown support in letters      |
| [Store](./plugin/store)       | Store sent letters in a database |
| [Template](./plugin/template) | Template support in letters      |

## Writing plugins

[Plugins](./plugin.go) have to provide a single `Install()` method that accepts a plugin context. Here is an example of a bad word filter:

```go
package main

import (
  "context"
  "strings"

  "github.com/bounoable/postdog"
  "github.com/bounoable/postdog/letter"
)

type badWordFilterPlugin struct {
  words []string
}

func (plug badWordFilterPlugin) Install(ctx postdog.PluginContext) {
  // register middleware
  ctx.WithMiddleware(
    postdog.MiddlewareFunc(
      ctx context.Context,
      let letter.Letter,
      next func(context.Context, letter.Letter) (letter.Letter, error),
    ) (letter.Letter, error) {
      for _, word := range plug.words {
        let.Text = strings.Replace(let.Text, word, "")
        let.HTML = strings.Replace(let.HTML, word, "")
      }

      // call the next middleware
      return next(ctx, let)
    }),
  )
}

func main() {
  po := postdog.New(
    postdog.WithPlugin(badWordFilterPlugin{
      words: []string{"very", "bad", "words"},
    }),
  )
}
```

You can also use a function as a plugin with `postdog.MiddlewareFunc`:

```go
package main

import (
  "context"
  "strings"

  "github.com/bounoable/postdog"
  "github.com/bounoable/postdog/letter"
)

func BadWordPlugin(words ...string) postdog.PluginFunc {
  return func(ctx postdog.PluginContext) {
    // register middleware
    ctx.WithMiddleware(
      postdog.MiddlewareFunc(
        ctx context.Context,
        let letter.Letter,
        next func(context.Context, letter.Letter) (letter.Letter, error),
      ) (letter.Letter, error) {
        for _, word := range plug.words {
          let.Text = strings.Replace(let.Text, word, "")
          let.HTML = strings.Replace(let.HTML, word, "")
        }

        // call the next middleware
        return next(ctx, let)
      }),
    )
  }
}

func main() {
  po := postdog.New(
    postdog.WithPlugin(
      BadWordPlugin("very", "bad", "words"),
    ),
  )
}
```
