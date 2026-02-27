package queue

import (
	"go_parser/internal/utils"
	"time"

	"github.com/rabbitmq/amqp091-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

const rabbitService = "RabbitMQ"

func ConnectToRabbitMQ(uri string) (*amqp.Connection, error) {
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

type Message struct {
	msg amqp091.Delivery
}

func NewMessage(msg amqp091.Delivery) *Message {
	return &Message{
		msg: msg,
	}
}

func (m *Message) Success() {
	m.msg.Ack(false)
}

func (m *Message) TryAgain() {
	m.msg.Nack(false, true)
}

func (m *Message) Reject() {
	m.msg.Reject(false)
}

func (m *Message) GetBody() []byte {
	return m.msg.Body
}
