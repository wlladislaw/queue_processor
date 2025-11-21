package processor

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	connection *amqp.Connection
	chanel     *amqp.Channel
}

func NewRabbitMQ(url string) (Processor, error) {
	rabbitConn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	rabbitChan, err := rabbitConn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{
		connection: rabbitConn,
		chanel:     rabbitChan,
	}, nil
}

func (r *RabbitMQ) Publish(qName string, data []byte) error {
	_, err := r.chanel.QueueDeclare(
		qName, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	errPub := r.chanel.Publish(
		"",    // exchange
		qName, // routing key
		false, // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         data,
			Timestamp:    time.Now(),
		})
	if errPub != nil {
		return errPub
	}
	return nil
}

func (r *RabbitMQ) Consume(qName string) (Tuber, error) {
	q, err := r.chanel.QueueDeclare(
		qName, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	err = r.chanel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return nil, err
	}

	tasks, errCons := r.chanel.Consume(
		qName, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if errCons != nil {
		return nil, errCons
	}
	fmt.Printf("queue %s have %d msg and %d consumers\n",
		q.Name, q.Messages, q.Consumers)

	return &RabbitTube{
		Tasks:  r.convertRabbitChTasks(tasks),
		Chanel: r.chanel,
		Qname:  qName,
	}, nil

}

func (r *RabbitMQ) Close() error {
	closeErr := r.chanel.Close()
	return closeErr
}

type RabbitTube struct {
	Tasks  <-chan Task
	Chanel *amqp.Channel
	Qname  string
}

func (rt *RabbitTube) GetTasks() <-chan Task {
	return rt.Tasks
}
func (rt *RabbitTube) GetTubeName() string {
	return rt.Qname
}
func (rt *RabbitTube) GetMesgsCount() int {
	q, err := rt.Chanel.QueueInspect(rt.Qname)
	if err != nil {
		fmt.Printf("err receiving data length tube: %v", err)
		return -1
	}
	return q.Messages
}

type RabbitMsg struct {
	amqpDelivery amqp.Delivery
}

func (rt *RabbitMsg) Body() []byte {
	return rt.amqpDelivery.Body
}
func (rt *RabbitMsg) GetCreatedAt() time.Time {
	return rt.amqpDelivery.Timestamp
}
func (rt *RabbitMsg) Ack(multiple bool) error {
	return rt.amqpDelivery.Ack(multiple)
}

func (rt *RabbitMsg) Nack(multile bool, requeue bool) error {
	return rt.amqpDelivery.Nack(multile, requeue)
}

func (r *RabbitMQ) convertRabbitChTasks(msgs <-chan amqp.Delivery) <-chan Task {
	out := make(chan Task)

	go func() {
		defer close(out)
		for t := range msgs {
			out <- &RabbitMsg{amqpDelivery: t}
		}
	}()

	return out
}
