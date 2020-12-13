package config_test

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/bounoable/postdog/config"
	mock_config "github.com/bounoable/postdog/config/mocks"
	mock_postdog "github.com/bounoable/postdog/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	Convey("Config", t, func() {
		ctrl := gomock.NewController(t)
		Reset(ctrl.Finish)

		Convey("Parse()", func() {
			Convey("Given a single-transport configuration", func() {
				var cfg config.Config
				raw := load("./testdata/single.yml")

				Convey("When I parse the config", func() {
					err := cfg.Parse(raw)

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("The parsed config should include the transport config", func() {
						trcfg, ok := cfg.Transport("test")

						So(ok, ShouldBeTrue)
						So(trcfg, ShouldResemble, config.Transport{
							Use: "trans1",
							Config: map[string]interface{}{
								"key1": 1,
								"key2": map[string]interface{}{
									"key2.1": "val2.1",
									"key2.2": "val2.2",
								},
							},
						})
					})
				})

			})

			Convey("Given a multi-transport configuration", func() {
				raw := load("./testdata/double.yml")

				Convey("When I parse the config", func() {
					var cfg config.Config
					err := cfg.Parse(raw)

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("The parsed config should include all transport configs", func() {
						trcfg, ok := cfg.Transport("test1")
						So(ok, ShouldBeTrue)
						So(trcfg, ShouldResemble, config.Transport{
							Use: "trans1",
							Config: map[string]interface{}{
								"key1": 1,
								"key2": map[string]interface{}{
									"key2.1": "val2.1",
									"key2.2": "val2.2",
								},
							},
						})

						trcfg, ok = cfg.Transport("test2")
						So(ok, ShouldBeTrue)
						So(trcfg, ShouldResemble, config.Transport{
							Use: "trans2",
							Config: map[string]interface{}{
								"key1": map[string]interface{}{
									"key1.1": "val1.1",
									"key1.2": "val1.2",
								},
								"key2": 2,
							},
						})
					})
				})
			})
		})

		Convey("Dog()", func() {
			Convey("Given a parsed single-transport configuration", WithParsedConfig("./testdata/single.yml", func(cfg *config.Config) {
				Convey("When I instantiate *postdog.Dog without providing a config.TransportFactory", func() {
					dog, err := cfg.Dog(context.Background())

					Convey("It should fail", func() {
						So(errors.Is(err, config.ErrUnknownTransport), ShouldBeTrue)
						So(dog, ShouldBeNil)
					})
				})

				Convey("When I instantiate *postdog.Dog and provide a config.TransportFactory", func() {
					factory := mock_config.NewMockTransportFactory(ctrl)
					mockTransport := mock_postdog.NewMockTransport(ctrl)
					factory.EXPECT().
						Transport(gomock.Any(), map[string]interface{}{
							"key1": 1,
							"key2": map[string]interface{}{
								"key2.1": "val2.1",
								"key2.2": "val2.2",
							},
						}).
						Return(mockTransport, nil)

					dog, err := cfg.Dog(context.Background(), config.WithTransportFactory("trans1", factory))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("dog should have the transport configured", func() {
						tr, err := dog.Transport("test")

						So(err, ShouldBeNil)
						So(tr, ShouldEqual, mockTransport)
					})
				})
			}))

			Convey("Given a parsed multi-transport configuration", WithParsedConfig("./testdata/double.yml", func(cfg *config.Config) {
				Convey("When I instantiate *postdog.Dog and provide the config.TransportFactories", func() {
					factory1 := mock_config.NewMockTransportFactory(ctrl)
					mockTransport1 := mock_postdog.NewMockTransport(ctrl)
					factory1.EXPECT().
						Transport(gomock.Any(), map[string]interface{}{
							"key1": 1,
							"key2": map[string]interface{}{
								"key2.1": "val2.1",
								"key2.2": "val2.2",
							},
						}).
						Return(mockTransport1, nil)

					factory2 := mock_config.NewMockTransportFactory(ctrl)
					mockTransport2 := mock_postdog.NewMockTransport(ctrl)
					factory2.EXPECT().
						Transport(gomock.Any(), map[string]interface{}{
							"key1": map[string]interface{}{
								"key1.1": "val1.1",
								"key1.2": "val1.2",
							},
							"key2": 2,
						}).
						Return(mockTransport2, nil)

					dog, err := cfg.Dog(
						context.Background(),
						config.WithTransportFactory("trans1", factory1),
						config.WithTransportFactory("trans2", factory2),
					)

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("dog should have the transports configured", func() {
						tr, err := dog.Transport("test1")
						So(err, ShouldBeNil)
						So(tr, ShouldEqual, mockTransport1)

						tr, err = dog.Transport("test2")
						So(err, ShouldBeNil)
						So(tr, ShouldEqual, mockTransport2)
					})
				})
			}))

			Convey("Given a configuration with a default transport", WithParsedConfig("./testdata/with_default.yml", func(cfg *config.Config) {
				Convey("When I instantiate postdog.Dog and provide the config.TransportFactories", func() {
					factory1 := mock_config.NewMockTransportFactory(ctrl)
					mockTransport1 := mock_postdog.NewMockTransport(ctrl)
					factory1.EXPECT().
						Transport(gomock.Any(), map[string]interface{}{
							"key1": 1,
							"key2": map[string]interface{}{
								"key2.1": "val2.1",
								"key2.2": "val2.2",
							},
						}).
						Return(mockTransport1, nil)

					factory2 := mock_config.NewMockTransportFactory(ctrl)
					mockTransport2 := mock_postdog.NewMockTransport(ctrl)
					factory2.EXPECT().
						Transport(gomock.Any(), map[string]interface{}{
							"key1": map[string]interface{}{
								"key1.1": "val1.1",
								"key1.2": "val1.2",
							},
							"key2": 2,
						}).
						Return(mockTransport2, nil)

					dog, err := cfg.Dog(
						context.Background(),
						config.WithTransportFactory("trans1", factory1),
						config.WithTransportFactory("trans2", factory2),
					)

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("dog should use the specified default transport", func() {
						tr, err := dog.Transport("")
						So(err, ShouldBeNil)
						So(tr, ShouldEqual, mockTransport2)
					})
				})
			}))

			Convey("Given filled environment variables", func() {
				os.Setenv("DEFAULT_TRANSPORT", "test2")
				os.Setenv("TRANSPORT1_USE", "trans1")
				os.Setenv("TRANSPORT2_USE", "trans2")
				os.Setenv("VAL1", "value1")
				os.Setenv("VAL2", "value2")

				Convey("Given a configuration with placeholder variables", WithParsedConfig("./testdata/with_placeholders.yml", func(cfg *config.Config) {
					Convey("When I instantiate postdog.Dog and provide the config.TransportFactories", func() {
						factory1 := mock_config.NewMockTransportFactory(ctrl)
						mockTransport1 := mock_postdog.NewMockTransport(ctrl)
						factory1.EXPECT().
							Transport(gomock.Any(), map[string]interface{}{
								"key1": "value1",
								"key2": "value2",
							}).
							Return(mockTransport1, nil)

						factory2 := mock_config.NewMockTransportFactory(ctrl)
						mockTransport2 := mock_postdog.NewMockTransport(ctrl)
						factory2.EXPECT().
							Transport(gomock.Any(), map[string]interface{}{
								"key3": map[string]interface{}{
									"key3.1": "",
								},
							}).
							Return(mockTransport2, nil)

						dog, err := cfg.Dog(
							context.Background(),
							config.WithTransportFactory("trans1", factory1),
							config.WithTransportFactory("trans2", factory2),
						)

						Convey("It shouldn't fail", func() {
							So(err, ShouldBeNil)
						})

						Convey("dog should use the specified default transport", func() {
							tr, err := dog.Transport("")
							So(err, ShouldBeNil)
							So(tr, ShouldEqual, mockTransport2)
						})
					})
				}))
			})
		})
	})
}

func WithParsedConfig(path string, fn func(*config.Config)) func() {
	return func() {
		var cfg config.Config
		err := cfg.Parse(load(path))
		if err != nil {
			panic(err)
		}
		fn(&cfg)
	}
}

var cacheMux sync.Mutex
var cache = map[string][]byte{}

func load(path string) []byte {
	cacheMux.Lock()
	defer cacheMux.Unlock()
	if b, ok := cache[path]; ok {
		return b
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	cache[path] = b
	return b
}
