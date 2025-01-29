package models

// Quepasa options, setted on start, so if want to changed then, you have to restart the entire service
type QpOptions struct {

	// default log level
	LogLevel string `json:"loglevel,omitempty"`
}
