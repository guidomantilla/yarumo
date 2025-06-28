package assert

import (
	"github.com/guidomantilla/yarumo/pkg/utils"
	"github.com/rs/zerolog/log"
)

func NotEmpty(object any, message string) {
	if utils.Empty(object) {
		log.Fatal().Msg(message)
	}
}

func NotNil(object any, message string) {
	if utils.Nil(object) {
		log.Fatal().Msg(message)
	}
}

func Equal(val1 any, val2 any, message string) {
	if utils.NotEqual(val1, val2) {
		log.Fatal().Msg(message)
	}
}

func NotEqual(val1 any, val2 any, message string) {
	if utils.Equal(val1, val2) {
		log.Fatal().Msg(message)
	}
}

func True(condition bool, message string) {
	if !condition {
		log.Fatal().Msg(message)
	}
}

func False(condition bool, message string) {
	if condition {
		log.Fatal().Msg(message)
	}
}
