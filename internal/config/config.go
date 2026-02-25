package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

// Config представляет структуру конфигурации.
type Config struct {
	MongoURI       string
	RabbitMQURI    string
	DatabaseName   string
	CollectionName string
	QueueName      string
}

// getProjectRoot возвращает абсолютный путь к корневой директории проекта.
func getProjectRoot() string {
	// Получаем путь к текущему файлу (config.go)
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Ошибка получения пути к текущему файлу")
	}
	// Переходим на три уровня вверх (config -> go_parser)
	projectRoot := filepath.Dir(filepath.Dir(filename))
	return projectRoot
}

// LoadConfig загружает конфигурацию из файла .env в корне проекта.
func LoadConfig() *Config {
	return LoadConfigFromFile(filepath.Join(getProjectRoot(), "./../.env"))
}

// LoadConfigFromFile загружает конфигурацию из указанного файла .env.
func LoadConfigFromFile(filename string) *Config {
	err := godotenv.Load(filename)
	if err != nil {
		log.Fatalf("Ошибка загрузки файла .env: %v", err)
	}

	return &Config{
		MongoURI:       os.Getenv("MONGO_URI"),
		RabbitMQURI:    os.Getenv("RABBITMQ_URL"),
		DatabaseName:   os.Getenv("DATABASE_NAME"),
		CollectionName: os.Getenv("COLLECTION_NAME"),
		QueueName:      os.Getenv("QUEUE_NAME"),
	}
}
