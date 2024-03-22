package proxy

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"testing"

	"github.com/robbyt/llm_proxy/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testServerListenAddr = "localhost:8182"
)

func webServer(hitCounter *atomic.Int32) (*http.Server, func()) {
	if hitCounter == nil {
		panic("hitCounter must be non-nil")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// increment the counter
		hitCounter.Add(1)
		resp := fmt.Sprintf("hits: %d\n", hitCounter.Load())

		w.WriteHeader(200)
		w.Write([]byte(resp))
	})

	srv := &http.Server{
		Addr:    testServerListenAddr,
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
	}, nil
}

func TestProxySimple(t *testing.T) {

	// Create a simple proxy config
	cfg := config.NewDefaultConfig()
	cfg.CertDir = t.TempDir()
	cfg.NoHttpUpgrader = true // disable TLS because our test server doesn't support it
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

	srv, srvShutdown := webServer(hitCounter)
	require.NotNil(t, srv)
	require.NotNil(t, srvShutdown)

	// Create a client that will use the proxy
	client, err := httpClient("http://" + cfg.Listen)
	require.NoError(t, err)

	// make a request using that client, through the proxy
	resp1, err := client.Get("http://" + testServerListenAddr)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	// check the response body from req1
	body1, err := io.ReadAll(resp1.Body)
	require.NoError(t, err)
	assert.Equal(t, []byte("hits: 1\n"), body1)

	// make another request using that client, through the proxy
	resp2, err := client.Get("http://" + testServerListenAddr)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	// check the response body from req2
	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)
	assert.Equal(t, []byte("hits: 2\n"), body2)

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
