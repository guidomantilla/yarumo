package assert

import (
    "bytes"
    "strings"
    "testing"

    "github.com/rs/zerolog"
)

type logSink struct {
    buffer bytes.Buffer
}

func (l *logSink) Write(p []byte) (n int, err error) {
	return l.buffer.Write(p)
}

func (l *logSink) Reset() {
	l.buffer.Reset()
}

func (l *logSink) Bytes() []byte {
	return l.buffer.Bytes()
}

func (l *logSink) String() string {
    return l.buffer.String()
}

// Close is invoked by zerolog when logging at Fatal level to flush the writer
// before calling os.Exit(1). We make it panic so that tests can recover and
// continue execution without exiting the process.
func (l *logSink) Close() error {
    panic("no-exit")
}

func init() {
    testSink = &logSink{}
    logger = zerolog.New(testSink).With().Timestamp().Logger()
}

// testSink holds the logger output for assertions
var testSink *logSink

// helper to check that provided text is present in the sink
func mustContain(t *testing.T, want string) {
    t.Helper()
    got := testSink.String()
    if !strings.Contains(got, want) {
        t.Fatalf("expected log to contain %q, got: %s", want, got)
    }
}

// helper to ensure that no logs were written
func mustBeEmpty(t *testing.T) {
    t.Helper()
    if s := testSink.String(); s != "" {
        t.Fatalf("expected no logs, got: %s", s)
    }
}

func TestAssertions_LogBehavior(t *testing.T) {
    // NotEmpty
    testSink.Reset()
    NotEmpty("value", "should not log")
    mustBeEmpty(t)

    testSink.Reset()
    func() {
        defer func() { _ = recover() }()
        NotEmpty("", "empty error")
    }()
    mustContain(t, "empty error")

    // NotNil
    testSink.Reset()
    NotNil(struct{}{}, "should not log")
    mustBeEmpty(t)

    testSink.Reset()
    func() {
        defer func() { _ = recover() }()
        NotNil(nil, "nil error")
    }()
    mustContain(t, "nil error")

    // Equal
    testSink.Reset()
    Equal(10, 10, "should not log")
    mustBeEmpty(t)

    testSink.Reset()
    func() {
        defer func() { _ = recover() }()
        Equal(10, 11, "equal error")
    }()
    mustContain(t, "equal error")

    // NotEqual
    testSink.Reset()
    NotEqual(10, 11, "should not log")
    mustBeEmpty(t)

    testSink.Reset()
    func() {
        defer func() { _ = recover() }()
        NotEqual(10, 10, "not equal error")
    }()
    mustContain(t, "not equal error")

    // True
    testSink.Reset()
    True(true, "should not log")
    mustBeEmpty(t)

    testSink.Reset()
    func() {
        defer func() { _ = recover() }()
        True(false, "true error")
    }()
    mustContain(t, "true error")

    // False
    testSink.Reset()
    False(false, "should not log")
    mustBeEmpty(t)

    testSink.Reset()
    func() {
        defer func() { _ = recover() }()
        False(true, "false error")
    }()
    mustContain(t, "false error")
}
