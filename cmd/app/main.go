package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go_parser/internal/config"
	"go_parser/internal/handlers"
	"go_parser/internal/services"
	"go_parser/internal/utils"
)

func main() {
	// Загрузка конфигурации
	cfg := config.LoadConfig()
	utils.Logger.Println("Конфигурация загружена.")

	// Подключение к MongoDB
	utils.Logger.Println("Подключение к MongoDB...")
	mongoClient, err := services.ConnectToMongoDB(cfg.MongoURI)
	if err != nil {
		utils.Logger.Fatalf("Ошибка подключения к MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			utils.Logger.Printf("Ошибка отключения от MongoDB: %v", err)
		} else {
			utils.Logger.Println("Успешно отключено от MongoDB.")
		}
	}()
	utils.Logger.Println("Успешно подключено к MongoDB.")

	// Подключение к RabbitMQ
	utils.Logger.Println("Подключение к RabbitMQ...")
	rabbitMQConn, err := services.ConnectToRabbitMQ(cfg.RabbitMQURI)
	if err != nil {
		utils.Logger.Fatalf("Ошибка подключения к RabbitMQ: %v", err)
	}
	defer func() {
		if err := rabbitMQConn.Close(); err != nil {
			utils.Logger.Printf("Ошибка закрытия соединения с RabbitMQ: %v", err)
		} else {
			utils.Logger.Println("Успешно закрыто соединение с RabbitMQ.")
		}
	}()
	utils.Logger.Println("Успешно подключено к RabbitMQ.")

	// Создание канала RabbitMQ
	utils.Logger.Println("Создание канала RabbitMQ...")
	ch, err := rabbitMQConn.Channel()
	if err != nil {
		utils.Logger.Fatalf("Ошибка создания канала RabbitMQ: %v", err)
	}
	defer func() {
		if err := ch.Close(); err != nil {
			utils.Logger.Printf("Ошибка закрытия канала RabbitMQ: %v", err)
		} else {
			utils.Logger.Println("Успешно закрыт канал RabbitMQ.")
		}
	}()
	utils.Logger.Println("Канал RabbitMQ успешно создан.")

	// Объявление очереди
	utils.Logger.Println("Объявление очереди...")
	q, err := ch.QueueDeclare(
		cfg.QueueName, // Имя очереди
		false,         // durable (не сохранять на диск)
		false,         // autoDelete (не удалять при отсутствии потребителей)
		false,         // exclusive (очередь доступна для других соединений)
		false,         // noWait (ждать ответа от сервера)
		nil,           // arguments (дополнительные аргументы)
	)
	if err != nil {
		utils.Logger.Fatalf("Ошибка объявления очереди: %v", err)
	}
	utils.Logger.Printf("Очередь '%s' успешно объявлена.\n", q.Name)

	// Подписка на очередь
	utils.Logger.Println("Подписка на очередь...")
	msgs, err := ch.Consume(
		q.Name, // Имя очереди
		"",     // consumer tag (пустое значение для автоматической генерации)
		false,  // autoAck (не подтверждать сообщения автоматически)
		false,  // exclusive (очередь доступна для других потребителей)
		false,  // noLocal (доставлять сообщения, отправленные тем же соединением)
		false,  // noWait (ждать ответа от сервера)
		nil,    // arguments (дополнительные аргументы)
	)
	if err != nil {
		utils.Logger.Fatalf("Ошибка подписки на очередь: %v", err)
	}
	utils.Logger.Println("Успешно подписались на очередь.")

	// Graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	utils.Logger.Println("Ожидание сообщений. Для выхода нажмите CTRL+C.")

	// Обработка сообщений
	go func() {
		for msg := range msgs {
			utils.Logger.Printf("Получено новое сообщение: %s\n", msg.Body)

			// Обработка сообщения
			ctx := context.Background()
			err := handlers.ProcessMessage(ctx, msg.Body, mongoClient.Database(cfg.DatabaseName).Collection(cfg.CollectionName))
			if err != nil {
				utils.Logger.Printf("Ошибка обработки сообщения: %v\n", err)
				continue
			}

			// Подтверждение сообщения
			err = msg.Ack(false) // false = подтвердить только это сообщение
			if err != nil {
				utils.Logger.Printf("Ошибка подтверждения сообщения: %v\n", err)
			} else {
				utils.Logger.Println("Сообщение успешно обработано и подтверждено.")
			}
		}
	}()

	// Ожидание сигнала завершения
	<-sigs
	utils.Logger.Println("Завершение работы...")
}
