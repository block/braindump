package output

import (
	"encoding/json"
	"io"
	"time"

	"github.com/block/braindump/internal/model"
)

const Version = "1.0.0"

// Writer handles writing output in JSON format
type Writer struct {
	writer io.Writer
	pretty bool
}

// NewWriter creates a new output writer
func NewWriter(w io.Writer, pretty bool) *Writer {
	return &Writer{
		writer: w,
		pretty: pretty,
	}
}

// Write writes sessions to output
func (w *Writer) Write(sessions []model.Session) error {
	output := model.Output{
		Version:     Version,
		GeneratedAt: time.Now(),
		Sessions:    sessions,
	}

	var data []byte
	var err error

	if w.pretty {
		data, err = json.MarshalIndent(output, "", "  ")
	} else {
		data, err = json.Marshal(output)
	}

	if err != nil {
		return err
	}

	_, err = w.writer.Write(data)
	if err != nil {
		return err
	}

	// Add newline at end
	_, err = w.writer.Write([]byte("\n"))
	return err
}
