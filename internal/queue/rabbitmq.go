package queue

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.Mutex
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// اطمینان از وجود صف پیش‌فرض
	_, err = ch.QueueDeclare(
		"telegram_messages", // نام صف پیش‌فرض
		true,                // Durable
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		// اگر QueueDeclare خطا بده، کانکشن رو ببندیم و خطا برگردونیم
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	return &RabbitMQ{conn: conn, channel: ch}, nil
}

// Publish اکنون واریادیک است و هم signature قدیمی (payload) و هم حالت (exchange, routingKey, payload) را پشتیبانی می‌کند.
func (r *RabbitMQ) Publish(args ...interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var exchange string
	var routingKey string
	var payload interface{}

	switch len(args) {
	case 1:
		// Publish(payload)
		payload = args[0]
		exchange = ""                    // exchange خالی -> default exchange
		routingKey = "telegram_messages" // routing key پیش‌فرض
	case 3:
		// Publish(exchange, routingKey, payload)
		var ok bool
		exchange, ok = args[0].(string)
		if !ok {
			return fmt.Errorf("publish: first arg (exchange) must be string")
		}
		routingKey, ok = args[1].(string)
		if !ok {
			return fmt.Errorf("publish: second arg (routingKey) must be string")
		}
		payload = args[2]
	default:
		return fmt.Errorf("publish: invalid number of arguments (%d). use Publish(payload) or Publish(exchange, routingKey, payload)", len(args))
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return r.channel.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

func (r *RabbitMQ) Close() {
	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
}
