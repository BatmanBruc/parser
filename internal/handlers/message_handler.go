package handlers

import (
	"context"
	"go_parser/internal/utils"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

const handlerTimeout = 30 * time.Second

func ProcessMessageWithTimeout(body []byte, collection *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), handlerTimeout)
	defer cancel()

	// Передаем контекст в ProcessMessage
	if err := ProcessMessage(ctx, body, collection); err != nil {
		return utils.NewError("MessageHandler", err)
	}

	return nil
}
