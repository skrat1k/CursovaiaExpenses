package main

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

type Expense struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Amount int    `json:"amount"`
}

const (
	rabbitURL    = "amqp://guest:guest@localhost:5672/"
	exchangeName = "expenses"
)

var queueBindings = []struct {
	QueueName  string
	RoutingKey string
}{
	{"add_notification_queue", "expense.Add"},
	{"get_notification_queue", "expense.Get"},
	{"getmonth_notification_queue", "expense.GetMounth"},
}

func main() {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatal("Cannot connect to rabbit", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Cannot open channel", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Error create exchange", err)
	}

	for _, binding := range queueBindings {

		q, err := ch.QueueDeclare(
			binding.QueueName,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Fatalf("Error create queue %s: %v", binding.QueueName, err)
		}

		err = ch.QueueBind(
			q.Name,
			binding.RoutingKey,
			exchangeName,
			false,
			nil,
		)
		if err != nil {
			log.Fatalf("Error binding queue %s: %v", binding.QueueName, err)
		}

		msg, err := ch.Consume(
			q.Name,
			"",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Fatalf("Error consuming from queue %s: %v", binding.QueueName, err)
		}

		log.Println("NotificationService listening queue", q.Name)

		go func(queueName string, messages <-chan amqp.Delivery) {
			for d := range messages {
				var expense Expense
				if err := json.Unmarshal(d.Body, &expense); err != nil {
					log.Printf("[%s] JSON unmarshal error: %v", queueName, err)
					continue
				}
				if expense.ID != 0 {
					log.Printf("[%s] Expense: ID=%d, Title=%s, Amount=%d", queueName, expense.ID, expense.Title, expense.Amount)
				} else {
					log.Printf("[%s] MonthExpense: %d", queueName, expense.Amount)
				}
			}
		}(binding.QueueName, msg)
	}
	forever := make(chan bool)
	<-forever
}
