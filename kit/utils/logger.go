package utils

import (
	"encoding/json"
	"fmt"
	kitlog "github.com/go-kit/kit/log"
	"io"
)

// FormattedJSONLogger is a go-kit logger that formats output as JSON.
type FormattedJSONLogger struct {
	w io.Writer
}

// NewFormattedJSONLogger initializes a new FormattedJSONLogger.
func NewFormattedJSONLogger(w io.Writer) kitlog.Logger {
	return &FormattedJSONLogger{w}
}

// Log implements the go-kit log.Logger interface.
func (l *FormattedJSONLogger) Log(keyvals ...interface{}) error {
	m := make(map[string]interface{}, (len(keyvals)+1)/2)

	for i := 0; i < len(keyvals); i += 2 {
		var v interface{} = "(!MISSING)"
		if i+1 < len(keyvals) {
			v = keyvals[i+1]
		}
		m[fmt.Sprintf("%s", keyvals[i])] = v
	}

	b, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(l.w, string(b))
	return err
}
