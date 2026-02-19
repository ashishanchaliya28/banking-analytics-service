package model

import (
	"time"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Event struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     bson.ObjectID `bson:"user_id" json:"user_id"`
	EventName  string        `bson:"event_name" json:"event_name"`
	Properties interface{}   `bson:"properties,omitempty" json:"properties,omitempty"`
	CreatedAt  time.Time     `bson:"created_at" json:"created_at"`
}

type Segment struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    bson.ObjectID `bson:"user_id" json:"user_id"`
	Segments  []string      `bson:"segments" json:"segments"` // e.g., "high_value", "fd_holder", "upi_active"
	UpdatedAt time.Time     `bson:"updated_at" json:"updated_at"`
}

type CrossSellRule struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Segment     string        `bson:"segment" json:"segment"`
	ProductType string        `bson:"product_type" json:"product_type"`
	Title       string        `bson:"title" json:"title"`
	Description string        `bson:"description" json:"description"`
	IsActive    bool          `bson:"is_active" json:"is_active"`
}

type RecordEventRequest struct {
	EventName  string      `json:"event_name"`
	Properties interface{} `json:"properties,omitempty"`
}

type CrossSellOffer struct {
	ProductType string `json:"product_type"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
