package comm

import (
	"fmt"
	"net/http"
	"strings"
)

type HttpHeader http.Header

func (header HttpHeader) String() string {
	if header == nil {
		return ""
	}
	var builder strings.Builder
	for key, values := range header {
		builder.WriteString(fmt.Sprintf("%s: %s\n", key, strings.Join(values, ", ")))
	}
	return builder.String()
}
