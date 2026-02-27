package record

import (
	"go_parser/internal/database"
	"time"
)

type Record struct {
	database.BaseEntity // встраиваем BaseEntity с GetID/SetID

	URL       string                 `json:"url" bson:"url"`
	Domain    string                 `json:"domain" bson:"domain"`
	PlanName  string                 `json:"plan" bson:"plan"`
	Depth     int                    `json:"depth" bson:"depth"`
	Data      map[string]interface{} `json:"data" bson:"data"`
	Links     []string               `json:"links" bson:"links"`
	ParsedAt  time.Time              `json:"parsed_at" bson:"parsed_at"`
	CreatedAt time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" bson:"updated_at"`
}
