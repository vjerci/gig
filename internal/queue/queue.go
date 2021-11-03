package queue

import (
	"log"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"github.com/vjerci/gig/internal/settings"
)

type Channel interface {
	Close() error
	Consume(string, string, bool, bool, bool, bool, amqp.Table) (<-chan amqp.Delivery, error)
	Publish(string, string, bool, bool, amqp.Publishing) error
}

type Connection interface {
	Close() error
}

type Store struct {
	conn    Connection
	channel Channel
	name    string

	close       chan bool
	unsubscribe chan *Subscriber

	subscribers map[string]*Subscriber
}

func New(queueName string) *Store {
	conn, err := amqp.Dial(settings.Rabbit.Address)
	if err != nil {
		log.Fatalf("fail to establish Queue to rabbit %e", err)
	}

	ch, err := conn.Channel()

	if err != nil {
		log.Fatalf("fail to open a channel %e", err)
	}

	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	if err != nil {
		log.Fatalf("fail to declare queue %e", err)
	}

	Queue := &Store{
		conn:        conn,
		channel:     ch,
		name:        q.Name,
		close:       make(chan bool),
		unsubscribe: make(chan *Subscriber),
		subscribers: make(map[string]*Subscriber),
	}

	return Queue
}

func (store *Store) Shutdown() {
	store.close <- true
	store.channel.Close()
	store.conn.Close()
}

func (store *Store) Publish(msg string) {
	log.Printf("publishing message to queue: '%s'", msg)

	err := store.channel.Publish(
		"",         // exchange
		store.name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		},
	)

	if err != nil {
		log.Fatalf("fail to publish to queue %e", err)
	}
}

func (store *Store) Read() {
	msgs, err := store.channel.Consume(
		store.name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)

	if err != nil {
		log.Fatalf("fail to consume queue %e", err)
	}

	for {
		select {
		case <-store.close:
			return
		case subscriber := <-store.unsubscribe:
			delete(store.subscribers, subscriber.id)
		case msg := <-msgs:
			log.Printf("read message from queue '%s'", string(msg.Body))

			for _, subscriber := range store.subscribers {
				subscriber.receiving <- string(msg.Body)
			}
		}
	}
}

type Subscription interface {
	Channel() <-chan string
	Unsubscribe()
}

type Subscriber struct {
	receiving chan string
	id        string
	store     *Store
}

func (store *Store) Subscribe() Subscription {
	sub := &Subscriber{
		id:        uuid.NewString(),
		receiving: make(chan string, 1),
		store:     store,
	}
	store.subscribers[sub.id] = sub

	return sub
}

func (sub *Subscriber) Unsubscribe() {
	sub.store.unsubscribe <- sub
}

func (sub *Subscriber) Channel() <-chan string {
	return sub.receiving
}
