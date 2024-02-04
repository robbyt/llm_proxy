package proxy

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/kardianos/mitmproxy/cert"
	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"

	"github.com/robbyt/llm_proxy/addons"
	"github.com/robbyt/llm_proxy/config"
)

func getDebugLevel() int {
	if log.GetLevel() >= log.DebugLevel {
		return 1
	}
	return 0
}

func setupCA(certDir string) (*cert.CA, error) {
	log.Debugf("Loading certs from directory: %v", certDir)
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

func Run(cfg *config.Config) error {
	debugLevel := getDebugLevel()

	log.Debugf("Loading certs from directory: %v", cfg.CertDir)
	ca, err := setupCA(cfg.CertDir)
	if err != nil {
		return fmt.Errorf("setupCA error: %v", err)
	}

	opts := &px.Options{
		Debug:                 debugLevel,
		Addr:                  cfg.Listen,
		InsecureSkipVerifyTLS: cfg.InsecureSkipVerifyTLS,
		CA:                    ca,
		StreamLargeBodies:     1024 * 1024 * 100, // responses larger than 100MB will be streamed
	}

	p, err := px.NewProxy(opts)
	if err != nil {
		return fmt.Errorf("failed to create proxy: %v", err)
	}

	if debugLevel > 0 {
		log.Debugf("Debug level is set to %v, enabling stdout request logs", debugLevel)
		p.AddAddon(&addons.StdOutLogger{})
	}

	if cfg.NoHttpUpgrader {
		log.Debug("NoHttpUpgrader is true, not upgrading http requests to https")
	} else {
		// upgrade all http requests to https
		log.Debug("NoHttpUpgrader is false, enabling http to https upgrade")
		p.AddAddon(&addons.SchemeUpgrader{})
	}

	if cfg.OutputDir == "" {
		log.Debug("OutputDir is empty, skipping the request dump")
	} else {
		log.Debugf("OutputDir is set to %v, enabling request dump", cfg.OutputDir)
		dumper, err := addons.NewDumperWithLogRoot(cfg.OutputDir, addons.WRITE_REQ_BODY_AND_RESP_BODY)
		if err != nil {
			return fmt.Errorf("failed to create dumper: %v", err)
		}
		p.AddAddon(dumper)
	}

	// setup background signal handler for clean shutdown
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		log.Info("Received SIGINT, shutting down now...")
		p.Shutdown(context.TODO())
	}()

	log.Infof("Starting proxy on: %v", cfg.Listen)

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
