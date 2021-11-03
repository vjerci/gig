package e2e_test

import (
	"os"
	"testing"

	"github.com/gorilla/websocket"
)

func TestMulitpleConsumers(t *testing.T) {
	pubEndpoint := os.Getenv("PUB_ENDPOINT")
	subEndpoint := os.Getenv("SUB_ENDPOINT")

	pubConn, _, errPub := websocket.DefaultDialer.Dial(pubEndpoint, nil)
	subConn1, _, errSub1 := websocket.DefaultDialer.Dial(subEndpoint, nil)
	subConn2, _, errSub2 := websocket.DefaultDialer.Dial(subEndpoint, nil)

	if errPub != nil {
		t.Errorf("want open websocket to publisher got error %e", errPub)
	}

	if errSub1 != nil {
		t.Errorf("want open websocket to subscriber got error %e", errSub1)
	}

	if errSub2 != nil {
		t.Errorf("want open websocket to subscriber got error %e", errSub2)
	}

	msg := "Hello world"

	// wait for connections to be established

	err := pubConn.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		t.Errorf("want succesful write to pub websocket got err %e", err)
	}

	_, readMsg, err := subConn1.ReadMessage()
	if err != nil {
		t.Errorf("want successful read of sub websocket, got err %e", err)
	}

	if string(readMsg) != msg {
		t.Errorf("Read msg %s not coressponding to published message %s", msg, readMsg)
	}

	_, readMsg, err = subConn2.ReadMessage()
	if err != nil {
		t.Errorf("want successful read of sub websocket, got err %e", err)
	}

	if string(readMsg) != msg {
		t.Errorf("Read msg %s not coressponding to published message %s", msg, readMsg)
	}

	errPubClose := pubConn.Close()
	if errPubClose != nil {
		t.Errorf("failed closing connecting to publisher %e", errPubClose)
	}

	errSub1Close := subConn1.Close()

	if errSub1Close != nil {
		t.Errorf("failed closing connection to subscriber 1 %e", errPubClose)
	}

	errSub2Close := subConn2.Close()
	if errSub1Close != nil {
		t.Errorf("failed closing connection to subscriber 1 %e", errSub2Close)
	}
}
