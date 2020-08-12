package mongostore_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/store"
	"github.com/bounoable/postdog/plugin/store/mongostore"
	"github.com/bounoable/postdog/plugin/store/query"
	"github.com/bounoable/postdog/plugin/store/storetest"
	"github.com/google/uuid"
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

	store, err := mongostore.New(client)
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

		store, err := mongostore.New(client)
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

func TestStore_Query_projection(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	ctx := context.Background()
	client, err := connect(ctx)
	if err != nil {
		t.Fatal(fmt.Errorf("connect to client: %w", err))
	}

	noAttachmentContent := bson.D{{Key: "attachments.content", Value: 0}}

	s, err := mongostore.New(client, mongostore.Projection(noAttachmentContent))
	if err != nil {
		t.Fatal(fmt.Errorf("create store: %w", err))
	}

	letters := []store.Letter{
		{
			ID: uuid.New(),
			Letter: letter.Write(
				letter.MustAttach(bytes.NewReader([]byte{1, 2, 3}), "Attachment 1"),
				letter.MustAttach(bytes.NewReader([]byte{2, 3, 4}), "Attachment 2"),
			),
		},
		{
			ID: uuid.New(),
			Letter: letter.Write(
				letter.MustAttach(bytes.NewReader([]byte{3, 4, 5}), "Attachment 1"),
				letter.MustAttach(bytes.NewReader([]byte{4, 5, 6}), "Attachment 2"),
			),
		},
	}

	for _, let := range letters {
		if err = s.Insert(ctx, let); err != nil {
			t.Fatal(fmt.Errorf("insert: %w", err))
		}
	}

	for i, let := range letters {
		for a := range let.Attachments {
			letters[i].Attachments[a].Content = nil
		}
	}

	cur, err := s.Query(ctx, query.New())
	if err != nil {
		t.Fatal(fmt.Errorf("query: %w", err))
	}
	defer cur.Close(ctx)

	var result []store.Letter
	for cur.Next(ctx) {
		result = append(result, cur.Current())
	}

	assert.Equal(t, letters, result)
}

func TestStore_Get(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	storetest.Get(t, func(letters ...store.Letter) query.Repository {
		ctx := context.Background()
		client, err := connect(ctx)
		if err != nil {
			t.Fatal(err)
		}

		store, err := mongostore.New(client)
		if err != nil {
			t.Fatal(t, err)
		}

		for _, let := range letters {
			if err := store.Insert(ctx, let); err != nil {
				t.Fatal(t, err)
			}
		}

		return store
	})
}

func connect(ctx context.Context) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	names, err := client.ListDatabaseNames(ctx, bson.D{
		{Key: "name", Value: bson.D{
			{Key: "$ne", Value: "admin"},
		}},
	})

	if err != nil {
		return nil, fmt.Errorf("list database names: %w", err)
	}

	for _, name := range names {
		if err := client.Database(name).Drop(ctx); err != nil {
			return nil, fmt.Errorf("drop database '%s': %w", name, err)
		}
	}

	return client, nil
}
