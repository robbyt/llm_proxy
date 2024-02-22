package addons

import (
	"fmt"

	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"

	md "github.com/robbyt/llm_proxy/addons/megadumper"
	"github.com/robbyt/llm_proxy/addons/megadumper/formatters"
	"github.com/robbyt/llm_proxy/addons/megadumper/writers"
	"github.com/robbyt/llm_proxy/config"
)

type MegaDumpAddon struct {
	px.BaseAddon
	formatter         formatters.MegaDumpFormatter
	logSources        config.LogSourceConfig
	writers           []writers.MegaDumpWriter
	filterReqHeaders  []string
	filterRespHeaders []string
}

// Requestheaders is a callback for the Requestheaders event
func (d *MegaDumpAddon) Requestheaders(f *px.Flow) {
	go func() {
		<-f.Done()
		// load the selected fields into a container object
		dumpContainer := md.NewLogDumpContainer(*f, d.logSources, d.filterReqHeaders, d.filterRespHeaders)

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
			bytesWritten, err := w.Write(id, formattedDump)
			if err != nil {
				log.Error(err)
				continue
			}
			log.Debugf("Writer %v wrote %v bytes for %v", w, bytesWritten, id)
		}
	}()
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

	log.Debugf("Created MegaDirDumper with %s sources and %v writer(s)", logSources.String(), len(w))

	return mda, nil

}
