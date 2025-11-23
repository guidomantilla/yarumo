package hashes

import (
	"bytes"
	"errors"
	"testing"
)

func TestGetByName_Supported(t *testing.T) {
	tests := []struct {
		name string
		want HashFn
	}{
		{name: Sha256, want: SHA256},
		{name: Sha3_256, want: SHA3_256},
		{name: Sha512, want: SHA512},
		{name: Sha3_512, want: SHA3_512},
		{name: Blake2b_512, want: BLAKE2b_512},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetByName(tt.name)
			if err != nil {
				t.Fatalf("GetByName(%s) unexpected error: %v", tt.name, err)
			}
			if got == nil {
				t.Fatalf("GetByName(%s) returned nil function", tt.name)
			}
			// validate mapping by comparing output of returned function with the
			// known function for a fixed input
			input := []byte("abc")
			outGot := got(input)
			outWant := tt.want(input)
			if len(outGot) == 0 || len(outWant) == 0 {
				t.Fatalf("hash function for %s returned empty output", tt.name)
			}
			if !bytes.Equal(outGot, outWant) {
				t.Fatalf("hash output for %s mismatch", tt.name)
			}
		})
	}
}

func TestGetByName_Unsupported(t *testing.T) {
	unknown := "UNKNOWN_HASH"
	fn, err := GetByName(unknown)
	if fn != nil {
		t.Fatalf("expected nil function for unknown name, got %v", fn)
	}
	if err == nil {
		t.Fatalf("expected error for unknown name")
	}

	var he *Error
	if !errors.As(err, &he) || he == nil {
		t.Fatalf("error is not *Error: %T", err)
	}
	if he.Type != HashNotFound {
		t.Fatalf("error type = %q, want %q", he.Type, HashNotFound)
	}
	if he.Err == nil || he.Err.Error() == "" {
		t.Fatalf("inner error should be set")
	}
}
