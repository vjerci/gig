package main

import (
	"errors"
	"testing"

	"github.com/vjerci/gig/internal/queue"
)

type MockQueue struct {
	subscriber *MockSubscription
}

func (q *MockQueue) Shutdown() {}

func (q *MockQueue) Read() {}

func (q *MockQueue) Subscribe() queue.Subscription {
	return q.subscriber
}

type MockWebsocket struct {
	writtenMsg string
	close      chan bool
	written    chan bool
}

func (ws *MockWebsocket) Close() error {
	return nil
}

func (ws *MockWebsocket) ReadMessage() (int, []byte, error) {
	<-ws.close
	return 0, []byte{}, errors.New("got error")
}

func (ws *MockWebsocket) WriteMessage(messageType int, msg []byte) (err error) {
	ws.writtenMsg = string(msg)
	ws.written <- true
	return nil
}

type MockSubscription struct {
	sendChannel chan string
}

func (m *MockSubscription) Channel() <-chan string {
	return m.sendChannel
}

func (m *MockSubscription) Unsubscribe() {}

func TestWsHandler(t *testing.T) {
	msg := "hello world"

	sendChannel := make(chan string)
	close := make(chan bool)
	finished := make(chan bool)
	written := make(chan bool)

	channel := &MockQueue{
		subscriber: &MockSubscription{
			sendChannel: sendChannel,
		},
	}
	ws := &MockWebsocket{
		close:   close,
		written: written,
	}

	queueConnection = channel

	go func() {
		wsHandler(ws)
		finished <- true
	}()

	sendChannel <- msg
	<-written

	if ws.writtenMsg != msg {
		t.Errorf("want publish msg to be same that was pushed trough websocket %s %s", ws.writtenMsg, msg)
	}

	close <- true
	<-finished
}
