// Package store persists flags and the audit log in MongoDB.
package store

import (
	"context"
	"errors"
	"time"

	"github.com/mralaminahamed/flagcast/packages/shared/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const opTimeout = 5 * time.Second

// ErrNotFound is returned when a flag key does not exist.
var ErrNotFound = errors.New("flag not found")

// ErrConflict is returned when creating a flag whose key already exists.
var ErrConflict = errors.New("flag already exists")

type FlagStore struct {
	client *mongo.Client
	flags  *mongo.Collection
	audit  *mongo.Collection
}

func NewFlagStore(ctx context.Context, uri, db string) (*FlagStore, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	d := client.Database(db)
	audit := d.Collection("audit")
	_, _ = audit.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "flag_key", Value: 1}, {Key: "timestamp", Value: -1}},
	})
	return &FlagStore{client: client, flags: d.Collection("flags"), audit: audit}, nil
}

func (s *FlagStore) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	return s.client.Ping(ctx, nil)
}

func (s *FlagStore) Close(ctx context.Context) error { return s.client.Disconnect(ctx) }

// Create inserts a new flag. Returns ErrConflict if the key already exists.
func (s *FlagStore) Create(ctx context.Context, f models.Flag) (models.Flag, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	now := time.Now().UTC()
	f.CreatedAt, f.UpdatedAt = now, now
	if _, err := s.flags.InsertOne(ctx, f); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return models.Flag{}, ErrConflict
		}
		return models.Flag{}, err
	}
	return f, nil
}

// Get returns one flag or ErrNotFound.
func (s *FlagStore) Get(ctx context.Context, key string) (models.Flag, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	var f models.Flag
	if err := s.flags.FindOne(ctx, bson.M{"_id": key}).Decode(&f); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.Flag{}, ErrNotFound
		}
		return models.Flag{}, err
	}
	return f, nil
}

// List returns all flags sorted by key.
func (s *FlagStore) List(ctx context.Context) ([]models.Flag, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	cur, err := s.flags.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []models.Flag{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Update replaces the mutable fields of an existing flag. Returns ErrNotFound.
func (s *FlagStore) Update(ctx context.Context, f models.Flag) (models.Flag, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	f.UpdatedAt = time.Now().UTC()
	res := s.flags.FindOneAndUpdate(ctx,
		bson.M{"_id": f.Key},
		bson.M{"$set": bson.M{
			"name":        f.Name,
			"description": f.Description,
			"enabled":     f.Enabled,
			"rollout":     f.Rollout,
			"tags":        f.Tags,
			"updated_at":  f.UpdatedAt,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var out models.Flag
	if err := res.Decode(&out); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.Flag{}, ErrNotFound
		}
		return models.Flag{}, err
	}
	return out, nil
}

// Delete removes a flag. Returns ErrNotFound if it did not exist.
func (s *FlagStore) Delete(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	res, err := s.flags.DeleteOne(ctx, bson.M{"_id": key})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// AppendAudit records a mutation. Best-effort — callers log but don't fail on it.
func (s *FlagStore) AppendAudit(ctx context.Context, e models.AuditEntry) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	e.Timestamp = time.Now().UTC()
	_, err := s.audit.InsertOne(ctx, e)
	return err
}

// ListAudit returns recent audit entries, newest first, optionally by flag.
func (s *FlagStore) ListAudit(ctx context.Context, flagKey string, limit int64) ([]models.AuditEntry, error) {
	if limit <= 0 {
		limit = 100
	} else if limit > 1000 {
		limit = 1000
	}
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	filter := bson.M{}
	if flagKey != "" {
		filter["flag_key"] = flagKey
	}
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(limit)
	cur, err := s.audit.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []models.AuditEntry{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
