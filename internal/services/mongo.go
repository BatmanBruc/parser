package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go_parser/internal/utils"
)

const mongoService = "MongoDB"

func ConnectToMongoDB(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, utils.NewError(mongoService, err)
	}

	// Проверка подключения
	if err := client.Ping(ctx, nil); err != nil {
		return nil, utils.NewError(mongoService, err)
	}

	return client, nil
}
