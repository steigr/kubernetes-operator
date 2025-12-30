package configuration

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWaitForJenkinsHealth(t *testing.T) {
	t.Run("success on first attempt", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/health/", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := &Configuration{}
		err := c.waitForJenkinsHealth(server.URL)

		assert.NoError(t, err)
	})

	t.Run("success on second attempt", func(t *testing.T) {
		var attempts int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempt := atomic.AddInt32(&attempts, 1)
			if attempt == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := &Configuration{}
		err := c.waitForJenkinsHealth(server.URL)

		assert.NoError(t, err)
		assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
	})

	t.Run("success on third attempt", func(t *testing.T) {
		var attempts int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempt := atomic.AddInt32(&attempts, 1)
			if attempt < 3 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := &Configuration{}
		err := c.waitForJenkinsHealth(server.URL)

		assert.NoError(t, err)
		assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
	})

	t.Run("failure after all retries with non-200 status", func(t *testing.T) {
		var attempts int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&attempts, 1)
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		c := &Configuration{}
		err := c.waitForJenkinsHealth(server.URL)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Jenkins health check failed after 3 retries")
		assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
	})

	t.Run("failure with connection error", func(t *testing.T) {
		c := &Configuration{}
		// Use an invalid URL that will fail to connect
		err := c.waitForJenkinsHealth("http://localhost:99999")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Jenkins health check failed after 3 retries")
	})

	t.Run("trims trailing slash from URL", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/health/", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := &Configuration{}
		err := c.waitForJenkinsHealth(server.URL + "/")

		assert.NoError(t, err)
	})

	t.Run("handles URL with existing path", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/jenkins/health/", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := &Configuration{}
		err := c.waitForJenkinsHealth(server.URL + "/jenkins")

		assert.NoError(t, err)
	})
}
