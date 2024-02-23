package proxy

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

// Run is the main entry point for the proxy, full of imperative code, config processing, and error handling
func Run(cfg *config.Config) error {
	debugLevel := cfg.GetDebugLevel()

	ca, err := newCA(cfg.CertDir)
	if err != nil {
		return fmt.Errorf("setupCA error: %v", err)
	}

	p, err := newProxy(debugLevel, cfg.Listen, cfg.InsecureSkipVerifyTLS, ca)
	if err != nil {
		return fmt.Errorf("failed to create proxy: %v", err)
	}

	if debugLevel > 0 {
		log.Debugf("Debug level is set to %v, enabling traffic logging to terminal", debugLevel)
		p.AddAddon(addons.NewStdOutLogger())
	}

	if !cfg.NoHttpUpgrader {
		// upgrade all http requests to https
		log.Debug("NoHttpUpgrader is false, enabling http to https upgrade")
		p.AddAddon(&addons.SchemeUpgrader{})
	}

	if cfg.OutputDir != "" {
		log.Debugf("OutputDir is set, dumping traffic to: %v", cfg.OutputDir)

		// creates a formatted []LogSource containing various enum settings, pulled from the bools set in the config
		logSources := config.LogSourceConfig{
			LogRequestHeaders:  !cfg.NoLogReqHeaders,
			LogRequestBody:     !cfg.NoLogReqBody,
			LogResponseHeaders: !cfg.NoLogRespHeaders,
			LogResponseBody:    !cfg.NoLogRespBody,
		}

		log.Debugf("Will log these fields: %v", logSources)

		// create and configure MegaDirDumper addon object
		dumper, err := addons.NewMegaDirDumper(
			cfg.OutputDir,
			md.Format_JSON,
			logSources,
			[]md.LogDestination{md.WriteToDir},
			cfg.FilterReqHeaders, cfg.FilterRespHeaders,
		)
		if err != nil {
			return fmt.Errorf("failed to create dumper: %v", err)
		}

		// add the dumper to the proxy
		p.AddAddon(dumper)
	}

	// setup background signal handler for clean shutdown
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		log.Info("Received SIGINT, shutting down now...")
		p.Shutdown(context.TODO())
	}()

	// log.Infof("Starting proxy on: %v", cfg.Listen)

	// block here while the proxy is running
	err = p.Start()
	if err != nil {
		/*
			when `p` gets a shutdown signal, it returns with an error "http: Server closed"
			We want handle that error here, and avoid passing it back up the stack to the caller.
			A string compare is ugly, but I can't find where the shutdown error obj is defined.
		*/
		if err.Error() != "http: Server closed" {
			return fmt.Errorf("failed to start proxy: %v", err)
		}
	}

	return nil
}
