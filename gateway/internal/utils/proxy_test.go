package utils_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/maksroxx/DeliveryService/gateway/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestProxyRequest_Success(t *testing.T) {
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "bar", r.Header.Get("X-Test"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer targetServer.Close()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Test", "bar")
	rr := httptest.NewRecorder()

	err := utils.ProxyRequest(rr, req, targetServer.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"status":"ok"}`, rr.Body.String())
}

func TestProxyRequest_BadTarget(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	err := utils.ProxyRequest(rr, req, "http://badhost.invalid")
	assert.Error(t, err)
}

func TestProxyRequest_FailureFromTarget(t *testing.T) {
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "failed", http.StatusBadRequest)
	}))
	defer targetServer.Close()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	err := utils.ProxyRequest(rr, req, targetServer.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "failed")
}

func TestProxyRequest_POST(t *testing.T) {
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		assert.Equal(t, "test-body", string(body))
		w.WriteHeader(http.StatusCreated)
	}))
	defer targetServer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test-body"))
	rr := httptest.NewRecorder()

	err := utils.ProxyRequest(rr, req, targetServer.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rr.Code)
}
