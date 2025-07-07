package log

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net"

	"github.com/guidomantilla/yarumo/pkg/common/utils"
)

func Configure(name string, version string, opts ...Option) zerolog.Logger {

	conn, err := net.Dial("tcp", "localhost:5044")
	if err != nil {
		fmt.Println("Error connecting to the server:", err)
	}

	options := NewOptions(opts...)
	logger := zerolog.New(conn).With()

	if utils.NotEmpty(name) {
		logger = logger.Str("name", name)
	}
	if utils.NotEmpty(version) {
		logger = logger.Str("version", version)
	}

	if options.caller {
		logger = logger.Caller()
	}

	log.Logger = logger.Timestamp().Logger()
	return log.Logger
}
