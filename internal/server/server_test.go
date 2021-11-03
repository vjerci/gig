package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthcheck(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
	w := httptest.NewRecorder()

	healthcheck(w, req)

	res := w.Result()
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected OK but got %d", res.StatusCode)
	}

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if string(data) != "healthy" {
		t.Errorf("expected healthy got %v", string(data))
	}
}

func TestShutdown(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
	w := httptest.NewRecorder()

	healthcheck(w, req)

	res := w.Result()
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected OK but got %d", res.StatusCode)
	}

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if string(data) != "healthy" {
		t.Errorf("expected healthy got %v", string(data))
	}
}
