package sdl_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/CharlesWinter/sdl"
	"github.com/sirupsen/logrus"
)

func TestLoggingErrors(t *testing.T) {
	t.Run("it logs an error ", func(t *testing.T) {
		buffer := new(bytes.Buffer)
		subject, err := sdl.New(sdl.Config{
			LoggingLevel:  "INFO",
			ServiceName:   "test",
			Version:       "1.0.0",
			WriteLocation: buffer,
		})
		if err != nil {
			t.Fatal(err)
		}

		subject.Error(errors.New("oh no"))

		// gross cast, but it lets us check some of the fields easily
		var got logrus.Fields
		if err := json.Unmarshal(buffer.Bytes(), &got); err != nil {
			t.Fatal(err)
		}

		if got["message"] != "oh no" {
			t.Fatalf("expected a message, got json %s", string(buffer.Bytes()))
		}
	})
}
