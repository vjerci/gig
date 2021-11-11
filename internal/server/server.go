package server

import (
	"context"
	"errors"
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

func New(ctx context.Context, port string, notifyClose chan os.Signal, wsHandler func(ctx context.Context, ws Websocket)) *Server {
	server := &Server{
		notifyClose: notifyClose,
	}

	router := mux.NewRouter()

	router.HandleFunc("/ws", socketHandlerFactory(ctx, wsHandler))
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
	context, contextCancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer contextCancel()

	err := server.http.Shutdown(context)

	if err != nil {
		log.Printf("failed to close server got %e", err)
	}
}

func shutdownFactory(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("got shutdown request")

		w.Write([]byte("shutting down"))
		r.Body.Close()

		server.notifyClose <- os.Kill
	}
}

func socketHandlerFactory(context context.Context, wsHandler func(ctx context.Context, ws Websocket)) func(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		defer r.Body.Close()

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("fail upgrade connection %e", err)
			return
		}
		defer ws.Close()

		wsHandler(context, ws)

		err = ws.WriteMessage(websocket.CloseMessage, []byte{})
		if err != nil && !errors.Is(err, websocket.ErrCloseSent) {
			log.Printf("error writing close message %s", err)
		}
	}
}
