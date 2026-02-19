package service

import (
	"context"
	"errors"
	"github.com/banking-superapp/analytics-service/model"
	"github.com/banking-superapp/analytics-service/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrUnauthorized = errors.New("unauthorized")

type AnalyticsService interface {
	RecordEvent(ctx context.Context, userID string, req *model.RecordEventRequest) (*model.Event, error)
	GetSegment(ctx context.Context, userID string) (*model.Segment, error)
	GetCrossSellOffers(ctx context.Context, userID string) ([]model.CrossSellOffer, error)
}

type analyticsService struct {
	eventRepo     repository.EventRepo
	segmentRepo   repository.SegmentRepo
	crossSellRepo repository.CrossSellRuleRepo
}

func NewAnalyticsService(er repository.EventRepo, sr repository.SegmentRepo, cr repository.CrossSellRuleRepo) AnalyticsService {
	return &analyticsService{er, sr, cr}
}

func (s *analyticsService) RecordEvent(ctx context.Context, userID string, req *model.RecordEventRequest) (*model.Event, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	event := &model.Event{
		UserID:     oid,
		EventName:  req.EventName,
		Properties: req.Properties,
	}

	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, err
	}

	// Trigger async segment update based on event (simplified inline logic)
	go s.updateSegmentOnEvent(context.Background(), oid, req.EventName)

	return event, nil
}

func (s *analyticsService) GetSegment(ctx context.Context, userID string) (*model.Segment, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	segment, err := s.segmentRepo.FindByUserID(ctx, oid)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Return default segment
			return &model.Segment{
				UserID:   oid,
				Segments: []string{"new_user"},
			}, nil
		}
		return nil, err
	}
	return segment, nil
}

func (s *analyticsService) GetCrossSellOffers(ctx context.Context, userID string) ([]model.CrossSellOffer, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	segment, err := s.segmentRepo.FindByUserID(ctx, oid)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}

	var userSegments []string
	if segment != nil {
		userSegments = segment.Segments
	} else {
		userSegments = []string{"new_user"}
	}

	var offers []model.CrossSellOffer
	seen := map[string]bool{}

	for _, seg := range userSegments {
		rules, err := s.crossSellRepo.FindBySegment(ctx, seg)
		if err != nil {
			continue
		}
		for _, rule := range rules {
			key := rule.ProductType + ":" + rule.Title
			if !seen[key] {
				seen[key] = true
				offers = append(offers, model.CrossSellOffer{
					ProductType: rule.ProductType,
					Title:       rule.Title,
					Description: rule.Description,
				})
			}
		}
	}

	// If no segment-specific offers, return default offers from all active rules
	if len(offers) == 0 {
		rules, err := s.crossSellRepo.FindAll(ctx)
		if err == nil {
			for _, rule := range rules {
				key := rule.ProductType + ":" + rule.Title
				if !seen[key] {
					seen[key] = true
					offers = append(offers, model.CrossSellOffer{
						ProductType: rule.ProductType,
						Title:       rule.Title,
						Description: rule.Description,
					})
				}
			}
		}
	}

	return offers, nil
}

// updateSegmentOnEvent updates user segments based on events (simplified rule engine)
func (s *analyticsService) updateSegmentOnEvent(ctx context.Context, userID bson.ObjectID, eventName string) {
	existing, err := s.segmentRepo.FindByUserID(ctx, userID)

	var segments []string
	if err == nil && existing != nil {
		segments = existing.Segments
	}

	// Simple rule: add segments based on event names
	newSegment := ""
	switch eventName {
	case "fd_created":
		newSegment = "fd_holder"
	case "upi_payment":
		newSegment = "upi_active"
	case "high_value_transaction":
		newSegment = "high_value"
	case "loan_applied":
		newSegment = "loan_seeker"
	case "investment_viewed":
		newSegment = "investment_interested"
	}

	if newSegment != "" && !containsSegment(segments, newSegment) {
		segments = append(segments, newSegment)
		_ = s.segmentRepo.Upsert(ctx, &model.Segment{
			UserID:   userID,
			Segments: segments,
		})
	}
}

func containsSegment(segments []string, s string) bool {
	for _, seg := range segments {
		if seg == s {
			return true
		}
	}
	return false
}
