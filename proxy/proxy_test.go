package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/robbyt/llm_proxy/config"
	"github.com/robbyt/llm_proxy/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// ugly hack to wait for background async
	defaultSleepTime = 1 * time.Second
	outputSubdir     = "output"
	certSubdir       = "certs"
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

func runWebServer(hitCounter *atomic.Int32, listenAddr string) (*http.Server, func()) {
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

func runProxy(proxyPort, tempDir string, proxyAppMode config.AppMode) (shutdownFunc func(), err error) {
	// Create a simple proxy config
	cfg := config.NewDefaultConfig()
	cfg.Listen = proxyPort
	cfg.CertDir = filepath.Join(tempDir, certSubdir)
	cfg.OutputDir = filepath.Join(tempDir, outputSubdir)
	cfg.Debug = true
	cfg.AppMode = proxyAppMode
	cfg.NoHttpUpgrader = true // disable TLS because our test server doesn't support it

	// create a proxy with the test config
	p, err := configProxy(cfg)
	if err != nil {
		return nil, err
	}

	// setup external control of the proxy
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// start the proxy in the background
	go func() {
		err = startProxy(p, shutdownChan)
		if err != nil {
			log.Fatal(err)
		}
	}()
	// sleep while waiting for the proxy to start
	time.Sleep(defaultSleepTime)

	return func() {
		// returns a function that can be called to shutdown the proxy goroutine
		shutdownChan <- os.Interrupt
	}, nil
}

func BenchmarkProxySimple(b *testing.B) {
	// create a proxy with a test config
	proxyPort, err := getFreePort()
	require.NoError(b, err)
	tmpDir := b.TempDir()
	proxyShutdown, err := runProxy(proxyPort, tmpDir, config.SimpleMode)
	require.NoError(b, err)

	// Start a basic web server on another port
	hitCounter := new(atomic.Int32)
	testServerPort, err := getFreePort()
	require.NoError(b, err)
	srv, srvShutdown := runWebServer(hitCounter, testServerPort)
	require.NotNil(b, srv)
	require.NotNil(b, srvShutdown)

	// Create an http client that will use the proxy to connect to the web server
	client, err := httpClient("http://" + proxyPort)
	require.NoError(b, err)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hitCounter.Store(0) // reset the counter
		// make a request using that client, through the proxy
		b.StartTimer()
		resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
		b.StopTimer()
		require.NoError(b, err)
		assert.Equal(b, 200, resp.StatusCode)
	}
	b.Cleanup(func() {
		srvShutdown()
		proxyShutdown()
	})
}

func TestProxySimple(t *testing.T) {
	// create a proxy with a test config
	proxyPort, err := getFreePort()
	require.NoError(t, err)
	tmpDir := t.TempDir()
	proxyShutdown, err := runProxy(proxyPort, tmpDir, config.SimpleMode)
	require.NoError(t, err)

	// Start a basic web server on another port
	hitCounter := new(atomic.Int32)
	testServerPort, err := getFreePort()
	require.NoError(t, err)
	srv, srvShutdown := runWebServer(hitCounter, testServerPort)
	require.NotNil(t, srv)
	require.NotNil(t, srvShutdown)

	// Create an http client that will use the proxy to connect to the web server
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
		srvShutdown()
		proxyShutdown()
	})
}

func TestProxyDirLoggerMode(t *testing.T) {
	// create a proxy with a test config
	proxyPort, err := getFreePort()
	require.NoError(t, err)
	tmpDir := t.TempDir()
	proxyShutdown, err := runProxy(proxyPort, tmpDir, config.DirLoggerMode)
	require.NoError(t, err)

	// Start a basic web server on another port
	hitCounter := new(atomic.Int32)
	testServerPort, err := getFreePort()
	require.NoError(t, err)
	srv, srvShutdown := runWebServer(hitCounter, testServerPort)
	require.NotNil(t, srv)
	require.NotNil(t, srvShutdown)

	// Create an http client that will use the proxy to connect to the web server
	client, err := httpClient("http://" + proxyPort)
	require.NoError(t, err)

	t.Run("TestDirLoggerNormal", func(t *testing.T) {
		hitCounter.Store(0) // reset the counter
		// make a request using that client, through the proxy
		resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.Equal(t, int32(1), hitCounter.Load())

		// check the response body from req1
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, []byte("hits: 1\n"), body)

		// sleep to allow the proxy to write the log file
		time.Sleep(defaultSleepTime)

		// check that the log file was created
		logFiles, err := filepath.Glob(filepath.Join(tmpDir, outputSubdir, "*"))
		require.NoError(t, err)
		require.Equal(t, 1, len(logFiles))

		// read the log file, and check that it contains the expected content
		logFile, err := os.ReadFile(logFiles[0])
		require.NoError(t, err)
		assert.Contains(t, string(logFile), `hits: 1`)

		// delete that log file, and try again
		err = os.Remove(logFiles[0])
		require.NoError(t, err)

		// make another request using that client, through the proxy
		resp, err = client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.Equal(t, int32(2), hitCounter.Load())

		// sleep to allow the proxy to write the log file
		time.Sleep(defaultSleepTime)

		// check the log
		logFiles, err = filepath.Glob(filepath.Join(tmpDir, outputSubdir, "*"))
		require.NoError(t, err)
		require.Equal(t, 1, len(logFiles))

		// read the log file, and check that it contains the expected content
		logFile, err = os.ReadFile(logFiles[0])
		defer os.Remove(logFiles[0])
		require.NoError(t, err)
		assert.Contains(t, string(logFile), `hits: 2`)
	})

	t.Run("TestDirLoggerJSON", func(t *testing.T) {
		hitCounter.Store(0) // reset the counter

		// make another request using that client, through the proxy
		resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.Equal(t, int32(1), hitCounter.Load())

		// sleep to allow the proxy to write the log file
		time.Sleep(defaultSleepTime)

		// check the log
		logFiles, err := filepath.Glob(filepath.Join(tmpDir, outputSubdir, "*"))
		require.NoError(t, err)
		require.Equal(t, 1, len(logFiles))

		// read the log file, and check that it contains the expected content
		logFile, err := os.ReadFile(logFiles[0])
		require.NoError(t, err)

		// marshal the log file to a logDumpContainer
		lDump := schema.LogDumpContainer{}
		err = json.Unmarshal(logFile, &lDump)
		require.NoError(t, err)
		fmt.Println(string(logFile))

		// check the logDumpContainer
		assert.Equal(t, schema.SchemaVersion, lDump.SchemaVersion)
		assert.NotNil(t, lDump.Timestamp)
		assert.NotNil(t, lDump.ConnectionStats)

		require.NotNil(t, lDump.Request)
		assert.Equal(t, "POST", lDump.ConnectionStats.Method)

		require.NotNil(t, lDump.Response)
		assert.Equal(t, http.StatusOK, lDump.ConnectionStats.ResponseCode)
		assert.Equal(t, "hits: 1\n", lDump.Response.Body)

	})

	// done with tests, send shutdown signals
	t.Cleanup(func() {
		srvShutdown()
		proxyShutdown()
	})
}

func TestProxyCache(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a simple proxy config
	cfg := config.NewDefaultConfig()
	cfg.CertDir = tmpDir + "/certs"
	cfg.Cache.Dir = tmpDir + "/cache"
	cfg.NoHttpUpgrader = true // disable TLS because our test server doesn't support it
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
	assert.Equal(t, int32(1), hitCounter.Load())
	assert.Equal(t, resp1.Header.Get("X-Cache"), "MISS")

	// make another request using that client, through the proxy
	resp2, err := client.Get("http://" + testServerListenAddr)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	// check the response body from req2 (should be the cached response with value=1, not the incremented value 2)
	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)
	assert.Equal(t, []byte("hits: 1\n"), body2)
	assert.Equal(t, int32(1), hitCounter.Load()) // the counter should not have incremented because the server shouldn't have been hit
	assert.Equal(t, resp2.Header.Get("X-Cache"), "HIT")

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
