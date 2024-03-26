package proxy

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"

	"github.com/robbyt/llm_proxy/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// randomly finds an available port to bind to
func getFreePort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return "", err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", err
	}
	defer l.Close()
	port := l.Addr().(*net.TCPAddr).Port
	return fmt.Sprintf("localhost:%d", port), nil
}

func webServer(hitCounter *atomic.Int32, listenAddr string) (*http.Server, func()) {
	if hitCounter == nil {
		panic("hitCounter must be non-nil")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// increment the counter
		hitCounter.Add(1)
		resp := fmt.Sprintf("hits: %d\n", hitCounter.Load())
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(resp))
	})

	srv := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return srv, func() {
		if err := srv.Close(); err != nil {
			log.Printf("HTTP server Close: %v", err)
		}
	}
}

func httpClient(proxyAddr string) (*http.Client, error) {
	proxyURL, err := url.Parse(proxyAddr)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	return &http.Client{
		Transport: transport,
		Timeout:   10 * http.DefaultClient.Timeout,
	}, nil
}

func TestProxySimple(t *testing.T) {
	proxyPort, err := getFreePort()
	require.NoError(t, err)

	testServerPort, err := getFreePort()
	require.NoError(t, err)

	// Create a simple proxy config
	cfg := config.NewDefaultConfig()
	cfg.CertDir = t.TempDir()
	cfg.NoHttpUpgrader = true // disable TLS because our test server doesn't support it
	cfg.Listen = proxyPort
	cfg.AppMode = config.SimpleMode

	// create a proxy with the test config
	p, err := configProxy(cfg)
	require.NoError(t, err)
	require.NotNil(t, p)

	// external control of the proxy
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// start the proxy in the background
	go func() {
		err = startProxy(p, shutdown)
		require.NoError(t, err)
	}()

	// Start a basic web server on another port
	hitCounter := new(atomic.Int32)
	srv, srvShutdown := webServer(hitCounter, testServerPort)
	require.NotNil(t, srv)
	require.NotNil(t, srvShutdown)

	// Create a client that will use the proxy
	client, err := httpClient("http://" + proxyPort)
	require.NoError(t, err)

	t.Run("TestSimpleProxy", func(t *testing.T) {
		hitCounter.Store(0) // reset the counter
		// make a request using that client, through the proxy
		resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		// check the response body from req1
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, []byte("hits: 1\n"), body)
		assert.Equal(t, int32(1), hitCounter.Load())
	})

	t.Run("TestSimpleProxy2", func(t *testing.T) {
		hitCounter.Store(5) // reset the counter
		// make another request using that client, through the proxy
		resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		// check the response body from req2
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, []byte("hits: 6\n"), body)
		assert.Equal(t, int32(6), hitCounter.Load())
	})

	// done with tests, send shutdown signals
	t.Cleanup(func() {
		srvShutdown()            // close the simple web server
		shutdown <- os.Interrupt // close the proxy
	})
}

func TestProxyCache(t *testing.T) {
	proxyPort, err := getFreePort()
	require.NoError(t, err)

	testServerPort, err := getFreePort()
	require.NoError(t, err)

	tmpDir := t.TempDir()
	// Create a simple proxy config
	cfg := config.NewDefaultConfig()
	cfg.CertDir = tmpDir + "/certs"
	cfg.Cache.Dir = tmpDir + "/cache"
	cfg.NoHttpUpgrader = true // disable TLS because our test server doesn't support it
	cfg.Listen = proxyPort
	cfg.AppMode = config.CacheMode

	// create a proxy with the test config
	p, err := configProxy(cfg)
	require.NoError(t, err)
	require.NotNil(t, p)

	// external control of the proxy
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// start the proxy in the background
	go func() {
		err = startProxy(p, shutdown)
		require.NoError(t, err)
	}()

	// Start a basic web server on another port
	hitCounter := new(atomic.Int32)
	srv, srvShutdown := webServer(hitCounter, testServerPort)
	require.NotNil(t, srv)
	require.NotNil(t, srvShutdown)

	// Create a client that will use the proxy
	client, err := httpClient("http://" + cfg.Listen)
	require.NoError(t, err)

	t.Run("TestCacheMiss", func(t *testing.T) {
		hitCounter.Store(0) // reset the counter
		// make a request using the client, through the proxy
		resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// check the response body from this request
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, []byte("hits: 1\n"), body)
		assert.Equal(t, int32(1), hitCounter.Load())
		assert.Equal(t, "MISS", resp.Header.Get("X-Cache"))
	})

	t.Run("TestCacheHit", func(t *testing.T) {
		hitCounter.Store(5) // reset the counter
		// make another request using the client, through the proxy
		resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		// check the response body from this request
		// (should be the cached response with value=1, not the incremented value)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, []byte("hits: 1\n"), body)
		assert.Equal(t, int32(5), hitCounter.Load()) // the counter should not be 6, because we got a cache hit
		assert.Equal(t, "HIT", resp.Header.Get("X-Cache"))
	})

	// done with tests, send shutdown signals
	t.Cleanup(func() {
		srvShutdown()            // close the simple web server
		shutdown <- os.Interrupt // close the proxy
	})
}

// Testing imperative code is tough
func TestNewProxy(t *testing.T) {
	tempDir := t.TempDir()

	ca, err := newCA(tempDir)
	assert.NoError(t, err)

	p, err := newProxy(1, "localhost:8080", false, ca)
	assert.NoError(t, err)
	assert.NotNil(t, p)
}

func TestNewCA(t *testing.T) {
	tempDir := t.TempDir()

	ca, err := newCA(tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, ca)
}

func TestConfigProxy(t *testing.T) {
	// Create a mock configuration
	cfg := config.NewDefaultConfig()
	cfg.CertDir = t.TempDir()
	cfg.AppMode = config.SimpleMode

	// Call the function with the mock configuration
	p, err := configProxy(cfg)

	// Assert that no error was returned
	assert.NoError(t, err)

	// Assert that a proxy was returned
	assert.NotNil(t, p)

	assert.Equal(t, 1, len(p.Addons))
}
