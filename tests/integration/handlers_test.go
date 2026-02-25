package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go_parser/internal/config"
	"go_parser/internal/handlers"
	"go_parser/internal/models"
	"go_parser/internal/services"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestProcessMessage_AlreadyParsed(t *testing.T) {
	// Загружаем конфигурацию
	cfg := config.LoadConfigFromFile("../../.env")

	// Подключение к MongoDB
	var mongoClient *mongo.Client
	mongoClient, err := services.ConnectToMongoDB(cfg.MongoURI)
	if err != nil {
		t.Fatalf("Ошибка подключения к MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	collection := mongoClient.Database(cfg.DatabaseName).Collection(cfg.CollectionName)

	// Подготовка тестовых данных
	testURL := "http://ya.ru/"
	testBody, _ := json.Marshal(map[string]string{"url": testURL})

	// Создаем запись с меткой ParsedAt
	record := models.MongoRecord{
		URL:       testURL,
		Content:   "<html>...</html>",
		CreatedAt: time.Now(),
		ParsedAt:  time.Now(),
	}
	_, err = collection.InsertOne(context.Background(), record)
	if err != nil {
		t.Fatalf("Ошибка вставки тестовой записи: %v", err)
	}

	// Обработка сообщения
	ctx := context.Background()
	err = handlers.ProcessMessage(ctx, testBody, collection)
	if err != nil {
		t.Fatalf("Ошибка обработки сообщения: %v", err)
	}

	// Проверка, что запись не была обновлена
	var updatedRecord models.MongoRecord
	err = collection.FindOne(context.Background(), bson.M{"url": testURL}).Decode(&updatedRecord)
	if err != nil {
		t.Fatalf("Ошибка поиска записи: %v", err)
	}

	// Убедимся, что ParsedAt не изменился
	if !updatedRecord.ParsedAt.Equal(record.ParsedAt) {
		t.Errorf("ParsedAt изменился, хотя не должен был")
	}

	// Удаление тестовой записи
	_, err = collection.DeleteOne(context.Background(), bson.M{"url": testURL})
	if err != nil {
		t.Fatalf("Ошибка удаления записи: %v", err)
	}
}
