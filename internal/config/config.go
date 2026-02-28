package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI       string
	RabbitMQURI    string
	DatabaseName   string
	CollectionName string
	QueueName      string
}

func LoadConfig() *Config {
	return &Config{
		MongoURI:       GetEnv("MONGO_URI", "mongodb://mongo:27017/"),
		RabbitMQURI:    GetEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		DatabaseName:   GetEnv("DATABASE_NAME", "parser"),
		CollectionName: GetEnv("COLLECTION_NAME", "tasks"),
		QueueName:      GetEnv("QUEUE_NAME", "parser_queue"),
	}
}

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

func GetEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func GetEnvAsInt(key string, defaultVal int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func GetEnvAsBool(key string, defaultVal bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultVal
}
