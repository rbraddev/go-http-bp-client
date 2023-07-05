package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseProxies(t *testing.T) {
	t.Run("tests correct []string to []Proxy", func(t *testing.T) {
		rawProxies := []string{"127.0.0.1:8080", "127.0.0.1:8080:admin:password"}
		parsedProxies, err := parseProxies(rawProxies)

		assert.Equal(t, Proxy{"127.0.0.1", "8080", "", ""}, parsedProxies[0])
		assert.Equal(t, Proxy{"127.0.0.1", "8080", "admin", "password"}, parsedProxies[1])
		assert.NoError(t, err)
	})

	t.Run("tests incorrect proxy string", func(t *testing.T) {
		rawProxies := []string{"127.0.0.18080"}
		_, err := parseProxies(rawProxies)

		assert.Errorf(t, err, "invalid proxy string: 127.0.0.18080")
	})
}
