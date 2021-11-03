package main

import (
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

	queueConnection = queue.New(settings.Rabbit.Queue)
	go queueConnection.Read()

	shutdownChannel := make(chan os.Signal, 1)

	httpEndpoint := server.New(settings.Http.Address, shutdownChannel, wsHandler)

	signal.Notify(shutdownChannel, os.Interrupt, syscall.SIGTERM)

	<-shutdownChannel
	httpEndpoint.Shutdown()
	queueConnection.Shutdown()
}

func wsHandler(conn server.Websocket) {
	subscriber := queueConnection.Subscribe()

	socketClose := make(chan bool, 1)

	go func() {
		_, _, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			subscriber.Unsubscribe()
			socketClose <- true
		}
	}()

	for {
		select {
		case <-socketClose:
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
