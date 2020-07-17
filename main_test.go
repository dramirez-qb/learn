package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func xestHelloRoute(t *testing.T) {
	handler := http.HandlerFunc(helloHandler)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "Hello")
}

func xestPingRoute(t *testing.T) {
	handler := http.HandlerFunc(pongHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func xestHealthRoute(t *testing.T) {
	handler := http.HandlerFunc(healthCheckHandler)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/healthz", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "alive")
}
