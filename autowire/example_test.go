package autowire_test

import (
	"bytes"
	"context"
	"strings"

	"github.com/bounoable/postdog/autowire"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/markdown"
	"github.com/bounoable/postdog/transport/smtp"
)

func Example() {
	config := `
transports:
  dev:
	provider: smtp
	config:
	  host: smtp.mailtrap.io
	  username: abcdef123456
	  password: 123456abcdef

plugins:
  - name: markdown
	  config:
		use: goldmark
		overrideHTML: true
`

	// Load YAML config
	cfg, err := autowire.Load(
		strings.NewReader(config),
		smtp.Register,     // register SMTP transport factory
		markdown.Register, // register Markdown plugin factory
	)
	if err != nil {
		panic(err)
	}

	// Initialize office
	po, err := cfg.Office(
		context.Background(),
		// Office options ...
	)
	if err != nil {
		panic(err)
	}

	// Send mail with default transport
	err = po.Send(
		context.Background(),
		letter.Write(
			letter.From("Bob", "bob@belcher.test"),
			letter.To("Calvin", "calvin@fishoeder.test"),
			letter.To("Felix", "felix@fishoeder.test"),
			letter.BCC("Jimmy", "jimmy@pesto.test"),
			letter.Subject("Hi, buddy."),
			letter.Text("# Have a drink later?"),
			letter.MustAttach(bytes.NewReader([]byte("tasty")), "burgerrecipe.txt", letter.ContentType("text/plain")),
		),
	)

	// or use a specific transport
	err = po.SendWith(
		context.Background(),
		"test2",
		letter.Write(
			letter.From("Bob", "bob@belcher.test"),
			// ...
		),
	)
}
