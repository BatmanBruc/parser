package plan

import (
	"go_parser/internal/domain/task"
	"time"
)

type Plan interface {
	Name() string
	Domen() string
	Match(url string) bool
	Execute(task *task.Task) (*PlanResult, []FoundURL, error)
}

type PlanResult struct {
	ID         string                 `json:"id" bson:"_id,omitempty"`
	TaskID     string                 `json:"task_id" bson:"task_id"`
	URL        string                 `json:"url" bson:"url"`
	PlanName   string                 `json:"plan" bson:"plan"`
	Depth      int                    `json:"depth" bson:"depth"`
	MaxDepth   int                    `json:"maxdepth" bson:"maxdepth"`
	Title      string                 `json:"title" bson:"title"`
	Content    string                 `json:"content,omitempty" bson:"content,omitempty"`
	HTML       string                 `json:"html,omitempty" bson:"html,omitempty"`
	Data       map[string]interface{} `json:"data" bson:"data"`
	FoundURLs  []FoundURL             `json:"found_urls,omitempty" bson:"found_urls,omitempty"`
	StatusCode int                    `json:"status_code" bson:"status_code"`
	ParsedAt   time.Time              `json:"parsed_at" bson:"parsed_at"`
	Duration   int64                  `json:"duration_ms" bson:"duration_ms"`
	Error      string                 `json:"error,omitempty" bson:"error,omitempty"`
}

type FoundURL struct {
	URL      string                 `json:"url" bson:"url"`
	Plan     string                 `json:"plan" bson:"plan"`
	Priority int                    `json:"priority" bson:"priority"`
	Type     string                 `json:"type" bson:"type"`
	Context  map[string]interface{} `json:"context,omitempty" bson:"context,omitempty"`
	FoundAt  time.Time              `json:"found_at" bson:"found_at"`
}
