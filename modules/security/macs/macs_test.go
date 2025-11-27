package macs

import (
	"bytes"
	"errors"
	"testing"
)

func TestGetByName_Supported(t *testing.T) {
	tests := []struct {
		name string
		want MacFn
	}{
		{name: HmacSha256, want: HMAC_SHA256},
		{name: HmacSha3_256, want: HMAC_SHA3_256},
		{name: Blake2b_256Mac, want: BLAKE2b_256_MAC},
		{name: HmacSha512, want: HMAC_SHA512},
		{name: HmacSha3_512, want: HMAC_SHA3_512},
		{name: Blake2b_512Mac, want: BLAKE2b_512_MAC},
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
			key := []byte("key")
			msg := []byte("abc")
			outGot := got(key, msg)
			outWant := tt.want(key, msg)
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
	unknown := "UNKNOWN_MAC"
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
	if he.Type != MacNotFound {
		t.Fatalf("error type = %q, want %q", he.Type, MacNotFound)
	}
	if he.Err == nil || he.Err.Error() == "" {
		t.Fatalf("inner error should be set")
	}
}
