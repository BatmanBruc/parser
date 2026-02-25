package services

import (
	"go_parser/internal/utils"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const rabbitService = "RabbitMQ"

func ConnectToRabbitMQ(uri string) (*amqp.Connection, error) {
	// Добавляем несколько попыток подключения
	var conn *amqp.Connection
	var err error

	for i := 0; i < 3; i++ {
		conn, err = amqp.Dial(uri)
		if err == nil {
			return conn, nil
		}
		time.Sleep(time.Second * 2)
	}

	return nil, utils.NewError(rabbitService, err)
}

func CreateChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, utils.NewError(rabbitService, err)
	}

	return ch, nil
}
