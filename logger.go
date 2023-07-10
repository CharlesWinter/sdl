// Package logging is a lightweight wrapper around logrus. The wrapper provides
// several contraints/ functions on top of logrus:
// It ensures logs are in a format to be well consumed by stackdriver.
// It will report any errors to stackdriver by writing errors in the correct
// way
package logging

import (
	"fmt"
	"io"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
)

type serviceContext struct {
	serviceName string
	version     string
}

// Logger is a wrapper around a logrus entry. It embedds the entry, giving
// the user access to all the logrus methods if they want, but certain methods
// are overridden to provide output in a strackdriver compatiable format
type Logger struct {
	*logrus.Entry
	serviceContext serviceContext
}

// Config is a simple structure used to configure a logger in New
type Config struct {
	// LoggingLevel determines above which level to actually log.
	// For example setting this to error would ignore all logs below, such as
	// debug and trace
	LoggingLevel  string
	ServiceName   string
	Version       string
	WriteLocation io.Writer
}

// New returns a logger configured to output JSON and with some of the
// stackdriver required fields already set.
func New(conf Config) (*Logger, error) {
	logrusLogger := logrus.New()

	// by not always defaulting to stdOut allows testing through writing to
	// buffers
	if conf.WriteLocation != nil {
		logrusLogger.Out = conf.WriteLocation
	}

	logrusLogger.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			// stackdriver expects the field to be "message" not msg
			logrus.FieldKeyMsg: "message",
			// stackdriver expects the field to be called "timestamp" not time
			logrus.FieldKeyTime: "timestamp",
			// stackdriver expects the field to be called "severity" not level
			logrus.FieldKeyLevel: "severity",
		},
		TimestampFormat: time.RFC3339Nano,
	})

	// Anything below this log level will not be logged
	logLvl, err := logrus.ParseLevel(conf.LoggingLevel)
	if err != nil {
		return nil, err
	}
	logrusLogger.SetLevel(logLvl)
	return &Logger{
		Entry: logrus.NewEntry(logrusLogger),
		serviceContext: serviceContext{
			serviceName: conf.ServiceName,
			version:     conf.Version,
		},
	}, nil
}

func (l Logger) Error(err error) {
	// append the error to the top of the stacktrace to appease stackdriver Error
	// reporting formatting. Putting the error in message makes it read nice in
	// the logs, and at the top of the trace makes it read nicely in error
	// reporting
	stackTrace := fmt.Sprintf("%s\n%s", err, string(debug.Stack()))

	// https://cloud.google.com/error-reporting/reference/rest/v1beta1/ErrorEvent
	errorFields := logrus.Fields{
		"serviceContext": logrus.Fields{
			"service": l.serviceContext.serviceName,
			"version": l.serviceContext.version,
		},
		"eventTime": time.Now(),
		// https://cloud.google.com/error-reporting/docs/formatting-error-messages#json_representation
		"stack_trace": stackTrace,
	}

	l.Entry.WithFields(errorFields).Error(err)
}

// Info logs an entry at the info level
func (l Logger) Info(args ...interface{}) {
	l.Entry.Info(args...)
}

// Payload is a redefinement of the logrus.Fields type
type Payload map[string]interface{}

func convertToLogrusFields(jp Payload) logrus.Fields {
	lf := make(logrus.Fields, len(jp))
	for k, v := range jp {
		lf[k] = v
	}
	return lf
}

// JSONPayload adheres to the stackdriver LogEntry format specified here
// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry This is
// difference to the "error" event, which requires a message as its main
// informatory field.
//
// Passing marshalled JSON/a map/ a struct to the Info method prints a string
// representation not nested JSON. Logrus seems to produce the best JSON output
// with the "WithFields" method
func (l Logger) JSONPayload(jp Payload) *logrus.Entry {
	return l.Entry.WithFields(convertToLogrusFields(jp))
}
