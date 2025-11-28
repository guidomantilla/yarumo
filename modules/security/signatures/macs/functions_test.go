package macs

import (
    "crypto/hmac"
    "crypto/sha256"
    "crypto/sha512"
    "testing"

    "github.com/guidomantilla/yarumo/common/types"
    "golang.org/x/crypto/blake2b"
    "golang.org/x/crypto/sha3"
)

func TestHMAC_SHA256(t *testing.T) {
    // 32-byte key for 256-bit HMAC
    key := bytesRepeat('k', 32)
    msg := []byte("abc")

    hm := hmac.New(sha256.New, key)
    hm.Write(msg)
    want := hm.Sum(nil)

    got, err := HMAC_SHA256(key, msg)
    if err != nil {
        t.Fatalf("HMAC_SHA256 error: %v", err)
    }
    if !hmac.Equal(got, want) {
        t.Fatalf("HMAC_SHA256 mismatch")
    }
}

func TestHMAC_SHA3_256(t *testing.T) {
    key := bytesRepeat('k', 32)
    msg := []byte("abc")

    hm := hmac.New(sha3.New256, key)
    hm.Write(msg)
    want := hm.Sum(nil)

    got, err := HMAC_SHA3_256(key, msg)
    if err != nil {
        t.Fatalf("HMAC_SHA3_256 error: %v", err)
    }
    if !hmac.Equal(got, want) {
        t.Fatalf("HMAC_SHA3_256 mismatch")
    }
}

func TestBLAKE2b_256_MAC(t *testing.T) {
    key := bytesRepeat('k', 32)
    msg := []byte("abc")

    d, err := blake2b.New256(key)
    if err != nil {
        t.Fatalf("init blake2b-256: %v", err)
    }
    d.Write(msg)
    want := d.Sum(nil)

    got, err := BLAKE2b_256_MAC(key, msg)
    if err != nil {
        t.Fatalf("BLAKE2b_256_MAC error: %v", err)
    }
    if !hmac.Equal(got, want) {
        t.Fatalf("BLAKE2b_256_MAC mismatch")
    }
}

func TestHMAC_SHA512(t *testing.T) {
    // 64-byte key for 512-bit HMAC
    key := bytesRepeat('k', 64)
    msg := []byte("abc")

    hm := hmac.New(sha512.New, key)
    hm.Write(msg)
    want := hm.Sum(nil)

    got, err := HMAC_SHA512(key, msg)
    if err != nil {
        t.Fatalf("HMAC_SHA512 error: %v", err)
    }
    if !hmac.Equal(got, want) {
        t.Fatalf("HMAC_SHA512 mismatch")
    }
}

func TestHMAC_SHA3_512(t *testing.T) {
    key := bytesRepeat('k', 64)
    msg := []byte("abc")

    hm := hmac.New(sha3.New512, key)
    hm.Write(msg)
    want := hm.Sum(nil)

    got, err := HMAC_SHA3_512(key, msg)
    if err != nil {
        t.Fatalf("HMAC_SHA3_512 error: %v", err)
    }
    if !hmac.Equal(got, want) {
        t.Fatalf("HMAC_SHA3_512 mismatch")
    }
}

func TestBLAKE2b_512(t *testing.T) {
    key := bytesRepeat('k', 64)
    msg := []byte("abc")

    d, err := blake2b.New512(key)
    if err != nil {
        t.Fatalf("init blake2b-512: %v", err)
    }
    d.Write(msg)
    want := d.Sum(nil)

    got, err := BLAKE2b_512_MAC(key, msg)
    if err != nil {
        t.Fatalf("BLAKE2b_512_MAC error: %v", err)
    }
    if !hmac.Equal(got, want) {
        t.Fatalf("BLAKE2b_512_MAC mismatch")
    }
}

// Edge cases to reach 100% coverage
func TestEdgeCases(t *testing.T) {
    msg := []byte("data")

    t.Run("HMAC_SHA256 empty key", func(t *testing.T) {
        if got, err := HMAC_SHA256(nil, msg); got != nil || err != ErrKeySizeInvalid {
            t.Fatalf("expected ErrKeySizeInvalid and nil bytes, got bytes=%v err=%v", got, err)
        }
    })
    t.Run("HMAC_SHA256 empty data", func(t *testing.T) {
        if got, err := HMAC_SHA256(bytesRepeat('k', 32), nil); got != nil || err != ErrDataEmpty {
            t.Fatalf("expected ErrDataEmpty and nil bytes, got bytes=%v err=%v", got, err)
        }
    })
    t.Run("HMAC_SHA256 wrong key size", func(t *testing.T) {
        if got, err := HMAC_SHA256([]byte("short"), msg); got != nil || err != ErrKeySizeInvalid {
            t.Fatalf("expected ErrKeySizeInvalid for wrong key size, got bytes=%v err=%v", got, err)
        }
    })

    t.Run("HMAC_SHA3_256 empty key", func(t *testing.T) {
        if got, err := HMAC_SHA3_256(nil, msg); got != nil || err != ErrKeySizeInvalid {
            t.Fatalf("expected ErrKeySizeInvalid and nil bytes, got bytes=%v err=%v", got, err)
        }
    })
    t.Run("HMAC_SHA3_256 empty data", func(t *testing.T) {
        if got, err := HMAC_SHA3_256(bytesRepeat('k', 32), nil); got != nil || err != ErrDataEmpty {
            t.Fatalf("expected ErrDataEmpty and nil bytes, got bytes=%v err=%v", got, err)
        }
    })
    t.Run("HMAC_SHA3_256 wrong key size", func(t *testing.T) {
        if got, err := HMAC_SHA3_256([]byte("short"), msg); got != nil || err != ErrKeySizeInvalid {
            t.Fatalf("expected ErrKeySizeInvalid for wrong key size, got bytes=%v err=%v", got, err)
        }
    })

    t.Run("HMAC_SHA512 empty key", func(t *testing.T) {
        if got, err := HMAC_SHA512(nil, msg); got != nil || err != ErrKeySizeInvalid {
            t.Fatalf("expected ErrKeySizeInvalid and nil bytes, got bytes=%v err=%v", got, err)
        }
    })
    t.Run("HMAC_SHA512 empty data", func(t *testing.T) {
        if got, err := HMAC_SHA512(bytesRepeat('k', 64), nil); got != nil || err != ErrDataEmpty {
            t.Fatalf("expected ErrDataEmpty and nil bytes, got bytes=%v err=%v", got, err)
        }
    })
    t.Run("HMAC_SHA512 wrong key size", func(t *testing.T) {
        if got, err := HMAC_SHA512([]byte("short"), msg); got != nil || err != ErrKeySizeInvalid {
            t.Fatalf("expected ErrKeySizeInvalid for wrong key size, got bytes=%v err=%v", got, err)
        }
    })

    t.Run("HMAC_SHA3_512 empty key", func(t *testing.T) {
        if got, err := HMAC_SHA3_512(nil, msg); got != nil || err != ErrKeySizeInvalid {
            t.Fatalf("expected ErrKeySizeInvalid and nil bytes, got bytes=%v err=%v", got, err)
        }
    })
    t.Run("HMAC_SHA3_512 empty data", func(t *testing.T) {
        if got, err := HMAC_SHA3_512(bytesRepeat('k', 64), nil); got != nil || err != ErrDataEmpty {
            t.Fatalf("expected ErrDataEmpty and nil bytes, got bytes=%v err=%v", got, err)
        }
    })
    t.Run("HMAC_SHA3_512 wrong key size", func(t *testing.T) {
        if got, err := HMAC_SHA3_512([]byte("short"), msg); got != nil || err != ErrKeySizeInvalid {
            t.Fatalf("expected ErrKeySizeInvalid for wrong key size, got bytes=%v err=%v", got, err)
        }
    })

    t.Run("BLAKE2b_256 empty key", func(t *testing.T) {
        if got, err := BLAKE2b_256_MAC(nil, msg); got != nil || err != ErrKeySizeInvalid {
            t.Fatalf("expected ErrKeySizeInvalid and nil bytes, got bytes=%v err=%v", got, err)
        }
    })
    t.Run("BLAKE2b_256 empty data", func(t *testing.T) {
        if got, err := BLAKE2b_256_MAC(bytesRepeat('k', 32), nil); got != nil || err != ErrDataEmpty {
            t.Fatalf("expected ErrDataEmpty and nil bytes, got bytes=%v err=%v", got, err)
        }
    })
    t.Run("BLAKE2b_256 wrong key size", func(t *testing.T) {
        if got, err := BLAKE2b_256_MAC([]byte("short"), msg); got != nil || err != ErrKeySizeInvalid {
            t.Fatalf("expected ErrKeySizeInvalid for wrong key size, got bytes=%v err=%v", got, err)
        }
    })

    t.Run("BLAKE2b_512 empty key", func(t *testing.T) {
        if got, err := BLAKE2b_512_MAC(nil, msg); got != nil || err != ErrKeySizeInvalid {
            t.Fatalf("expected ErrKeySizeInvalid and nil bytes, got bytes=%v err=%v", got, err)
        }
    })
    t.Run("BLAKE2b_512 empty data", func(t *testing.T) {
        if got, err := BLAKE2b_512_MAC(bytesRepeat('k', 64), nil); got != nil || err != ErrDataEmpty {
            t.Fatalf("expected ErrDataEmpty and nil bytes, got bytes=%v err=%v", got, err)
        }
    })
    t.Run("BLAKE2b_512 wrong key size", func(t *testing.T) {
        if got, err := BLAKE2b_512_MAC([]byte("short"), msg); got != nil || err != ErrKeySizeInvalid {
            t.Fatalf("expected ErrKeySizeInvalid for wrong key size, got bytes=%v err=%v", got, err)
        }
    })
}

// helpers
// bytesRepeat returns a byte slice of length n filled with the given byte.
func bytesRepeat(b byte, n int) []byte {
    out := make([]byte, n)
    for i := range out {
        out[i] = b
    }
    return out
}

func TestEqualHelpers(t *testing.T) {
    a := []byte("abc")
    b := []byte("abc")
    c := []byte("xyz")

    if !Equal(a, b) {
        t.Fatalf("Equal should return true for same contents")
    }
    if NotEqual(a, b) {
        t.Fatalf("NotEqual should return false for same contents")
    }
    if Equal(a, c) {
        t.Fatalf("Equal should return false for different contents")
    }
    if !NotEqual(a, c) {
        t.Fatalf("NotEqual should return true for different contents")
    }
}

func TestRegistry_Get_Register_Supported(t *testing.T) {
    // Existing algorithm should be retrievable
    alg, err := Get(HS_256)
    if err != nil {
        t.Fatalf("Get existing algorithm failed: %v", err)
    }
    if alg.KeySize != 32 || alg.Alias != "HS_256" {
        t.Fatalf("unexpected algorithm metadata: %+v", alg)
    }
    // Use the function through the registry
    out, err := alg.Fn(bytesRepeat('k', alg.KeySize), []byte("abc"))
    if err != nil || len(out) == 0 {
        t.Fatalf("algorithm Fn failed: out=%v err=%v", out, err)
    }

    // Unknown algorithm should return error
    if _, err := Get(Name("UNKNOWN")); err == nil {
        t.Fatalf("expected error for unknown algorithm")
    }

    // Register a new algorithm and retrieve it
    customName := Name("CUSTOM-TEST")
    called := false
    Register(Algorithm{
        Name:    customName,
        Alias:   "CSTM",
        KeySize: 32,
        Fn: func(key types.Bytes, data types.Bytes) (types.Bytes, error) {
            called = true
            return append([]byte("ok:"), data...), nil
        },
    })
    alg2, err := Get(customName)
    if err != nil {
        t.Fatalf("Get custom algorithm failed: %v", err)
    }
    out2, err := alg2.Fn(bytesRepeat('k', 32), []byte("zzz"))
    if err != nil || string(out2) != "ok:zzz" || !called {
        t.Fatalf("custom Fn unexpected: out=%s err=%v called=%v", string(out2), err, called)
    }

    // Supported should include at least the built-ins and our custom
    list := Supported()
    if len(list) < 7 { // 6 built-ins + 1 custom
        t.Fatalf("Supported list too short: %d", len(list))
    }
    // ensure our custom is present
    found := false
    for _, a := range list {
        if a.Name == customName && a.Alias == "CSTM" {
            found = true
            break
        }
    }
    if !found {
        t.Fatalf("custom algorithm not found in Supported list")
    }
}
