package letter_test

import (
	"bytes"
	"io/ioutil"
	"net/mail"
	"os"
	"path/filepath"
	"testing"

	"github.com/bounoable/postdog/letter"
	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	recipeText, err := os.Open(filepath.Join(wd, "testdata/burger-recipe.txt"))
	if err != nil {
		panic(err)
	}
	recipeTextBytes, err := ioutil.ReadAll(recipeText)
	if err != nil {
		panic(err)
	}

	recipeHTML, err := os.Open(filepath.Join(wd, "testdata/burger-recipe.html"))
	if err != nil {
		panic(err)
	}
	recipeHTMLBytes, err := ioutil.ReadAll(recipeHTML)
	if err != nil {
		panic(err)
	}

	let := letter.Write(
		letter.Subject("Where's the rent?"),
		letter.FromAddress(mail.Address{
			Name:    "Calvin",
			Address: "calvin@fischoeder.test",
		}),
		letter.ToAddress(
			mail.Address{
				Name:    "Bob",
				Address: "bob@belcher.test",
			},
			mail.Address{
				Name:    "Linda",
				Address: "linda@belcher.test",
			},
		),
		letter.ToAddress(mail.Address{
			Name:    "Gene",
			Address: "gene@belcher.test",
		}),
		letter.CCAddress(
			mail.Address{
				Name:    "Tina",
				Address: "tina@belcher.test",
			},
			mail.Address{
				Name:    "Louise",
				Address: "louise@belcher.test",
			},
		),
		letter.BCCAddress(
			mail.Address{
				Name:    "Jimmy",
				Address: "jimmy@pesto.test",
			},
		),
		letter.HTML("<p>You are terrible at what you do.</p>"),
		letter.Text("You are terrible at what you do."),
		letter.MustAttach(bytes.NewReader([]byte{1, 2, 3}), "burger recipe.html"),
		letter.MustAttach(bytes.NewReader([]byte{1, 2, 3}), "Burger Recipe"),
		letter.MustAttach(bytes.NewReader([]byte{1, 2, 3}), "Burger Recipe", letter.ContentType("text/plain")),
		letter.MustAttachFile(filepath.Join(wd, "testdata/burger-recipe.txt"), "Burger Recipe"),
		letter.MustAttachFile(filepath.Join(wd, "testdata/burger-recipe.html"), "Burger Recipe", letter.ContentType("image/jpeg")),
	)

	assert.Nil(t, err)

	assert.Equal(t, "Where's the rent?", let.Subject)

	assert.Equal(t, mail.Address{
		Name:    "Calvin",
		Address: "calvin@fischoeder.test",
	}, let.From)

	assert.Equal(t, []mail.Address{
		{
			Name:    "Bob",
			Address: "bob@belcher.test",
		},
		{
			Name:    "Linda",
			Address: "linda@belcher.test",
		},
		{
			Name:    "Gene",
			Address: "gene@belcher.test",
		},
	}, let.To)

	assert.Equal(t, []mail.Address{
		{
			Name:    "Tina",
			Address: "tina@belcher.test",
		},
		{
			Name:    "Louise",
			Address: "louise@belcher.test",
		},
	}, let.CC)

	assert.Equal(t, []mail.Address{{
		Name:    "Jimmy",
		Address: "jimmy@pesto.test",
	}}, let.BCC)

	assert.Equal(t, "<p>You are terrible at what you do.</p>", let.HTML)
	assert.Equal(t, "You are terrible at what you do.", let.Text)

	assert.Equal(t, []byte{1, 2, 3}, let.Attachments[0].Content)
	assert.Equal(t, "burger recipe.html", let.Attachments[0].Filename)
	assert.Contains(t, let.Attachments[0].Header.Get("Content-Type"), "text/html")

	assert.Equal(t, []byte{1, 2, 3}, let.Attachments[1].Content)
	assert.Equal(t, "Burger Recipe", let.Attachments[1].Filename)
	assert.Contains(t, let.Attachments[1].Header.Get("Content-Type"), "application/octet-stream")

	assert.Equal(t, []byte{1, 2, 3}, let.Attachments[2].Content)
	assert.Equal(t, "Burger Recipe", let.Attachments[2].Filename)
	assert.Contains(t, let.Attachments[2].Header.Get("Content-Type"), "text/plain")

	assert.Equal(t, recipeTextBytes, let.Attachments[3].Content)
	assert.Equal(t, "Burger Recipe", let.Attachments[3].Filename)
	assert.Contains(t, let.Attachments[3].Header.Get("Content-Type"), "text/plain")

	assert.Equal(t, recipeHTMLBytes, let.Attachments[4].Content)
	assert.Equal(t, "Burger Recipe", let.Attachments[4].Filename)
	assert.Contains(t, let.Attachments[4].Header.Get("Content-Type"), "image/jpeg")
}
