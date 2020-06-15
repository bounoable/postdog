package letter_test

import (
	"bytes"

	"github.com/bounoable/postdog/letter"
)

func ExampleWrite() {
	let := letter.Write(
		letter.From("Bob", "bob@belcher.test"),
		letter.To("Calvin", "calvin@fishoeder.test"),
		letter.To("Felix", "felix@fishoeder.test"),
		letter.BCC("Jimmy", "jimmy@pesto.test"),
		letter.Subject("Hi, buddy."),
		letter.Text("Have a drink later?"),
		letter.HTML("Have a <strong>drink</strong> later?"),
		letter.MustAttach(bytes.NewReader([]byte("tasty")), "burgerrecipe.txt", letter.ContentType("text/plain")),
	)

	_ = let
}
