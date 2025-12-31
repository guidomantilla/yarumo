package log

import (
	"net"
	"testing"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// startTCPServer starts a local TCP server and returns the listener and its address string.
func startTCPServer(t *testing.T) (net.Listener, string) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start tcp server: %v", err)
	}

	addr := ln.Addr().String()

	// Accept a single connection in the background to satisfy Dial and then keep it open until test ends.
	go func() {
		// We don't fail the test from a goroutine; the best effort accepts.
		conn, _ := ln.Accept()
		if conn != nil {
			// Keep it open for the duration of the test; closed by listener Close on cleanup.
			// If the test ends earlier, Close will drop it.
		}
	}()

	return ln, addr
}

func TestConfigure_DefaultInvalidLevel(t *testing.T) {
	// Ensure to clean viper state and restore global level afterward
	viper.Reset()

	prev := zerolog.GlobalLevel()

	t.Cleanup(func() {
		zerolog.SetGlobalLevel(prev)
		viper.Reset()
	})

	// Invalid level should default to info and set the global level accordingly
	viper.Set("LOG_LEVEL", "not-a-level")

	l := Configure("", "")
	_ = l // just ensure it returns a logger

	if got := zerolog.GlobalLevel(); got != zerolog.InfoLevel {
		t.Fatalf("GlobalLevel = %v, want %v (info)", got, zerolog.InfoLevel)
	}
}

func TestConfigure_DebugModeWithNameVersion(t *testing.T) {
	viper.Reset()

	prev := zerolog.GlobalLevel()

	t.Cleanup(func() {
		zerolog.SetGlobalLevel(prev)
		viper.Reset()
	})

	viper.Set("DEBUG_MODE", true)
	viper.Set("LOG_LEVEL", "debug")

	// Non-empty name/version should be attached to context; not directly asserted, but a path is executed.
	logger := Configure("svc", "1.2.3")
	// Use the logger once to ensure it's functional
	logger.Debug().Msg("hello")

	if got := zerolog.GlobalLevel(); got != zerolog.DebugLevel {
		t.Fatalf("GlobalLevel = %v, want %v (debug)", got, zerolog.DebugLevel)
	}
}

func TestConfigure_WithLogstashAddress(t *testing.T) {
	viper.Reset()

	prev := zerolog.GlobalLevel()

	t.Cleanup(func() {
		zerolog.SetGlobalLevel(prev)
		viper.Reset()
	})

	// Bring level to warn before to be sure Configure can change it to info
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	ln, addr := startTCPServer(t)
	t.Cleanup(func() { _ = ln.Close() })

	viper.Set("LOGSTASH_ADDRESS", addr)
	viper.Set("LOG_LEVEL", "info")

	// Should dial successfully and append writer; just ensure it does not panic
	_ = Configure("app", "0.0.1")

	if got := zerolog.GlobalLevel(); got != zerolog.InfoLevel {
		t.Fatalf("GlobalLevel = %v, want %v (info)", got, zerolog.InfoLevel)
	}
}

func TestConfigure_DisabledLevelDoesNotChangeGlobal(t *testing.T) {
	viper.Reset()
	// Set a known non-default level first
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	t.Cleanup(func() {
		zerolog.SetGlobalLevel(zerolog.InfoLevel) // restore typical default
		viper.Reset()
	})

	viper.Set("LOG_LEVEL", "disabled")

	// Call Configure; condition for SetGlobalLevel should be false when level==Disabled
	_ = Configure("", "")

	if got := zerolog.GlobalLevel(); got != zerolog.ErrorLevel {
		t.Fatalf("GlobalLevel changed = %v, want unchanged %v", got, zerolog.ErrorLevel)
	}

	// Also ensure the returned global logger is set (side effect line 57)
	zlog.Logger.Info().Msg("ok")
}
