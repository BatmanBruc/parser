package task

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct {
	ID         primitive.ObjectID     `bson:"_id,omitempty" json:"-"`
	URL        string                 `bson:"url" json:"url"`
	Plan       string                 `bson:"plan" json:"plan"`
	Depth      int                    `bson:"depth" json:"depth"`
	MaxDepth   int                    `bson:"max_depth" json:"max_depth"`
	ParentURL  string                 `bson:"parent_url,omitempty" json:"parent_url,omitempty"`
	Options    map[string]interface{} `bson:"options" json:"options"`
	Status     string                 `bson:"status" json:"-"`
	CreatedAt  time.Time              `bson:"created_at" json:"-"`
	UpdatedAt  time.Time              `bson:"updated_at" json:"-"`
	RetryCount int                    `bson:"retry_count" json:"-"`
}
