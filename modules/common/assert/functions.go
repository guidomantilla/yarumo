package assert

import (
	"github.com/rs/zerolog/log"

	"github.com/guidomantilla/yarumo/modules/common/utils"
)

// NotEmpty checks if the object is not empty and logs a fatal error if it is.
func NotEmpty(object any, message string) {
	if utils.Empty(object) {
		log.Fatal().Msg(message)
	}
}

// NotNil checks if the object is not nil and logs a fatal error if it is.
func NotNil(object any, message string) {
	if utils.Nil(object) {
		log.Fatal().Msg(message)
	}
}

// Equal checks if two values are equal and logs a fatal error if they are not.
func Equal(val1 any, val2 any, message string) {
	if utils.NotEqual(val1, val2) {
		log.Fatal().Msg(message)
	}
}

// NotEqual checks if two values are not equal and logs a fatal error if they are.
func NotEqual(val1 any, val2 any, message string) {
	if utils.Equal(val1, val2) {
		log.Fatal().Msg(message)
	}
}

// True checks if a condition is true and logs a fatal error if it is not.
func True(condition bool, message string) {
	if !condition {
		log.Fatal().Msg(message)
	}
}

// False checks if a condition is false and logs a fatal error if it is not.
func False(condition bool, message string) {
	if condition {
		log.Fatal().Msg(message)
	}
}
