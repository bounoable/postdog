package postdog_test

import (
	"bytes"
	"context"
	"strings"

	"github.com/bounoable/postdog/autowire"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/office"
	"github.com/bounoable/postdog/transport/gmail"
	"github.com/bounoable/postdog/transport/smtp"
)

// Configure postdog from YAML file.
func Example() {
	// Load YAML config
	cfg, err := autowire.File("/path/to/config.yml")
	if err != nil {
		panic(err)
	}

	// Build office
	off, err := cfg.Office(context.Background())
	if err != nil {
		panic(err)
	}

	// Send mail with default transport
	err = off.Send(
		context.Background(),
		letter.Write(
			letter.From("Bob", "bob@belcher.test"),
			letter.To("Calvin", "calvin@fishoeder.test"),
			letter.To("Felix", "felix@fishoeder.test"),
			letter.BCC("Jimmy", "jimmy@pesto.test"),
			letter.Subject("Hi, buddy."),
			letter.Text("Have a drink later?"),
			letter.HTML("Have a <strong>drink</strong> later?"),
			letter.MustAttach(bytes.NewReader([]byte("secret")), "my_burger_recipe.txt"),
		),
	)

	// or use a specific transport
	err = off.SendWith(
		context.Background(),
		"mytransport",
		letter.Write(
			letter.From("Bob", "bob@belcher.test"),
			// ...
		),
	)
}

func Example_manualConfiguration() {
	po := office.New()

	smtpTransport := smtp.NewTransport("smtp.mailtrap.io", 587, "abcdef123456", "123456abcdef")

	gmailTransport, err := gmail.NewTransport(
		context.Background(),
		gmail.CredentialsFile("/path/to/service_account.json"),
	)
	if err != nil {
		panic(err)
	}

	po.ConfigureTransport("mytransport1", smtpTransport)
	po.ConfigureTransport("mytransport2", gmailTransport, office.DefaultTransport()) // Make "mytransport2" the default

	err = po.Send(
		context.Background(),
		letter.Write(
			letter.From("Bob", "bob@belcher.test"),
			// ...
		),
	)
}

func Example_useTransportDirectly() {
	trans, err := gmail.NewTransport(
		context.Background(),
		gmail.CredentialsFile("/path/to/service_account.json"),
	)
	if err != nil {
		panic(err)
	}

	let := letter.Write(
		letter.From("Bob", "bob@belcher.test"),
		// ...
	)

	if err := trans.Send(context.Background(), let); err != nil {
		panic(err)
	}
}

func Example_middleware() {
	off := office.New(
		office.WithMiddleware(
			office.MiddlewareFunc(func(
				ctx context.Context,
				let letter.Letter,
				next func(context.Context, letter.Letter) (letter.Letter, error),
			) (letter.Letter, error) {
				let.Subject = strings.Title(let.Subject)
				return next(ctx, let)
			}),
		),
	)

	if err := off.Send(context.Background(), letter.Write(
		letter.Subject("this is a title"), // will be set to "This Is A Title" by the middleware before sending
	)); err != nil {
		panic(err)
	}
}
