package http

// Logger defines the interface used by each handler for logging purposes.
type Logger interface {
	Printf(format string, v ...interface{})
}
