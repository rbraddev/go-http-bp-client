package client

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestClientFunctions(t *testing.T) {
	t.Run("test default client can send request", func(t *testing.T) {
		myClient, _ := NewClient()

		req, _ := http.NewRequest(http.MethodGet, "https://www.google.com", nil)
		resp, err := myClient.httpClient.Do(req)

		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, 200)
	})

	t.Run("test client can initiate with proxy", func(t *testing.T) {
		proxyStrings := []string{"127.0.0.1:8080"}
		myClient, err := NewClient(WithProxy(proxyStrings))

		assert.NoError(t, err)
		assert.Equal(t, myClient.proxies[0], Proxy{"127.0.0.1", "8080", "", ""})
	})

	t.Run("test client can send request with proxy", func(t *testing.T) {
		proxyStrings := []string{"127.0.0.1:8080"}
		myClient, _ := NewClient(WithProxy(proxyStrings))

		req, _ := http.NewRequest(http.MethodGet, "https://www.google.com", nil)
		resp, err := myClient.httpClient.Do(req)

		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, 200)
	})
}
