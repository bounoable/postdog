package smtp_test

import (
	"context"
	"testing"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/transport/smtp"
	mock_smtp "github.com/bounoable/postdog/transport/smtp/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	host     = "mail.example.com"
	port     = 25
	username = "bob"
	password = "secret"
	addr     = "mail.example.com:25"
)

func TestTransport_Send(t *testing.T) {
	tests := []struct {
		name         string
		letterOpts   []letter.Option
		assertSender func(letter.Letter, *mock_smtp.MockMailSender)
	}{
		{
			name: "single recipient",
			letterOpts: []letter.Option{
				letter.From("Bob Belcher", "bob@example.com"),
				letter.To("Linda Belcher", "linda@example.com"),
				letter.Subject("Hi."),
				letter.Text("Hello."),
				letter.HTML("<p>Hello.</p>"),
			},
			assertSender: func(let letter.Letter, s *mock_smtp.MockMailSender) {
				s.EXPECT().
					SendMail(addr, gomock.Any(), "bob@example.com", []string{"linda@example.com"}, []byte(let.RFC())).
					Return(nil)
			},
		},
		{
			name: "multiple recipients",
			letterOpts: []letter.Option{
				letter.From("Bob Belcher", "bob@example.com"),
				letter.To("Linda Belcher", "linda@example.com"),
				letter.CC("Tina Belcher", "tina@example.com"),
				letter.BCC("Gene Belcher", "gene@example.com"),
				letter.Subject("Hi."),
				letter.Text("Hello."),
				letter.HTML("<p>Hello.</p>"),
			},
			assertSender: func(let letter.Letter, s *mock_smtp.MockMailSender) {
				s.EXPECT().
					SendMail(addr, gomock.Any(), "bob@example.com", []string{
						"linda@example.com",
						"tina@example.com",
						"gene@example.com",
					}, []byte(let.RFC())).
					Return(nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			let, err := letter.TryWrite(test.letterOpts...)
			assert.Nil(t, err)

			s := mock_smtp.NewMockMailSender(ctrl)
			if test.assertSender != nil {
				test.assertSender(let, s)
			}

			tr := smtp.TransportWithSender(s, host, port, username, password)
			err = tr.Send(context.Background(), let)
			assert.Nil(t, err)
		})
	}
}
