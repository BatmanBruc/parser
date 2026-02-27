package models

import "time"

type MongoRecord struct {
	URL       string    `bson:"url"`
	CreatedAt time.Time `bson:"created_at"`
	ParsedAt  time.Time `bson:"parsed_at,omitempty"`
	Content   string    `bson:"content,omitempty"`
}
