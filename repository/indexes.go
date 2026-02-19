package repository

import (
	"context"
	"time"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreateIndexes(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := db.Collection("events").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "event_name", Value: 1}}},
		// TTL: auto-delete events older than 365 days
		{Keys: bson.D{{Key: "created_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(86400 * 365)},
	})
	if err != nil {
		return err
	}

	_, err = db.Collection("segments").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}, Options: options.Index().SetUnique(true)},
	})
	if err != nil {
		return err
	}

	_, err = db.Collection("crosssell_rules").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "segment", Value: 1}, {Key: "is_active", Value: 1}}},
	})
	return err
}
