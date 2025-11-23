package hashes

import (
    "encoding/hex"
    "testing"
)

// helper to decode hex strings to bytes and fail test on error
func mustHex(t *testing.T, s string) []byte {
    t.Helper()
    b, err := hex.DecodeString(s)
    if err != nil {
        t.Fatalf("failed to decode hex: %v", err)
    }
    return b
}

func TestSHA256(t *testing.T) {
    // empty input returns nil
    if got := SHA256(nil); got != nil {
        t.Fatalf("SHA256(nil) expected nil, got %v", got)
    }
    if got := SHA256([]byte{}); got != nil {
        t.Fatalf("SHA256(empty) expected nil, got %v", got)
    }

    // known test vector for "abc"
    want := mustHex(t, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad")
    got := SHA256([]byte("abc"))
    if got == nil || len(got) != 32 {
        t.Fatalf("SHA256 length mismatch: got len=%d", len(got))
    }
    if hex.EncodeToString(got) != hex.EncodeToString(want) {
        t.Fatalf("SHA256 mismatch: got %s want %s", hex.EncodeToString(got), hex.EncodeToString(want))
    }
}

func TestSHA3_256(t *testing.T) {
    if got := SHA3_256(nil); got != nil {
        t.Fatalf("SHA3_256(nil) expected nil, got %v", got)
    }
    if got := SHA3_256([]byte{}); got != nil {
        t.Fatalf("SHA3_256(empty) expected nil, got %v", got)
    }

    want := mustHex(t, "3a985da74fe225b2045c172d6bd390bd855f086e3e9d525b46bfe24511431532")
    got := SHA3_256([]byte("abc"))
    if got == nil || len(got) != 32 {
        t.Fatalf("SHA3_256 length mismatch: got len=%d", len(got))
    }
    if hex.EncodeToString(got) != hex.EncodeToString(want) {
        t.Fatalf("SHA3_256 mismatch: got %s want %s", hex.EncodeToString(got), hex.EncodeToString(want))
    }
}

func TestSHA512(t *testing.T) {
    if got := SHA512(nil); got != nil {
        t.Fatalf("SHA512(nil) expected nil, got %v", got)
    }
    if got := SHA512([]byte{}); got != nil {
        t.Fatalf("SHA512(empty) expected nil, got %v", got)
    }

    want := mustHex(t, "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f")
    got := SHA512([]byte("abc"))
    if got == nil || len(got) != 64 {
        t.Fatalf("SHA512 length mismatch: got len=%d", len(got))
    }
    if hex.EncodeToString(got) != hex.EncodeToString(want) {
        t.Fatalf("SHA512 mismatch: got %s want %s", hex.EncodeToString(got), hex.EncodeToString(want))
    }
}

func TestSHA3_512(t *testing.T) {
    if got := SHA3_512(nil); got != nil {
        t.Fatalf("SHA3_512(nil) expected nil, got %v", got)
    }
    if got := SHA3_512([]byte{}); got != nil {
        t.Fatalf("SHA3_512(empty) expected nil, got %v", got)
    }

    want := mustHex(t, "b751850b1a57168a5693cd924b6b096e08f621827444f70d884f5d0240d2712e10e116e9192af3c91a7ec57647e3934057340b4cf408d5a56592f8274eec53f0")
    got := SHA3_512([]byte("abc"))
    if got == nil || len(got) != 64 {
        t.Fatalf("SHA3_512 length mismatch: got len=%d", len(got))
    }
    if hex.EncodeToString(got) != hex.EncodeToString(want) {
        t.Fatalf("SHA3_512 mismatch: got %s want %s", hex.EncodeToString(got), hex.EncodeToString(want))
    }
}

func TestBLAKE2b_512(t *testing.T) {
    if got := BLAKE2b_512(nil); got != nil {
        t.Fatalf("BLAKE2b_512(nil) expected nil, got %v", got)
    }
    if got := BLAKE2b_512([]byte{}); got != nil {
        t.Fatalf("BLAKE2b_512(empty) expected nil, got %v", got)
    }

    want := mustHex(t, "ba80a53f981c4d0d6a2797b69f12f6e94c212f14685ac4b74b12bb6fdbffa2d17d87c5392aab792dc252d5de4533cc9518d38aa8dbf1925ab92386edd4009923")
    got := BLAKE2b_512([]byte("abc"))
    if got == nil || len(got) != 64 {
        t.Fatalf("BLAKE2b_512 length mismatch: got len=%d", len(got))
    }
    if hex.EncodeToString(got) != hex.EncodeToString(want) {
        t.Fatalf("BLAKE2b_512 mismatch: got %s want %s", hex.EncodeToString(got), hex.EncodeToString(want))
    }
}
