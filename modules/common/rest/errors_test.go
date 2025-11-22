package rest

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestErrorFormatting(t *testing.T) {
	err := ErrCall(errors.New("root-cause"))
	if !strings.Contains(err.Error(), "rest request rest-request error: root-cause") {
		t.Fatalf("unexpected error string: %s", err.Error())
	}
}

func TestHTTPErrorFormatting(t *testing.T) {
	he := &HTTPError{StatusCode: 418, Status: http.StatusText(418)}
	if !strings.Contains(he.Error(), "unexpected status code 418") {
		t.Fatalf("unexpected http error string: %s", he.Error())
	}
}

func TestDecodeResponseErrorFormatting(t *testing.T) {
	// Use a non-pointer generic type to satisfy assert.NotNil(e.T)
	de := &DecodeResponseError[sample]{ContentType: "text/plain", T: sample{}}
	s := de.Error()
	if !strings.Contains(s, "content type text/plain not supported") {
		t.Fatalf("unexpected decode error string: %s", s)
	}
}
