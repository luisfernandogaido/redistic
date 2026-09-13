package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type BulkCause struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

type BulkError struct {
	Type     string    `json:"type"`
	Reason   string    `json:"reason"`
	CausedBy BulkCause `json:"caused_by"`
}

type DL struct {
	Id        bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Timestamp time.Time     `json:"timestamp" bson:"timestamp"`
	Index     string        `json:"index" bson:"index"`
	Payload   any           `json:"payload" bson:"payload"`
	BulkError BulkError     `json:"bulk_error" bson:"bulk_error"`
}

func NewDL(index string, payload any, bulkError BulkError) DL {
	return DL{
		Timestamp: time.Now(),
		Index:     index,
		Payload:   payload,
		BulkError: bulkError,
	}
}
