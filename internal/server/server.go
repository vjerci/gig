package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Websocket interface {
	Close() error
	WriteMessage(int, []byte) error
	ReadMessage() (int, []byte, error)
}

type Server struct {
	http        *http.Server
	notifyClose chan<- os.Signal
}

func New(port string, notifyClose chan os.Signal, wsHandler func(ws Websocket)) *Server {
	server := &Server{
		notifyClose: notifyClose,
	}

	router := mux.NewRouter()

	router.HandleFunc("/ws", socketHandlerFactory(wsHandler))
	router.HandleFunc("/shutdown", shutdownFactory(server)).Methods("DELETE")
	router.HandleFunc("/healthcheck", healthcheck).Methods("GET")

	server.http = &http.Server{
		Addr:    port,
		Handler: router,
	}

	log.Printf("Server listening on address %s", port)
	go func() {
		err := server.http.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Printf("error starting up server %e", err)
		}
	}()

	return server
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Write([]byte("healthy"))
}

func (server *Server) Shutdown() {
	log.Println("server shutting down")

	context, contextCancel := context.WithTimeout(context.Background(), 5*time.Second)
	err := server.http.Shutdown(context)
	if err != nil {
		log.Printf("failed to close server got %e", err)
	}
	contextCancel()
}

func shutdownFactory(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("got shutdown request")

		w.Write([]byte("shutting down"))
		r.Body.Close()

		server.notifyClose <- os.Kill
	}
}

func socketHandlerFactory(wsHandler func(ws Websocket)) func(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("fail upgrade connection %e", err)
		}

		wsHandler(ws)
	}
}
