package proxy

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kardianos/mitmproxy/cert"
	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"

	"github.com/robbyt/llm_proxy/addons"
	md "github.com/robbyt/llm_proxy/addons/megadumper"
	"github.com/robbyt/llm_proxy/config"
)

func newCA(certDir string) (*cert.CA, error) {
	if certDir == "" {
		log.Debug("No cert dir specified, defaulting to ~/.mitmproxy/")
	} else {
		log.Debugf("Loading certs from directory: %v", certDir)
	}

	l, err := cert.NewPathLoader(certDir)
	if err != nil {
		return nil, fmt.Errorf("unable to create or load certs from %v: %v", certDir, err)
	}

	ca, err := cert.New(l)
	if err != nil {
		return nil, fmt.Errorf("problem with CA config: %v", err)
	}

	return ca, nil
}

func newProxy(debugLevel int, listenOn string, skipVerifyTLS bool, ca *cert.CA) (*px.Proxy, error) {
	opts := &px.Options{
		Debug:                 debugLevel,
		Addr:                  listenOn,
		InsecureSkipVerifyTLS: skipVerifyTLS,
		CA:                    ca,
		StreamLargeBodies:     1024 * 1024 * 100, // responses larger than 100MB will be streamed
	}

	p, err := px.NewProxy(opts)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// configProxy is the main entry point for the proxy, full of imperative code, config processing, and error handling
func configProxy(cfg *config.Config) (*px.Proxy, error) {
	// create a slice of LogDestination objects, which are used to configure the MegaDirDumper addon
	logDest := []md.LogDestination{}
	debugLevel := cfg.IsDebugEnabled()

	ca, err := newCA(cfg.CertDir)
	if err != nil {
		return nil, fmt.Errorf("setupCA error: %v", err)
	}

	p, err := newProxy(debugLevel, cfg.Listen, cfg.InsecureSkipVerifyTLS, ca)
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy: %v", err)
	}

	if cfg.IsVerboseOrHigher() {
		log.Debugf("Enabling traffic logging to terminal")
		logDest = append(logDest, md.WriteToStdOut)
		p.AddAddon(addons.NewStdOutLogger())
	}

	if !cfg.NoHttpUpgrader {
		// upgrade all http requests to https
		log.Debug("NoHttpUpgrader is false, enabling http to https upgrade")
		p.AddAddon(&addons.SchemeUpgrader{})
	}

	switch cfg.AppMode {
	case config.MockMode:
		log.Debugf("AppMode is MockMode, adding mock/cache addon")
		// p.AddAddon(addons.NewMockCache(cfg.OutputDir, cfg.FilterReqHeaders, cfg.FilterRespHeaders))
	case config.DirLoggerMode:
		log.Debugf("AppMode is DirLoggerMode, dumping traffic to: %v", cfg.OutputDir)

		// struct of bools to toggle the various log outputs
		logSources := config.LogSourceConfig{
			LogConnectionStats: !cfg.NoLogConnStats,
			LogRequestHeaders:  !cfg.NoLogReqHeaders,
			LogRequestBody:     !cfg.NoLogReqBody,
			LogResponseHeaders: !cfg.NoLogRespHeaders,
			LogResponseBody:    !cfg.NoLogRespBody,
		}

		// append the WriteToDir LogDestination to the logDest slice, so megadumper will write to disk
		logDest = append(logDest, md.WriteToDir)

		// create and configure MegaDirDumper addon object
		dumper, err := addons.NewMegaDirDumper(
			cfg.OutputDir,
			md.Format_JSON,
			logSources,
			logDest,
			cfg.FilterReqHeaders, cfg.FilterRespHeaders,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create dumper: %v", err)
		}

		// add the dumper to the proxy
		p.AddAddon(dumper)
	case config.SimpleMode:
		log.Debugf("AppMode is SimpleMode, no addons will be added")
	default:
		return nil, fmt.Errorf("unknown app mode: %v", cfg.AppMode)
	}

	return p, nil
}

// startProxy receives a pointer to a proxy object, runs it, and handles the shutdown signal
func startProxy(p *px.Proxy) error {
	// setup background signal handler for clean shutdown
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-ch
		log.Info("Received SIGINT, shutting down now...")

		// Then close all of the addon connections
		for _, addon := range p.Addons {
			myAddon, ok := addon.(addons.LLM_Addon)
			if !ok {
				continue
			}
			log.Debugf("Closing addon: %s", myAddon)
			if err := myAddon.Close(); err != nil {
				log.Errorf("Error closing addon: %v", err)
			}
		}
		// Close the http client/server connections first
		log.Debug("Closing proxy server...")

		// Create a context that will be cancelled after N seconds
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(60*time.Second))
		defer cancel()

		if err := p.Shutdown(ctx); err != nil {
			log.Errorf("Error shutting down proxy server: %v", err)
		}
	}()

	// block here while the proxy is running
	if err := p.Start(); err != http.ErrServerClosed {
		return fmt.Errorf("proxy server error: %v", err)
	}
	return nil
}

// Run is the main entry point for the proxy, configures the proxy and runs it
func Run(cfg *config.Config) error {
	p, err := configProxy(cfg)
	if err != nil {
		return fmt.Errorf("failed to configure proxy: %v", err)
	}

	if err := startProxy(p); err != nil {
		return fmt.Errorf("failed to start proxy: %v", err)
	}

	return nil
}
