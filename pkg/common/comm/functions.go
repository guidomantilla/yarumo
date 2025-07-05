package comm

import (
	"bytes"
	"fmt"
	"io"
)

func ToReader(body []byte) io.Reader {
	if body == nil {
		return nil
	}
	return io.NopCloser(bytes.NewReader(body))
}

func ToReadNopCloser(reader io.ReadCloser) (io.ReadCloser, []byte, error) {
	if reader == nil {
		return nil, nil, fmt.Errorf("nil reader")
	}

	buffer, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading from reader: %w", err)
	}

	readerNopCloser := io.NopCloser(bytes.NewReader(buffer))
	return readerNopCloser, buffer, nil
}
