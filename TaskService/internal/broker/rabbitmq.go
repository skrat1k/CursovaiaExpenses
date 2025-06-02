package broker

import (
	"ExpensesService/internal/model"
	"encoding/json"
	"log"
	"strconv"

	"github.com/streadway/amqp"
)

const (
	exchangeName = "expenses"
	exchangeType = "direct"
)

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewPublisher(amqpURL string) (*Publisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(exchangeName, exchangeType, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return &Publisher{conn: conn, channel: ch}, nil
}

func (p *Publisher) Close() {
	_ = p.channel.Close()
	_ = p.conn.Close()
}

func (p *Publisher) PublishTask(expense *model.Expense, monthExpense int, routingKey string) error {
	var body []byte
	if expense != nil {
		body, _ = json.Marshal(expense)
	} else {
		monthSample := `{"id":0,"title":"MonthExpense","amount":` + strconv.Itoa(monthExpense) + `}`
		body = []byte(monthSample)
	}

	log.Printf("Публикация задачи в RabbitMQ. RoutingKey: %s", routingKey)
	return p.channel.Publish(
		exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
