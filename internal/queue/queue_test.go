package queue

import (
	"context"
	"testing"

	"github.com/streadway/amqp"
)

type MockConnection struct {
	Closed bool
}

func (m *MockConnection) Close() error {
	m.Closed = true
	return nil
}

type MockChannel struct {
	Closed      bool
	Key         string
	Message     amqp.Publishing
	MessageChan chan amqp.Delivery
}

func (m *MockChannel) Close() error {
	m.Closed = true
	return nil
}

func (m *MockChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	m.Message = msg
	m.Key = key

	return nil
}

func (m *MockChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	m.Key = queue

	return m.MessageChan, nil
}

func TestShutdown(t *testing.T) {
	mockConnection := &MockConnection{}
	mockChannel := &MockChannel{}
	store := &Store{
		conn:    mockConnection,
		channel: mockChannel,
		close:   make(chan bool, 1),
	}

	store.Shutdown()

	if mockConnection.Closed != true {
		t.Errorf("expected closed connection")
	}

	if mockChannel.Closed != true {
		t.Errorf("expected closed channel ")
	}
}

func TestPublish(t *testing.T) {
	mockConnection := &MockConnection{}
	mockChannel := &MockChannel{}

	storeName := "storeName"
	msg := "hello"

	store := &Store{
		conn:    mockConnection,
		channel: mockChannel,
		name:    storeName,
	}

	store.Publish(msg)

	if string(mockChannel.Message.Body) != msg {
		t.Errorf("want message to match %s %s", string(mockChannel.Message.Body), msg)
	}

	if mockChannel.Key != storeName {
		t.Errorf("want publish to be called with correct name %s", store.name)
	}
}

func TestRead(t *testing.T) {
	mockConnection := &MockConnection{}
	mockChannel := &MockChannel{
		MessageChan: make(chan amqp.Delivery),
	}

	msg := "hello"

	ctx, closeStore := context.WithCancel(context.Background())
	defer closeStore()

	store := &Store{
		conn:        mockConnection,
		channel:     mockChannel,
		subscribers: make(map[string]*Subscriber),
		close:       make(chan bool, 1),
		ctx:         ctx,
	}

	sub1 := store.Subscribe()
	sub2 := store.Subscribe()

	bothRead := make(chan bool, 2)

	go store.Read()

	go func() {
		readMsg1 := <-sub1.Channel()

		if readMsg1 != msg {
			t.Errorf("want same message %s %s", msg, readMsg1)
		}
		bothRead <- true
	}()

	go func() {
		readMsg2 := <-sub2.Channel()

		if readMsg2 != msg {
			t.Errorf("want same message %s %s", msg, readMsg2)
		}

		bothRead <- true
	}()

	go func() {
		mockChannel.MessageChan <- amqp.Delivery{
			Body: []byte(msg),
		}
	}()

	<-bothRead
	<-bothRead

}

func TestUnsubscribe(t *testing.T) {
	mockConnection := &MockConnection{}
	mockChannel := &MockChannel{
		MessageChan: make(chan amqp.Delivery),
	}

	ctx, closeStore := context.WithCancel(context.Background())
	defer closeStore()

	store := &Store{
		conn:        mockConnection,
		channel:     mockChannel,
		subscribers: make(map[string]*Subscriber),
		unsubscribe: make(chan *Subscriber),
		close:       make(chan bool, 1),
		ctx:         ctx,
	}

	sub1 := store.Subscribe()
	go store.Read()

	if len(store.subscribers) != 1 {
		t.Errorf("expected subscriber to be 1 got %d", len(store.subscribers))
	}

	sub1.Unsubscribe()

	closeStore()

	subNumber := len(store.subscribers)
	if subNumber != 0 {
		t.Errorf("expected subscriber to be 0 got %d", subNumber)
	}
}
