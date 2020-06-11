package office_test

import (
	"errors"
	"testing"

	"github.com/bounoable/postdog/office"
	"github.com/bounoable/postdog/office/mock_office"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestOffice_Configure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	off := office.New()
	_, err := off.Transport("test")

	assert.True(t, errors.Is(err, office.UnconfiguredTransportError{
		Name: "test",
	}))

	_, err = off.DefaultTransport()

	assert.True(t, errors.Is(err, office.UnconfiguredTransportError{}))

	mockTrans := mock_office.NewMockTransport(ctrl)
	off.Configure("test", mockTrans)
	trans, err := off.Transport("test")

	assert.Nil(t, err)
	assert.Equal(t, mockTrans, trans)

	defaultTrans, err := off.DefaultTransport()

	assert.Nil(t, err)
	assert.Equal(t, mockTrans, defaultTrans)
}

func TestOffice_Configure_asDefault(t *testing.T) {
	cases := map[string]struct {
		configure func(*office.Office, *gomock.Controller)
		expected  string
	}{
		"default default": {
			configure: func(off *office.Office, ctrl *gomock.Controller) {
				off.Configure("test1", mock_office.NewMockTransport(ctrl))
				off.Configure("test2", mock_office.NewMockTransport(ctrl))
				off.Configure("test3", mock_office.NewMockTransport(ctrl))
			},
			expected: "test1",
		},
		"first as default": {
			configure: func(off *office.Office, ctrl *gomock.Controller) {
				off.Configure("test1", mock_office.NewMockTransport(ctrl), office.DefaultTransport())
				off.Configure("test2", mock_office.NewMockTransport(ctrl))
				off.Configure("test3", mock_office.NewMockTransport(ctrl))
			},
			expected: "test1",
		},
		"other-than-first as default": {
			configure: func(off *office.Office, ctrl *gomock.Controller) {
				off.Configure("test1", mock_office.NewMockTransport(ctrl))
				off.Configure("test2", mock_office.NewMockTransport(ctrl), office.DefaultTransport())
				off.Configure("test3", mock_office.NewMockTransport(ctrl))
			},
			expected: "test2",
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			off := office.New()
			tcase.configure(off, ctrl)

			expected, err := off.Transport(tcase.expected)
			assert.Nil(t, err)
			trans, err := off.DefaultTransport()
			assert.Nil(t, err)

			assert.Equal(t, expected, trans)
		})
	}
}

func TestOffice_MakeDefault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	off := office.New()
	off.Configure("test1", mock_office.NewMockTransport(ctrl))
	off.Configure("test2", mock_office.NewMockTransport(ctrl))
	off.Configure("test3", mock_office.NewMockTransport(ctrl))

	assertDefaultTransport(t, off, "test1")

	err := off.MakeDefault("test1")
	assert.Nil(t, err)
	assertDefaultTransport(t, off, "test1")

	err = off.MakeDefault("test2")
	assert.Nil(t, err)
	assertDefaultTransport(t, off, "test2")

	err = off.MakeDefault("test3")
	assert.Nil(t, err)
	assertDefaultTransport(t, off, "test3")

	err = off.MakeDefault("test4")
	assert.True(t, errors.Is(err, office.UnconfiguredTransportError{Name: "test4"}))
}

func assertDefaultTransport(t *testing.T, off *office.Office, name string) {
	trans, err := off.DefaultTransport()
	assert.Nil(t, err)

	expected, err := off.Transport(name)
	assert.Nil(t, err)

	assert.Equal(t, expected, trans)
}
