package mongostore_test

import (
	"context"
	"os"
	"testing"

	"github.com/bounoable/postdog/plugin/store"
	"github.com/bounoable/postdog/plugin/store/mongostore"
	"github.com/bounoable/postdog/plugin/store/query"
	"github.com/bounoable/postdog/plugin/store/storetest"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestNew(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	cases := map[string]struct {
		opts   []mongostore.Option
		assert func(*testing.T, *mongostore.Store)
	}{
		"default options": {
			assert: func(t *testing.T, store *mongostore.Store) {
				assert.Equal(t, mongostore.Config{
					DatabaseName:   "postdog",
					CollectionName: "letters",
					CreateIndexes:  true,
				}, store.Config())
			},
		},
		"custom database name": {
			opts: []mongostore.Option{
				mongostore.Database("mailing"),
			},
			assert: func(t *testing.T, store *mongostore.Store) {
				assert.Equal(t, "mailing", store.Config().DatabaseName)
			},
		},
		"custom collection name": {
			opts: []mongostore.Option{
				mongostore.Collection("mails"),
			},
			assert: func(t *testing.T, store *mongostore.Store) {
				assert.Equal(t, "mails", store.Config().CollectionName)
			},
		},
		"disable index creation": {
			opts: []mongostore.Option{
				mongostore.CreateIndexes(false),
			},
			assert: func(t *testing.T, store *mongostore.Store) {
				assert.Equal(t, false, store.Config().CreateIndexes)
			},
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			client, err := connect(ctx)
			if err != nil {
				t.Fatal(err)
			}

			mstore, err := mongostore.New(client, tcase.opts...)
			assert.Nil(t, err)
			tcase.assert(t, mstore)
		})
	}
}

func TestStore_Insert(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	ctx := context.Background()
	client, err := connect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	store, err := mongostore.New(client, mongostore.Database("mailing"), mongostore.Collection("mails"))
	if err != nil {
		t.Fatal(err)
	}

	storetest.Insert(t, store)
}

func TestStore_Query(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	storetest.Query(t, func(letters ...store.Letter) query.Repository {
		ctx := context.Background()
		client, err := connect(ctx)
		if err != nil {
			t.Fatal(err)
		}

		store, err := mongostore.New(client, mongostore.Database("mailing"), mongostore.Collection("mails"))
		if err != nil {
			t.Fatal(t, err)
		}

		for _, let := range letters {
			if err := store.Insert(ctx, let); err != nil {
				t.Fatal(err)
			}
		}

		return store
	})
}

func connect(ctx context.Context) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		return nil, err
	}

	names, err := client.ListDatabaseNames(ctx, bson.D{
		{Key: "name", Value: bson.D{
			{Key: "$ne", Value: "admin"},
		}},
	})

	if err != nil {
		return nil, err
	}

	for _, name := range names {
		if err := client.Database(name).Drop(ctx); err != nil {
			return nil, err
		}
	}

	return client, nil
}
