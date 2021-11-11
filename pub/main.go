package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vjerci/gig/internal/queue"
	"github.com/vjerci/gig/internal/server"
	"github.com/vjerci/gig/internal/settings"
)

type Queue interface {
	Shutdown()
	Publish(string)
}

var queueConnection Queue

func main() {
	settings.Load()

	ctx := context.Background()
	ctx, cancelWS := context.WithCancel(ctx)

	queueConnection = queue.New(ctx, settings.Rabbit.Queue)

	shutdownChannel := make(chan os.Signal, 1)

	httpEndpoint := server.New(ctx, settings.Http.Address, shutdownChannel, wsHandler)

	signal.Notify(shutdownChannel, os.Interrupt, syscall.SIGTERM)

	<-shutdownChannel
	httpEndpoint.Shutdown()
	cancelWS()
	queueConnection.Shutdown()
}

func wsHandler(ctx context.Context, conn server.Websocket) {
	readCtx, readFailed := context.WithCancel(ctx)

	go func() {
		for {
			_, p, err := conn.ReadMessage()
			if err != nil {
				readFailed()
				return
			}

			log.Printf("read message from websocket: '%s'", string(p))

			queueConnection.Publish(string(p))
		}
	}()

	<-readCtx.Done()
}
