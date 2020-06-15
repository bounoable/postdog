package autowire_test

import (
	"bytes"
	"context"

	"github.com/bounoable/postdog/autowire"
	"github.com/bounoable/postdog/letter"
)

func Example() {
	// Load YAML config
	cfg, err := autowire.File("/path/to/config.yml")
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
			letter.Text("Have a drink later?"),
			letter.HTML("Have a <strong>drink</strong> later?"),
			letter.MustAttach(bytes.NewReader([]byte{1, 2, 3}), "My burger recipe"),
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
