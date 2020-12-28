// +build store

package mongo_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bounoable/postdog/letter/rfc"
	"github.com/bounoable/postdog/plugin/archive"
	mongostore "github.com/bounoable/postdog/plugin/archive/mongo"
	"github.com/bounoable/postdog/plugin/archive/test"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"
)

func TestStore(t *testing.T) {
	if testing.Short() {
		t.Skip("[plugin/archive]: Skipping mongodb store test.")
	}

	now := time.Now()
	clock := rfc.ClockFunc(func() time.Time { return now })
	idgen := rfc.IDGeneratorFunc(func(rfc.Mail) string { return "<id@domain>" })

	var counter int32

	test.Store(t, func() archive.Store {
		client, err := connect(context.Background())
		if err != nil {
			panic(err)
		}

		count := atomic.AddInt32(&counter, 1)

		s, err := mongostore.NewStore(
			context.Background(),
			client,
			mongostore.Database(fmt.Sprintf("postdog_%d", count)),
			mongostore.Collection(fmt.Sprintf("mails_%d", count)),
			mongostore.CreateIndexes(false),
			mongostore.RFCOptions(rfc.WithClock(clock), rfc.WithIDGenerator(idgen)),
		)
		if err != nil {
			panic(err)
		}

		return s
	}, test.RoundTime(time.Millisecond))
}

var once sync.Once

func connect(ctx context.Context) (*mongo.Client, error) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		return nil, errors.New("environment variable MONGO_URI must be set")
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	once.Do(func() {
		var names []string
		if names, err = client.ListDatabaseNames(ctx, bson.D{
			{Key: "name", Value: bson.D{
				{Key: "$ne", Value: "admin"},
				{Key: "$ne", Value: "config"},
				{Key: "$ne", Value: "local"},
			}},
		}); err != nil {
			err = fmt.Errorf("list databases: %w", err)
			return
		}

		group, gctx := errgroup.WithContext(ctx)
		for _, name := range names {
			name := name
			group.Go(func() error {
				if err := client.Database(name).Drop(gctx); err != nil {
					return fmt.Errorf("drop database '%s': %w", name, err)
				}
				return nil
			})
		}

		err = group.Wait()
	})

	return client, err
}
