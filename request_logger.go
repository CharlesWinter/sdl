package sdl

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
)

// RequestLogger is used for logging information about a request at either the
// error or info level
type RequestLogger struct {
	method    string
	url       string
	userAgent string
	remoteIP  string
	status    int
	protocol  string
	logger    *Logger
}

// HTTPRequestFields are used to populate the httpRequest section of the log
// entry
type HTTPRequestFields struct {
	Request *http.Request
	Status  int
}

func stripQueryParam(u *url.URL, stripKey string) string {
	q := u.Query()
	q.Del(stripKey)
	u.RawQuery = q.Encode()
	return u.String()
}

// NewRequestLogger returns a logger designed for logging requests. Due to the
// very different formatting required for logging requests, its a completely
// different type to the base logger with its own methods
func (l *Logger) NewRequestLogger(rf HTTPRequestFields) (*RequestLogger, error) {
	if rf.Request == nil {
		return nil, errors.New("cannot log nil request")
	}

	return &RequestLogger{
		method:    rf.Request.Method,
		url:       stripQueryParam(rf.Request.URL, "key"),
		userAgent: rf.Request.UserAgent(),
		remoteIP:  rf.Request.RemoteAddr,
		status:    rf.Status,
		protocol:  rf.Request.Proto,
		logger:    l,
	}, nil
}

// InfoJSONPayload a message in a log entry is fundementally useless; everything
// depends on the JSONPayload.
// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry
// InfoJSONPayload logs a map as jsonPayload at info level, with the http
// request fields populated as well
func (rl *RequestLogger) InfoJSONPayload(jp Payload) {
	logFields := logrus.Fields{
		"httpRequest": logrus.Fields{
			"requestMethod": rl.method,
			"requestUrl":    rl.url,
			"userAgent":     rl.userAgent,
			"remoteIp":      rl.remoteIP,
			"status":        rl.status,
			"protocol":      rl.protocol,
		},
	}

	for k, v := range convertToLogrusFields(jp) {
		logFields[k] = v
	}
	rl.logger.WithFields(logFields).Info("")
}

// Error is intented to log an error that was generated in responded to
// a request. Infuriatingly, stackdriver requests the http information in a
// different format for LogEntry and an error report
// https://cloud.google.com/error-reporting/docs/formatting-error-messages
// so the RequestLogger method below cannot be reused. As its unlikely to want
// to log multiple errors within the same request context, this method doesn't
// return an error and just logs
func (rl *RequestLogger) Error(err error) {
	// append the error to the top of the stacktrace to appease stackdriver Error
	// reporting formatting. Putting the error in message makes it read nice in
	// the logs, and at the top of the trace makes it read nicely in error
	// reporting
	stackTrace := fmt.Sprintf("%s\n%s", err, string(debug.Stack()))

	// https://cloud.google.com/error-reporting/reference/rest/v1beta1/ErrorEvent
	errorFields := logrus.Fields{
		"serviceContext": logrus.Fields{
			"service": rl.logger.serviceContext.serviceName,
			"version": rl.logger.serviceContext.version,
		},
		"context": logrus.Fields{
			"httpRequest": logrus.Fields{
				"method":             rl.method,
				"url":                rl.url,
				"userAgent":          rl.userAgent,
				"responseStatusCode": rl.status,
				"remoteIp":           rl.remoteIP,
			},
		},
		"eventTime": time.Now(),
		// https://cloud.google.com/error-reporting/docs/formatting-error-messages#json_representation
		"stack_trace": stackTrace,
		// adding the type field forces stackdriver to parse the event as an error.
		// It might never actually be needed, but stackdriver can sometimes get
		// confused as to what to do with a log entry. Adding this means it will
		// always parse as an error event even if some other fields are added later
		"@type": "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent",
	}

	rl.logger.Entry.WithFields(errorFields).Error(err)
}
