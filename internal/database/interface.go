package database

import (
	"context"
)

type Entity interface {
	GetID() string
	SetID(string)
}

type BaseEntity struct {
	ID string `bson:"_id,omitempty" json:"id"`
}

func (e *BaseEntity) GetID() string   { return e.ID }
func (e *BaseEntity) SetID(id string) { e.ID = id }

type Filter map[string]interface{}

type Options struct {
	Limit  int64
	Offset int64
	Sort   map[string]int // поле: 1 (ASC), -1 (DESC)
}

type Repository[T Entity] interface {
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
	Ping(ctx context.Context) error

	Create(ctx context.Context, entity T) error
	Get(ctx context.Context, id string) (T, error)
	Update(ctx context.Context, entity T) error
	Delete(ctx context.Context, id string) error

	Find(ctx context.Context, filter Filter, opts *Options) ([]T, error)
	FindOne(ctx context.Context, filter Filter) (T, error)
	Count(ctx context.Context, filter Filter) (int64, error)
}
