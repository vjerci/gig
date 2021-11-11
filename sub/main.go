package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/vjerci/gig/internal/queue"
	"github.com/vjerci/gig/internal/server"
	"github.com/vjerci/gig/internal/settings"
)

type Queue interface {
	Read()
	Shutdown()
	Subscribe() queue.Subscription
}

var queueConnection Queue

func main() {
	settings.Load()

	ctx := context.Background()
	ctx, cancelWS := context.WithCancel(ctx)

	queueConnection = queue.New(ctx, settings.Rabbit.Queue)
	go queueConnection.Read()

	shutdownChannel := make(chan os.Signal, 1)

	httpEndpoint := server.New(ctx, settings.Http.Address, shutdownChannel, wsHandler)

	signal.Notify(shutdownChannel, os.Interrupt, syscall.SIGTERM)

	<-shutdownChannel
	httpEndpoint.Shutdown()
	cancelWS()
	queueConnection.Shutdown()
}

func wsHandler(ctx context.Context, conn server.Websocket) {
	subscriber := queueConnection.Subscribe()

	readCtx, readFailed := context.WithCancel(ctx)
	go func() {
		_, _, err := conn.ReadMessage()
		if err != nil && !errors.Is(err, websocket.ErrCloseSent) {
			subscriber.Unsubscribe()
			readFailed()
		}
	}()

	for {
		select {
		case <-readCtx.Done():
			return
		case msg := <-subscriber.Channel():
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg))

			if err != nil {
				log.Printf("error writing message to socket: %e", err)
				return
			}

			log.Printf("published message to websocket: '%s'", msg)
		}
	}
}
