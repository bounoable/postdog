package autowire_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/office"
	"github.com/bounoable/postdog/office/autowire"
	"github.com/bounoable/postdog/transport/gmail"
	"github.com/bounoable/postdog/transport/smtp"
	"github.com/stretchr/testify/assert"
)

func TestConfig_LoadFile(t *testing.T) {
	wd, _ := os.Getwd()

	cfg := autowire.New(
		smtp.Register,
		gmail.Register,
	)

	err := cfg.LoadFile(filepath.Join(wd, "testdata/config.yml"))
	assert.Nil(t, err)

	cases := []struct {
		name     string
		provider string
		config   map[string]interface{}
	}{
		{
			name:     "test1",
			provider: smtp.Provider,
			config: map[string]interface{}{
				"host":     "smtp.mailtrap.io",
				"username": "abcdef123456",
				"password": "123456abcdef",
			},
		},
		{
			name:     "test2",
			provider: gmail.Provider,
			config: map[string]interface{}{
				"serviceAccount": "/path/to/service_account.json",
				"scopes": []interface{}{
					"https://www.googleapis.com/auth/gmail.addons.current.action.compose",
					"https://www.googleapis.com/auth/gmail.send",
				},
			},
		},
	}

	for _, tcase := range cases {
		transportcfg, err := cfg.Get(tcase.name)
		assert.Nil(t, err)
		assert.Equal(t, tcase.provider, transportcfg.Provider)
		assert.Equal(t, tcase.config, transportcfg.Config)
	}
}

func TestConfig_Office(t *testing.T) {
	cfg := autowire.New()
	cfg.RegisterProvider("test", autowire.TransportFactoryFunc(
		func(ctx context.Context, cfg map[string]interface{}) (office.Transport, error) {
			return testTransport{val: cfg["val"].(int)}, nil
		},
	))
	cfg.Transports["test1"] = autowire.TransportConfig{
		Provider: "test",
		Config:   map[string]interface{}{"val": 1},
	}
	cfg.Transports["test2"] = autowire.TransportConfig{
		Provider: "test",
		Config:   map[string]interface{}{"val": 2},
	}

	off, err := cfg.Office(context.Background())
	assert.Nil(t, err)

	trans1, err := off.Transport("test1")
	assert.Nil(t, err)

	trans2, err := off.Transport("test2")
	assert.Nil(t, err)

	assert.Equal(t, testTransport{val: 1}, trans1)
	assert.Equal(t, testTransport{val: 2}, trans2)
}

type testTransport struct {
	val int
}

func (tt testTransport) Send(ctx context.Context, let *letter.Letter) error {
	return nil
}
