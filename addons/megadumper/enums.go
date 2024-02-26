package megadumper

// LogFormat is an enum for the format of the logs
type LogFormat int

func (f LogFormat) FileExtension() string {
	switch f {
	case Format_JSON:
		return "json"
	case Format_PLAINTEXT:
		return "log"
	default:
		return ""
	}
}

const (
	// Format_JSON logs in Format_JSON format
	Format_JSON LogFormat = iota

	// Format_PLAINTEXT logs in plain text format
	Format_PLAINTEXT
)

// LogDestination is an enum for the destination for where the logs are stored
type LogDestination int

const (
	// WriteToFile logs to a single file
	WriteToFile LogDestination = iota

	// WriteToDir logs to a directory
	WriteToDir

	// WriteToStdOut logs to standard out
	WriteToStdOut
)
