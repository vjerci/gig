package main

import (
	"context"
	"errors"
	"testing"
)

type MockQueue struct {
	publishedMsg string
}

func (q *MockQueue) Shutdown() {}

func (q *MockQueue) Publish(msg string) {
	q.publishedMsg = msg
}

type MockWebsocket struct {
	counter      int
	msg          string
	publishedMsg string
}

func (ws *MockWebsocket) Close() error {
	return nil
}

func (ws *MockWebsocket) ReadMessage() (int, []byte, error) {
	if ws.counter == 0 {
		return 0, []byte{}, errors.New("errored")
	}
	ws.counter--
	return 0, []byte(ws.msg), nil
}

func (ws *MockWebsocket) WriteMessage(messageType int, msg []byte) error {
	return nil
}

func TestWsHandler(t *testing.T) {
	msg := "hello world"

	channel := &MockQueue{}
	ws := &MockWebsocket{
		counter: 1,
		msg:     msg,
	}

	queueConnection = channel
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		wsHandler(ctx, ws)

		if channel.publishedMsg != msg {
			t.Errorf("want publish msg to be same that was pushed trough websocket %s %s", channel.publishedMsg, msg)
		}
	}()

	cancel()

}
