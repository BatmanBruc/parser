package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go_parser/internal/config"
	"go_parser/internal/database"
	"go_parser/internal/domain/record"
	"go_parser/internal/handler"
	"go_parser/internal/parser/plans"
	"go_parser/internal/queue"
	"go_parser/internal/utils"
	"go_parser/internal/worker"

	"github.com/playwright-community/playwright-go"
)

func main() {
	cfg := config.LoadConfig()
	utils.Logger.Println("Конфигурация загружена.")

	if err := playwright.Install(); err != nil {
		log.Fatal("Ошибка установки playwright:", err)
	}

	ctx := context.Background()

	recordRepo := database.NewMongoRepository[*record.Record](
		cfg.MongoURI,
		"parser_db",
		"records",
	)

	err := recordRepo.Connect(ctx)
	if err != nil {
		utils.Logger.Fatalf("Ошибка подключения к MongoDB: %v %s", err, cfg.MongoURI)
	}

	defer recordRepo.Close(ctx)

	utils.Logger.Println("Подключение к RabbitMQ...")
	rabbitMQConn, err := queue.ConnectToRabbitMQ(cfg.RabbitMQURI)
	if err != nil {
		utils.Logger.Fatalf("Ошибка подключения к RabbitMQ: %v %s", err, cfg.RabbitMQURI)
	}
	defer func() {
		if err := rabbitMQConn.Close(); err != nil {
			utils.Logger.Printf("Ошибка закрытия соединения с RabbitMQ: %v", err)
		} else {
			utils.Logger.Println("Успешно закрыто соединение с RabbitMQ.")
		}
	}()
	utils.Logger.Println("Успешно подключено к RabbitMQ.")

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

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	utils.Logger.Println("Ожидание сообщений. Для выхода нажмите CTRL+C.")

	h := handler.NewHandler(recordRepo, ch, cfg.QueueName)

	pr := plans.NewRegistr()
	pr.Register(plans.NewHackerNewsPlan())

	wp := worker.NewWorkerPool(3, pr, h)

	wp.Start()

	go func() {
		for msg := range msgs {
			utils.Logger.Printf("Получено новое сообщение: %s\n", msg.Body)
			wp.Msg <- queue.NewMessage(msg)
		}
	}()
	<-sigs
	ch.Close()
	wp.Stop()
}
