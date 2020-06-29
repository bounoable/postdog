// Package mongostore provides the mongodb store implementation.
package mongostore

import (
	"context"
	"errors"
	"time"

	"github.com/bounoable/mongoutil/index"
	"github.com/bounoable/postdog/plugin/store"
	"github.com/bounoable/postdog/plugin/store/query"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// Provider is the store provider name.
	Provider = "mongo"

	defaultConfig = Config{
		DatabaseName:   "postdog",
		CollectionName: "letters",
		CreateIndexes:  true,
	}
)

func init() {
	store.RegisterProvider(
		Provider,
		store.FactoryFunc(func(ctx context.Context, cfg map[string]interface{}) (store.Store, error) {
			clientOpts := options.Client()
			if uri, _ := cfg["uri"].(string); uri != "" {
				clientOpts.ApplyURI(uri)
			}

			client, err := mongo.Connect(ctx, clientOpts)
			if err != nil {
				return nil, err
			}

			if err = client.Ping(ctx, nil); err != nil {
				return nil, err
			}

			var opts []Option

			if db, ok := cfg["database"].(string); ok {
				opts = append(opts, Database(db))
			}

			if col, ok := cfg["collection"].(string); ok {
				opts = append(opts, Collection(col))
			}

			if ci, ok := cfg["createIndexes"].(bool); ok {
				opts = append(opts, CreateIndexes(ci))
			}

			return New(client, opts...)
		}),
	)
}

// Store is the mongodb store.
type Store struct {
	cfg Config
	col *mongo.Collection
}

// Config is the store configuration.
type Config struct {
	DatabaseName   string
	CollectionName string
	CreateIndexes  bool
}

// New creates a mongodb store.
// By default, the store uses the "postdog" database and "letters" collection.
// You can configure the store with opts.
func New(client *mongo.Client, opts ...Option) (*Store, error) {
	if client == nil {
		return nil, errors.New("nil client")
	}

	cfg := defaultConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	store := &Store{
		cfg: cfg,
		col: client.Database(cfg.DatabaseName).Collection(cfg.CollectionName),
	}

	if cfg.CreateIndexes {
		if err := store.createIndexes(); err != nil {
			return nil, err
		}
	}

	return store, nil
}

// Option is a store option.
type Option func(*Config)

// Database configures the used database.
// Defaults to "postdog".
func Database(name string) Option {
	return func(cfg *Config) {
		cfg.DatabaseName = name
	}
}

// Collection configures the used collection.
// Defaults to "letters".
func Collection(name string) Option {
	return func(cfg *Config) {
		cfg.CollectionName = name
	}
}

// CreateIndexes configures the store to create the indexes for the fields of the letters.
// By default, this option is set to true.
func CreateIndexes(create bool) Option {
	return func(cfg *Config) {
		cfg.CreateIndexes = create
	}
}

// Config returns the store configuration.
func (s *Store) Config() Config {
	return s.cfg
}

// Insert inserts let into s.
func (s *Store) Insert(ctx context.Context, let store.Letter) error {
	_, err := s.col.InsertOne(ctx, mapLetter(let))
	return err
}

// Query queries the store for letters, filtered and sorted by q.
func (s *Store) Query(ctx context.Context, q query.Query) (query.Cursor, error) {
	filter := buildFilter(q)
	opts := options.Find().SetSort(buildSort(q.Sort))

	if q.Paginate.Page != 0 && q.Paginate.PerPage != 0 {
		opts = opts.
			SetSkip(int64((q.Paginate.Page - 1) * q.Paginate.PerPage)).
			SetLimit(int64(q.Paginate.PerPage))
	}

	cur, err := s.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	return &cursor{cur: cur}, nil
}

func buildFilter(q query.Query) bson.D {
	filter := bson.D{}

	if len(q.Subjects) > 0 {
		filter = appendValsFilter(filter, []string{"subject"}, buildTextSubstringFilterValue(q.Subjects))
	}

	if len(q.From) > 0 {
		filter = appendValsFilter(filter, []string{"from.name", "from.address"}, buildTextSubstringFilterValue(q.From))
	}

	if len(q.To) > 0 {
		filter = appendValsFilter(filter, []string{"to.name", "to.address"}, buildTextSubstringFilterValue(q.To))
	}

	if len(q.CC) > 0 {
		filter = appendValsFilter(filter, []string{"cc.name", "cc.address"}, buildTextSubstringFilterValue(q.CC))
	}

	if len(q.BCC) > 0 {
		filter = appendValsFilter(filter, []string{"bcc.name", "bcc.address"}, buildTextSubstringFilterValue(q.BCC))
	}

	if len(q.Attachment.Names) > 0 {
		filter = appendValsFilter(filter, []string{"attachments.filename"}, buildTextSubstringFilterValue(q.Attachment.Names))
	}

	if len(q.Attachment.ContentTypes) > 0 {
		filter = appendValsFilter(filter, []string{"attachments.contentType"}, buildTextSubstringFilterValue(q.Attachment.ContentTypes))
	}

	if len(q.Attachment.Size.Exact) > 0 {
		filter = append(filter, bson.E{Key: "attachments.size", Value: bson.D{{Key: "$in", Value: q.Attachment.Size.Exact}}})
	}

	if len(q.Attachment.Size.Ranges) > 0 {
		or := make([]bson.D, len(q.Attachment.Size.Ranges))
		for i, rang := range q.Attachment.Size.Ranges {
			or[i] = bson.D{{Key: "attachments.size", Value: bson.D{
				{Key: "$gte", Value: rang[0]},
				{Key: "$lte", Value: rang[1]},
			}}}
		}
		filter = append(filter, bson.E{Key: "$or", Value: or})
	}

	if before := q.SentAt.Before; !before.IsZero() {
		// mongo only has millisecond precision
		if before.Nanosecond() > 0 {
			before = before.
				Add(-time.Duration(before.Nanosecond())).
				Add(time.Millisecond)
		}

		filter = append(filter, bson.E{
			Key:   "sentAt",
			Value: bson.D{{Key: "$lt", Value: before}},
		})
	}

	if after := q.SentAt.After; !after.IsZero() {
		// mongo only has millisecond precision
		if after.Nanosecond() > 0 {
			after = after.
				Add(-time.Duration(after.Nanosecond())).
				Add(-time.Millisecond)
		}

		filter = append(filter, bson.E{
			Key:   "sentAt",
			Value: bson.D{{Key: "$gt", Value: after}},
		})
	}

	return filter
}

func appendValsFilter(filter bson.D, keys []string, vals interface{}) bson.D {
	if len(keys) == 1 {
		return append(filter, bson.E{Key: keys[0], Value: vals})
	}

	or := make([]bson.D, len(keys))
	for i, key := range keys {
		or[i] = bson.D{{Key: key, Value: vals}}
	}

	return append(filter, bson.E{Key: "$or", Value: or})
}

func buildTextSubstringFilterValue(texts []string) bson.D {
	exprs := make([]primitive.Regex, len(texts))
	for i, text := range texts {
		exprs[i] = primitive.Regex{
			Pattern: text,
			Options: "i",
		}
	}

	return bson.D{{
		Key:   "$in",
		Value: exprs,
	}}
}

func buildSort(cfg query.SortConfig) (sort bson.D) {
	defer func() {
		if cfg.Dir == query.SortDesc {
			sort[0].Value = -1
			return
		}
		sort[0].Value = 1
	}()

	sort = bson.D{{}}

	switch cfg.SortBy {
	default:
		sort[0].Key = "sentAt"
	}
	return
}

func (s *Store) createIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	_, err := index.CreateFromConfig(ctx, s.col.Database(), index.Config{
		s.cfg.CollectionName: []mongo.IndexModel{
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "sentAt", Value: 1}}},
			{Keys: bson.D{{Key: "size", Value: 1}}},
			{Keys: bson.D{{Key: "subject", Value: 1}}},
			{Keys: bson.D{{Key: "from.name", Value: 1}}},
			{Keys: bson.D{{Key: "from.address", Value: 1}}},
			{Keys: bson.D{{Key: "to.name", Value: 1}}},
			{Keys: bson.D{{Key: "to.address", Value: 1}}},
			{Keys: bson.D{{Key: "cc.name", Value: 1}}},
			{Keys: bson.D{{Key: "cc.address", Value: 1}}},
			{Keys: bson.D{{Key: "bcc.name", Value: 1}}},
			{Keys: bson.D{{Key: "bcc.address", Value: 1}}},
			{Keys: bson.D{{Key: "attachments.filename", Value: 1}}},
			{Keys: bson.D{{Key: "attachments.contentType", Value: 1}}},
		},
	})

	return err
}

type cursor struct {
	cur     *mongo.Cursor
	err     error
	current store.Letter
}

func (cur *cursor) Next(ctx context.Context) bool {
	if !cur.cur.Next(ctx) {
		cur.err = cur.cur.Err()
		return false
	}

	var dblet dbLetter
	if cur.err = cur.cur.Decode(&dblet); cur.err != nil {
		cur.current = store.Letter{}
		return false
	}
	cur.current = dblet.store()

	return true
}

func (cur *cursor) Current() store.Letter {
	return cur.current
}

func (cur *cursor) Close(ctx context.Context) error {
	cur.err = cur.cur.Close(ctx)
	return cur.err
}

func (cur *cursor) Err() error {
	return cur.err
}

// Get retrieves the letter with the given id
func (s *Store) Get(ctx context.Context, id uuid.UUID) (store.Letter, error) {
	res := s.col.FindOne(ctx, bson.M{"id": id})

	var dblet dbLetter
	if err := res.Decode(&dblet); err != nil {
		return store.Letter{}, err
	}

	return dblet.store(), nil
}
