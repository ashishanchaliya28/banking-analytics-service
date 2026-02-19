package repository

import (
	"context"
	"time"
	"github.com/banking-superapp/analytics-service/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type EventRepo interface {
	Create(ctx context.Context, e *model.Event) error
	FindByUserID(ctx context.Context, userID bson.ObjectID, limit int64) ([]model.Event, error)
}

type SegmentRepo interface {
	Upsert(ctx context.Context, s *model.Segment) error
	FindByUserID(ctx context.Context, userID bson.ObjectID) (*model.Segment, error)
}

type CrossSellRuleRepo interface {
	FindBySegment(ctx context.Context, segment string) ([]model.CrossSellRule, error)
	FindAll(ctx context.Context) ([]model.CrossSellRule, error)
}

type eventRepo struct{ col *mongo.Collection }
type segmentRepo struct{ col *mongo.Collection }
type crossSellRuleRepo struct{ col *mongo.Collection }

func NewEventRepo(db *mongo.Database) EventRepo { return &eventRepo{col: db.Collection("events")} }
func NewSegmentRepo(db *mongo.Database) SegmentRepo {
	return &segmentRepo{col: db.Collection("segments")}
}
func NewCrossSellRuleRepo(db *mongo.Database) CrossSellRuleRepo {
	return &crossSellRuleRepo{col: db.Collection("crosssell_rules")}
}

func (r *eventRepo) Create(ctx context.Context, e *model.Event) error {
	e.CreatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, e)
	return err
}

func (r *eventRepo) FindByUserID(ctx context.Context, userID bson.ObjectID, limit int64) ([]model.Event, error) {
	if limit <= 0 {
		limit = 50
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(limit)
	cursor, err := r.col.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var events []model.Event
	cursor.All(ctx, &events)
	return events, nil
}

func (r *segmentRepo) Upsert(ctx context.Context, s *model.Segment) error {
	s.UpdatedAt = time.Now()
	_, err := r.col.UpdateOne(ctx,
		bson.M{"user_id": s.UserID},
		bson.M{"$set": s},
		&mongo.UpdateOptions{Upsert: boolPtr(true)},
	)
	return err
}

func (r *segmentRepo) FindByUserID(ctx context.Context, userID bson.ObjectID) (*model.Segment, error) {
	var s model.Segment
	err := r.col.FindOne(ctx, bson.M{"user_id": userID}).Decode(&s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *crossSellRuleRepo) FindBySegment(ctx context.Context, segment string) ([]model.CrossSellRule, error) {
	cursor, err := r.col.Find(ctx, bson.M{"segment": segment, "is_active": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var rules []model.CrossSellRule
	cursor.All(ctx, &rules)
	return rules, nil
}

func (r *crossSellRuleRepo) FindAll(ctx context.Context) ([]model.CrossSellRule, error) {
	cursor, err := r.col.Find(ctx, bson.M{"is_active": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var rules []model.CrossSellRule
	cursor.All(ctx, &rules)
	return rules, nil
}

func boolPtr(b bool) *bool { return &b }
