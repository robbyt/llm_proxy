package addons

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"

	md "github.com/robbyt/llm_proxy/addons/megadumper"
	"github.com/robbyt/llm_proxy/addons/megadumper/formatters"
	"github.com/robbyt/llm_proxy/addons/megadumper/writers"
	"github.com/robbyt/llm_proxy/config"
	"github.com/robbyt/llm_proxy/schema"
)

type MegaDumpAddon struct {
	px.BaseAddon
	formatter         formatters.MegaDumpFormatter
	logSources        config.LogSourceConfig
	writers           []writers.MegaDumpWriter
	filterReqHeaders  []string
	filterRespHeaders []string
	wg                sync.WaitGroup
	closed            atomic.Bool
}

// Requestheaders is a callback that will receive a "flow" from the proxy, will create a
// NewLogDumpContainer and will use the embedded writers to finally write the log.
func (d *MegaDumpAddon) Requestheaders(f *px.Flow) {
	if d.closed.Load() {
		log.Warn("MegaDirDumper is being closed, not logging a request")
		return
	}

	start := time.Now()

	d.wg.Add(1) // for blocking this addon during shutdown in .Close()
	go func() {
		defer d.wg.Done()
		<-f.Done()
		doneAt := time.Since(start).Milliseconds()

		// load the selected fields into a container object
		dumpContainer := schema.NewLogDumpContainer(*f, d.logSources, doneAt, d.filterReqHeaders, d.filterRespHeaders)

		id := f.Id.String() // TODO: is the internal request ID unique enough?

		// format the container object, reformatted into a byte array
		formattedDump, err := d.formatter.Read(dumpContainer)
		if err != nil {
			log.Error(err)
			return
		}

		// write the formatted log data to... somewhere
		for _, w := range d.writers {
			if w == nil {
				log.Error("Writer is nil, skipping")
				continue
			}
			_, err := w.Write(id, formattedDump)
			if err != nil {
				log.Error(err)
				continue
			}
		}
	}()
}

func (d *MegaDumpAddon) String() string {
	return "MegaDirDumper"
}

func (d *MegaDumpAddon) Close() error {
	if !d.closed.Swap(true) {
		log.Debug("Waiting for MegaDirDumper shutdown...")
		d.wg.Wait()
	}

	return nil
}

// NewMegaDirDumper creates a new dumper that creates a new log file for each request
func NewMegaDirDumper(
	logTarget string,
	logFormat md.LogFormat,
	logSources config.LogSourceConfig,
	logDestinations []md.LogDestination,
	filterReqHeaders, filterRespHeaders []string,
) (*MegaDumpAddon, error) {
	var f formatters.MegaDumpFormatter
	var w = make([]writers.MegaDumpWriter, 0)

	switch logFormat {
	case md.Format_JSON:
		log.Debug("Logging format set to JSON file")
		f = &formatters.JSON{}
	case md.Format_PLAINTEXT:
		log.Debug("Logging format set to plaintext file")
		f = &formatters.PlainText{}
	default:
		return nil, fmt.Errorf("invalid log format: %v", logFormat)
	}

	for _, logDest := range logDestinations {
		switch logDest {
		case md.WriteToDir:
			log.Debug("Directory logger enabled")
			dirWriter, err := writers.NewToDir(logTarget, logFormat)
			if err != nil {
				return nil, err
			}
			w = append(w, dirWriter)

		case md.WriteToFile:
			log.Debug("Single file logger enabled")
			fileWriter, err := writers.NewToFile(logTarget, logFormat)
			if err != nil {
				return nil, err
			}
			w = append(w, fileWriter)
		case md.WriteToStdOut:
			log.Debug("Standard out logger enabled")
			stdoutWriter, err := writers.NewToStdOut()
			if err != nil {
				return nil, err
			}
			w = append(w, stdoutWriter)
		default:
			return nil, fmt.Errorf("invalid log destination: %v", logDest)
		}
	}

	mda := &MegaDumpAddon{
		formatter:         f,
		logSources:        logSources,
		writers:           w,
		filterReqHeaders:  filterReqHeaders,
		filterRespHeaders: filterRespHeaders,
	}
	mda.closed.Store(false) // initialize the atomic bool with closed = false

	log.Debugf("Created MegaDirDumper with %s sources and %v writer(s)", logSources.String(), len(w))

	return mda, nil

}
