package uids

import (
	"errors"
	"strings"
	"testing"
)

// helper to snapshot and restore the global methods map for isolation
func snapshotMethods() map[string]UID {
	cp := make(map[string]UID, len(methods))
	for k, v := range methods {
		cp[k] = v
	}
	return cp
}

func restoreMethods(m map[string]UID) {
	// replace the global map with the snapshot
	methods = make(map[string]UID, len(m))
	for k, v := range m {
		methods[k] = v
	}
}

func TestNewUID_Name_Generate(t *testing.T) {
	const expected = "fixed-id"
	fn := func() string { return expected }

	uid := NewUID("TEST", fn)
	if uid == nil {
		t.Fatalf("NewUID returned nil")
	}
	if uid.Name() != "TEST" {
		t.Fatalf("Name() = %q, want %q", uid.Name(), "TEST")
	}
	if got := uid.Generate(); got != expected {
		t.Fatalf("Generate() = %q, want %q", got, expected)
	}
}

func TestFunctionsReturnNonEmpty(t *testing.T) {
	cases := []struct {
		name string
		fn   UIDFn
	}{
		{"UUIDv4", UUIDv4},
		{"NANOID", NANOID},
		{"CUID2", CUID2},
		{"UUIDv7", UUIDv7},
		{"ULID", ULID},
		{"XID", XID},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := tc.fn()
			if id == "" {
				t.Fatalf("%s returned empty string", tc.name)
			}
		})
	}
}

func TestGetSupportedAndRegister(t *testing.T) {
	snap := snapshotMethods()
	defer restoreMethods(snap)

	// Known existing UID
	uid, err := Get("UUIDv4")
	if err != nil {
		t.Fatalf("Get(UUIDv4) error: %v", err)
	}
	if uid == nil || uid.Name() != "UUIDv4" {
		t.Fatalf("Get(UUIDv4) returned wrong UID: %+v", uid)
	}

	// Unknown UID should error
	_, err = Get("DOES_NOT_EXIST")
	if err == nil {
		t.Fatalf("expected error for unknown UID name")
	}
	var e *Error
	if !errors.As(err, &e) {
		t.Fatalf("expected *Error, got %T", err)
	}

	// Register a new UID and retrieve it
	testFn := func() string { return "xyz" }
	testUID := NewUID("TESTID", testFn)
	Register(*testUID)

	got, err := Get("TESTID")
	if err != nil {
		t.Fatalf("Get(TESTID) error: %v", err)
	}
	if got == nil || got.Generate() != "xyz" {
		t.Fatalf("retrieved UID not working, got: %+v, gen: %q", got, got.Generate())
	}

	// Ensure Supported includes the new one
	list := Supported()
	found := false
	for _, u := range list {
		if u.name == "TESTID" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Supported() does not include TESTID")
	}
}

func TestErrAlgorithmNotSupportedFormatting(t *testing.T) {
	err := ErrAlgorithmNotSupported("ABC")
	if err == nil {
		t.Fatalf("expected error")
	}
	// Ensure Error() produces a meaningful string
	s := err.Error()
	if !strings.Contains(s, "uid ") || !strings.Contains(s, "ABC") {
		t.Fatalf("unexpected error string: %q", s)
	}
}
