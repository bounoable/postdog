package autowire_test

// func TestLoad(t *testing.T) {
// 	wd, err := os.Getwd()
// 	assert.Nil(t, err)

// 	office, err := autowire.File(filepath.Join(wd, "testdata/config.yml"))
// 	assert.Nil(t, err)

// 	cases := []struct {
// 		name     string
// 		provider string
// 		config   map[string]interface{}
// 	}{
// 		{
// 			name:     "test1",
// 			provider: "smtp",
// 			config: map[string]interface{}{
// 				"host":     "smtp.mailtrap.io",
// 				"username": "abcdef123456",
// 				"password": "123456abcdef",
// 			},
// 		},
// 		{
// 			name:     "test2",
// 			provider: "gmail",
// 			config: map[string]interface{}{
// 				"serviceAccount": "/path/to/service_account.json",
// 				"scopes": []interface{}{
// 					"https://www.googleapis.com/auth/gmail.addons.current.action.compose",
// 					"https://www.googleapis.com/auth/gmail.send",
// 				},
// 			},
// 		},
// 	}

// 	for _, tcase := range cases {

// 		// transportcfg, ok := cfg.Tr
// 		// carriercfg, ok := cfg.Carriers[test.name]

// 		// assert.True(t, ok)
// 		// assert.Equal(t, test.provider, carriercfg.Provider)
// 		// assert.Equal(t, test.config, carriercfg.Config)
// 	}

// 	assert.Equal(t, "test2", cfg.DefaultCarrier)
// }

// func TestConfig_Load(t *testing.T) {
// 	wd, err := os.Getwd()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	cfg := office.NewConfig()

// 	err = cfg.Load(filepath.Join(wd, "testdata/config.yml"))
// 	assert.Nil(t, err)

// 	tests := []struct {
// 		name     string
// 		provider string
// 		config   map[string]interface{}
// 	}{
// 		{
// 			name:     "test1",
// 			provider: "smtp",
// 			config: map[string]interface{}{
// 				"host":     "smtp.mailtrap.io",
// 				"username": "abcdef123456",
// 				"password": "123456abcdef",
// 			},
// 		},
// 		{
// 			name:     "test2",
// 			provider: "gmail",
// 			config: map[string]interface{}{
// 				"serviceAccount": "/path/to/service_account.json",
// 				"scopes": []interface{}{
// 					"https://www.googleapis.com/auth/gmail.addons.current.action.compose",
// 					"https://www.googleapis.com/auth/gmail.send",
// 				},
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		carriercfg, ok := cfg.Carriers[test.name]

// 		assert.True(t, ok)
// 		assert.Equal(t, test.provider, carriercfg.Provider)
// 		assert.Equal(t, test.config, carriercfg.Config)
// 	}

// 	assert.Equal(t, "test2", cfg.DefaultCarrier)
// }

// func TestNewOffice(t *testing.T) {
// 	cfg := postdog.NewConfig()

// 	cfg.RegisterFactory("test", postdog.CarrierFactoryFunc(testCarrierFactory))
// 	cfg.Configure("main", "test", map[string]interface{}{
// 		"key1": "value1",
// 		"key2": 2,
// 		"key3": true,
// 	})

// 	m, err := cfg.NewOffice(context.Background())
// 	assert.Nil(t, err)

// 	carrier, err := m.Carrier("main")
// 	assert.Nil(t, err)

// 	tcarrier, ok := carrier.(testCarrier)
// 	assert.True(t, ok)

// 	assert.Equal(t, "value1", tcarrier.Key1)
// 	assert.Equal(t, 2, tcarrier.Key2)
// 	assert.True(t, tcarrier.Key3)
// }

// type testCarrier struct {
// 	Key1 string
// 	Key2 int
// 	Key3 bool
// }

// func (d testCarrier) Send(_ context.Context, let *postdog.Letter) error { return nil }

// func testCarrierFactory(_ context.Context, cfg map[string]interface{}) (postdog.Carrier, error) {
// 	return testCarrier{
// 		Key1: cfg["key1"].(string),
// 		Key2: cfg["key2"].(int),
// 		Key3: cfg["key3"].(bool),
// 	}, nil
// }
