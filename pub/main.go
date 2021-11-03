package main

import (
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

	queueConnection = queue.New(settings.Rabbit.Queue)
	shutdownChannel := make(chan os.Signal, 1)

	httpEndpoint := server.New(settings.Http.Address, shutdownChannel, wsHandler)

	signal.Notify(shutdownChannel, os.Interrupt, syscall.SIGTERM)

	<-shutdownChannel
	httpEndpoint.Shutdown()
	queueConnection.Shutdown()
}

func wsHandler(conn server.Websocket) {
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			return
		}

		log.Printf("read message from websocket: '%s'", string(p))

		queueConnection.Publish(string(p))
	}
}
