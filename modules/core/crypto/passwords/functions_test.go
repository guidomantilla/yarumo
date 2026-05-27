package passwords

import (
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestArgon2_Encode_Verify_UpgradeNeeded(t *testing.T) {
	t.Parallel()

	t.Run("encode produces prefixed output", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2id.Encode("argon2-password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.HasPrefix(encoded, Argon2idPrefixKey) {
			t.Fatalf("expected prefix %q, got %q", Argon2idPrefixKey, encoded)
		}
	})

	t.Run("verify matches correct password", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2id.Encode("argon2-verify")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := Argon2id.Verify(encoded, "argon2-verify")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected password to match")
		}
	})

	t.Run("verify rejects wrong password", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2id.Encode("argon2-correct")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := Argon2id.Verify(encoded, "argon2-wrong")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatal("expected password not to match")
		}
	})

	t.Run("verify returns error for wrong prefix", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2id.Verify("{bcrypt}$something", "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("upgrade needed returns false for same params", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2id.Encode("argon2-upgrade")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		needed, err := Argon2id.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if needed {
			t.Fatal("expected no upgrade needed")
		}
	})

	t.Run("upgrade needed returns true for higher iterations", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2id.Encode("argon2-upgrade-iter")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongIter", Argon2idPrefixKey,
			WithArgon2Params(Argon2Iterations+1, Argon2Memory, Argon2Threads, Argon2SaltLength, Argon2KeyLength))

		needed, err := stronger.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !needed {
			t.Fatal("expected upgrade needed for stronger iterations")
		}
	})

	t.Run("upgrade needed returns true for higher memory", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2id.Encode("argon2-upgrade-mem")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongMem", Argon2idPrefixKey,
			WithArgon2Params(Argon2Iterations, Argon2Memory+1024, Argon2Threads, Argon2SaltLength, Argon2KeyLength))

		needed, err := stronger.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !needed {
			t.Fatal("expected upgrade needed for stronger memory")
		}
	})

	t.Run("upgrade needed returns true for more threads", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2id.Encode("argon2-upgrade-threads")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongThreads", Argon2idPrefixKey,
			WithArgon2Params(Argon2Iterations, Argon2Memory, Argon2Threads+1, Argon2SaltLength, Argon2KeyLength))

		needed, err := stronger.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !needed {
			t.Fatal("expected upgrade needed for more threads")
		}
	})

	t.Run("upgrade needed returns true for longer salt", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2id.Encode("argon2-upgrade-salt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongSalt", Argon2idPrefixKey,
			WithArgon2Params(Argon2Iterations, Argon2Memory, Argon2Threads, Argon2SaltLength+8, Argon2KeyLength))

		needed, err := stronger.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !needed {
			t.Fatal("expected upgrade needed for longer salt")
		}
	})

	t.Run("upgrade needed returns true for longer key", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2id.Encode("argon2-upgrade-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongKey", Argon2idPrefixKey,
			WithArgon2Params(Argon2Iterations, Argon2Memory, Argon2Threads, Argon2SaltLength, Argon2KeyLength+16))

		needed, err := stronger.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !needed {
			t.Fatal("expected upgrade needed for longer key")
		}
	})

	t.Run("upgrade needed returns error for wrong prefix", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2id.UpgradeNeeded("{bcrypt}$something")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for wrong number of fields", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2id.Verify("{argon2}$not$enough", "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for non-numeric version", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2id.Verify("{argon2}$abc$1$1$1$"+validB64()+"$"+validB64(), "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for non-numeric iterations", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2id.Verify("{argon2}$19$abc$1$1$"+validB64()+"$"+validB64(), "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for non-numeric memory", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2id.Verify("{argon2}$19$1$abc$1$"+validB64()+"$"+validB64(), "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for non-numeric threads", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2id.Verify("{argon2}$19$1$1$abc$"+validB64()+"$"+validB64(), "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for invalid base64 salt", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2id.Verify("{argon2}$19$1$1$1$!!!invalid!!!$"+validB64(), "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for invalid base64 key", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2id.Verify("{argon2}$19$1$1$1$"+validB64()+"$!!!invalid!!!", "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("upgrade decode returns error for malformed password", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2id.UpgradeNeeded("{argon2}$not$enough")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// TestArgon2i_Encode_Verify_UpgradeNeeded exercises the argon2i (side-channel
// resistant) round trip. The predefined Argon2i method calls argon2.Key (not
// argon2.IDKey) and emits the {argon2i} prefix.
func TestArgon2i_Encode_Verify_UpgradeNeeded(t *testing.T) {
	t.Parallel()

	t.Run("encode produces {argon2i} prefixed output", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2i.Encode("argon2i-password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.HasPrefix(encoded, Argon2iPrefixKey) {
			t.Fatalf("expected prefix %q, got %q", Argon2iPrefixKey, encoded)
		}
	})

	t.Run("verify matches correct password", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2i.Encode("argon2i-verify")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := Argon2i.Verify(encoded, "argon2i-verify")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatal("expected password to match")
		}
	})

	t.Run("verify rejects wrong password", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2i.Encode("argon2i-correct")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := Argon2i.Verify(encoded, "argon2i-wrong")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatal("expected password not to match")
		}
	})

	t.Run("argon2i and argon2id produce distinct keys for same inputs", func(t *testing.T) {
		t.Parallel()

		// Cross-variant sanity: argon2.Key (i) and argon2.IDKey (id) differ
		// even when iterations / memory / threads / salt are identical.
		// This guards against a regression where useArgon2i is ignored and
		// both predefined methods silently call the same KDF.
		raw := "cross-variant-check"

		idEncoded, err := Argon2id.Encode(raw)
		if err != nil {
			t.Fatalf("argon2id encode failed: %v", err)
		}
		idDecoded, err := argon2Decode(idEncoded)
		if err != nil {
			t.Fatalf("argon2id decode failed: %v", err)
		}

		// Re-derive an argon2i key with the SAME salt and parameters so the
		// only difference is the KDF variant. The keys must differ.
		iKey := argon2DeriveKey(true, []byte(raw), idDecoded.salt, idDecoded.iterations, idDecoded.memory, idDecoded.threads, len(idDecoded.key))
		if subtleEqual(idDecoded.key, iKey) {
			t.Fatal("expected argon2i and argon2id to derive distinct keys for identical inputs")
		}
	})

	t.Run("upgrade needed returns false for same params", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2i.Encode("argon2i-upgrade")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		needed, err := Argon2i.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if needed {
			t.Fatal("expected no upgrade needed")
		}
	})

	t.Run("verify returns error for wrong prefix", func(t *testing.T) {
		t.Parallel()

		// {argon2id} and {argon2i} are distinct prefixes; cross-feeding
		// must fail at the prefix check.
		_, err := Argon2i.Verify("{argon2id}$something", "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// TestArgon2_LegacyPrefix_Compat pins the YA-0030 dual-match invariant:
// hashes encoded under the pre-rename {argon2} prefix continue to verify
// against the renamed Argon2id method, and ByPrefix routes the legacy
// prefix to Argon2id.
func TestArgon2_LegacyPrefix_Compat(t *testing.T) {
	t.Parallel()

	t.Run("hash with legacy {argon2} prefix verifies under Argon2id", func(t *testing.T) {
		t.Parallel()

		// Build a Method that emits the legacy {argon2} prefix while using
		// the same argon2id KDF the old code used. This faithfully mimics a
		// stored hash created before YA-0030.
		legacy := NewMethod("LegacyArgon2Compat", Argon2PrefixKey,
			WithArgon2Params(Argon2Iterations, Argon2Memory, Argon2Threads, Argon2SaltLength, Argon2KeyLength))

		encoded, err := legacy.Encode("legacy-argon2-password")
		if err != nil {
			t.Fatalf("unexpected encode error: %v", err)
		}
		if !strings.HasPrefix(encoded, Argon2PrefixKey) {
			t.Fatalf("expected legacy prefix %q, got %q", Argon2PrefixKey, encoded)
		}

		ok, err := Argon2id.Verify(encoded, "legacy-argon2-password")
		if err != nil {
			t.Fatalf("unexpected verify error: %v", err)
		}
		if !ok {
			t.Fatal("expected legacy {argon2} hash to verify under Argon2id")
		}
	})

	t.Run("upgrade-needed accepts legacy {argon2} prefix under Argon2id", func(t *testing.T) {
		t.Parallel()

		legacy := NewMethod("LegacyArgon2UpgradeCompat", Argon2PrefixKey,
			WithArgon2Params(Argon2Iterations, Argon2Memory, Argon2Threads, Argon2SaltLength, Argon2KeyLength))

		encoded, err := legacy.Encode("legacy-upgrade-check")
		if err != nil {
			t.Fatalf("unexpected encode error: %v", err)
		}

		// Argon2id should accept the legacy prefix and report no upgrade
		// needed (parameters match defaults).
		needed, err := Argon2id.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected upgrade-needed error: %v", err)
		}
		if needed {
			t.Fatal("expected no upgrade needed for legacy hash at current defaults")
		}
	})
}

// subtleEqual is a non-crypto helper used by the cross-variant assertion to
// detect identity collisions between argon2i and argon2id outputs. Crypto
// callers use subtle.ConstantTimeCompare; this test only needs simple
// equality.
func subtleEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestPbkdf2_Encode_Verify_UpgradeNeeded(t *testing.T) {
	t.Parallel()

	t.Run("encode produces prefixed output", func(t *testing.T) {
		t.Parallel()

		encoded, err := Pbkdf2.Encode("pbkdf2-password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.HasPrefix(encoded, Pbkdf2PrefixKey) {
			t.Fatalf("expected prefix %q, got %q", Pbkdf2PrefixKey, encoded)
		}
	})

	t.Run("verify matches correct password", func(t *testing.T) {
		t.Parallel()

		encoded, err := Pbkdf2.Encode("pbkdf2-verify")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := Pbkdf2.Verify(encoded, "pbkdf2-verify")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected password to match")
		}
	})

	t.Run("verify rejects wrong password", func(t *testing.T) {
		t.Parallel()

		encoded, err := Pbkdf2.Encode("pbkdf2-correct")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := Pbkdf2.Verify(encoded, "pbkdf2-wrong")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatal("expected password not to match")
		}
	})

	t.Run("verify returns error for wrong prefix", func(t *testing.T) {
		t.Parallel()

		_, err := Pbkdf2.Verify("{bcrypt}$something", "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("upgrade needed returns false for same params", func(t *testing.T) {
		t.Parallel()

		encoded, err := Pbkdf2.Encode("pbkdf2-upgrade")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		needed, err := Pbkdf2.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if needed {
			t.Fatal("expected no upgrade needed")
		}
	})

	t.Run("upgrade needed returns true for higher iterations", func(t *testing.T) {
		t.Parallel()

		encoded, err := Pbkdf2.Encode("pbkdf2-upgrade-iter")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongPbkdf2", Pbkdf2PrefixKey,
			WithPbkdf2Params(Pbkdf2Iterations+1, Pbkdf2SaltLength, Pbkdf2KeyLength, defaultHashFunc()))

		needed, err := stronger.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !needed {
			t.Fatal("expected upgrade needed")
		}
	})

	t.Run("upgrade needed returns true for longer salt", func(t *testing.T) {
		t.Parallel()

		encoded, err := Pbkdf2.Encode("pbkdf2-upgrade-salt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongPbkdf2Salt", Pbkdf2PrefixKey,
			WithPbkdf2Params(Pbkdf2Iterations, Pbkdf2SaltLength+32, Pbkdf2KeyLength, defaultHashFunc()))

		needed, err := stronger.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !needed {
			t.Fatal("expected upgrade needed")
		}
	})

	t.Run("upgrade needed returns true for longer key", func(t *testing.T) {
		t.Parallel()

		encoded, err := Pbkdf2.Encode("pbkdf2-upgrade-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongPbkdf2Key", Pbkdf2PrefixKey,
			WithPbkdf2Params(Pbkdf2Iterations, Pbkdf2SaltLength, Pbkdf2KeyLength+16, defaultHashFunc()))

		needed, err := stronger.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !needed {
			t.Fatal("expected upgrade needed")
		}
	})

	t.Run("upgrade needed returns error for wrong prefix", func(t *testing.T) {
		t.Parallel()

		_, err := Pbkdf2.UpgradeNeeded("{bcrypt}$something")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for wrong field count", func(t *testing.T) {
		t.Parallel()

		_, err := Pbkdf2.Verify("{pbkdf2}$too$many$fields$here", "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for non-numeric iterations", func(t *testing.T) {
		t.Parallel()

		_, err := Pbkdf2.Verify("{pbkdf2}$abc$"+validB64()+"$"+validB64(), "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for invalid base64 salt", func(t *testing.T) {
		t.Parallel()

		_, err := Pbkdf2.Verify("{pbkdf2}$600000$!!!invalid!!!$"+validB64(), "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for invalid base64 key", func(t *testing.T) {
		t.Parallel()

		_, err := Pbkdf2.Verify("{pbkdf2}$600000$"+validB64()+"$!!!invalid!!!", "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("upgrade decode returns error for malformed password", func(t *testing.T) {
		t.Parallel()

		_, err := Pbkdf2.UpgradeNeeded("{pbkdf2}$too$many$fields$here")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestScrypt_Encode_Verify_UpgradeNeeded(t *testing.T) {
	t.Parallel()

	t.Run("encode produces prefixed output", func(t *testing.T) {
		t.Parallel()

		encoded, err := Scrypt.Encode("scrypt-password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.HasPrefix(encoded, ScryptPrefixKey) {
			t.Fatalf("expected prefix %q, got %q", ScryptPrefixKey, encoded)
		}
	})

	t.Run("verify matches correct password", func(t *testing.T) {
		t.Parallel()

		encoded, err := Scrypt.Encode("scrypt-verify")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := Scrypt.Verify(encoded, "scrypt-verify")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected password to match")
		}
	})

	t.Run("verify rejects wrong password", func(t *testing.T) {
		t.Parallel()

		encoded, err := Scrypt.Encode("scrypt-correct")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := Scrypt.Verify(encoded, "scrypt-wrong")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatal("expected password not to match")
		}
	})

	t.Run("verify returns error for wrong prefix", func(t *testing.T) {
		t.Parallel()

		_, err := Scrypt.Verify("{bcrypt}$something", "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("upgrade needed returns false for same params", func(t *testing.T) {
		t.Parallel()

		encoded, err := Scrypt.Encode("scrypt-upgrade")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		needed, err := Scrypt.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if needed {
			t.Fatal("expected no upgrade needed")
		}
	})

	t.Run("upgrade needed returns true for higher N", func(t *testing.T) {
		t.Parallel()

		encoded, err := Scrypt.Encode("scrypt-upgrade-n")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongN", ScryptPrefixKey,
			WithScryptParams(ScryptN*2, ScryptR, ScryptP, ScryptSaltLength, ScryptKeyLength))

		needed, err := stronger.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !needed {
			t.Fatal("expected upgrade needed")
		}
	})

	t.Run("upgrade needed returns true for higher R", func(t *testing.T) {
		t.Parallel()

		encoded, err := Scrypt.Encode("scrypt-upgrade-r")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongR", ScryptPrefixKey,
			WithScryptParams(ScryptN, ScryptR+1, ScryptP, ScryptSaltLength, ScryptKeyLength))

		needed, err := stronger.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !needed {
			t.Fatal("expected upgrade needed")
		}
	})

	t.Run("upgrade needed returns true for higher P", func(t *testing.T) {
		t.Parallel()

		encoded, err := Scrypt.Encode("scrypt-upgrade-p")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongP", ScryptPrefixKey,
			WithScryptParams(ScryptN, ScryptR, ScryptP+1, ScryptSaltLength, ScryptKeyLength))

		needed, err := stronger.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !needed {
			t.Fatal("expected upgrade needed")
		}
	})

	t.Run("upgrade needed returns true for longer salt", func(t *testing.T) {
		t.Parallel()

		encoded, err := Scrypt.Encode("scrypt-upgrade-salt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongSalt", ScryptPrefixKey,
			WithScryptParams(ScryptN, ScryptR, ScryptP, ScryptSaltLength+8, ScryptKeyLength))

		needed, err := stronger.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !needed {
			t.Fatal("expected upgrade needed")
		}
	})

	t.Run("upgrade needed returns true for longer key", func(t *testing.T) {
		t.Parallel()

		encoded, err := Scrypt.Encode("scrypt-upgrade-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongKey", ScryptPrefixKey,
			WithScryptParams(ScryptN, ScryptR, ScryptP, ScryptSaltLength, ScryptKeyLength+16))

		needed, err := stronger.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !needed {
			t.Fatal("expected upgrade needed")
		}
	})

	t.Run("upgrade needed returns error for wrong prefix", func(t *testing.T) {
		t.Parallel()

		_, err := Scrypt.UpgradeNeeded("{bcrypt}$something")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for wrong field count", func(t *testing.T) {
		t.Parallel()

		_, err := Scrypt.Verify("{scrypt}$not$enough", "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for non-numeric N", func(t *testing.T) {
		t.Parallel()

		_, err := Scrypt.Verify("{scrypt}$abc$8$1$"+validB64()+"$"+validB64(), "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for non-numeric R", func(t *testing.T) {
		t.Parallel()

		_, err := Scrypt.Verify("{scrypt}$32768$abc$1$"+validB64()+"$"+validB64(), "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for non-numeric P", func(t *testing.T) {
		t.Parallel()

		_, err := Scrypt.Verify("{scrypt}$32768$8$abc$"+validB64()+"$"+validB64(), "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for invalid base64 salt", func(t *testing.T) {
		t.Parallel()

		_, err := Scrypt.Verify("{scrypt}$32768$8$1$!!!invalid!!!$"+validB64(), "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for invalid base64 key", func(t *testing.T) {
		t.Parallel()

		_, err := Scrypt.Verify("{scrypt}$32768$8$1$"+validB64()+"$!!!invalid!!!", "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("upgrade decode returns error for malformed password", func(t *testing.T) {
		t.Parallel()

		_, err := Scrypt.UpgradeNeeded("{scrypt}$not$enough")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestBcrypt_Encode_Verify_UpgradeNeeded(t *testing.T) {
	t.Parallel()

	t.Run("encode with invalid cost returns error", func(t *testing.T) {
		t.Parallel()

		m := &Method{
			name:         "BadCost",
			prefix:       BcryptPrefixKey,
			bcryptParams: &bcryptConfig{cost: bcrypt.MaxCost + 1},
			encodeFn:     encode,
		}

		_, err := m.Encode("password")
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrBcryptCostNotAllowed) {
			t.Fatalf("expected ErrBcryptCostNotAllowed, got %v", err)
		}
	})

	t.Run("upgrade needed returns true for higher cost", func(t *testing.T) {
		t.Parallel()

		encoded, err := Bcrypt.Encode("bcrypt-upgrade")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongBcrypt", BcryptPrefixKey,
			WithBcryptParams(BcryptDefaultCost+1))

		needed, err := stronger.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !needed {
			t.Fatal("expected upgrade needed")
		}
	})

	t.Run("verify returns error for wrong prefix", func(t *testing.T) {
		t.Parallel()

		_, err := Bcrypt.Verify("{argon2}$something", "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("upgrade needed returns error for wrong prefix", func(t *testing.T) {
		t.Parallel()

		_, err := Bcrypt.UpgradeNeeded("{argon2}$something")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("upgrade needed returns error for malformed bcrypt hash", func(t *testing.T) {
		t.Parallel()

		_, err := Bcrypt.UpgradeNeeded("{bcrypt}not-a-valid-bcrypt-hash")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func Test_encode_returns_error_for_empty_password(t *testing.T) {
	t.Parallel()

	_, err := Bcrypt.Encode("")
	if err == nil {
		t.Fatal("expected error")
	}
}

func Test_encode_returns_error_for_no_config(t *testing.T) {
	t.Parallel()

	m := &Method{
		name:     "empty",
		prefix:   "{empty}",
		encodeFn: encode,
	}

	_, err := m.Encode("password")
	if err == nil {
		t.Fatal("expected error")
	}
}

func Test_verify_returns_error_for_empty_raw_password(t *testing.T) {
	t.Parallel()

	_, err := Bcrypt.Verify("{bcrypt}encoded", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func Test_verify_returns_error_for_empty_encoded_password(t *testing.T) {
	t.Parallel()

	_, err := Bcrypt.Verify("", "password")
	if err == nil {
		t.Fatal("expected error")
	}
}

func Test_verify_returns_error_for_no_config(t *testing.T) {
	t.Parallel()

	m := &Method{
		name:     "empty",
		prefix:   "{empty}",
		verifyFn: verify,
	}

	_, err := m.Verify("{empty}encoded", "password")
	if err == nil {
		t.Fatal("expected error")
	}
}

func Test_upgradeNeeded_returns_error_for_empty_encoded(t *testing.T) {
	t.Parallel()

	_, err := Bcrypt.UpgradeNeeded("")
	if err == nil {
		t.Fatal("expected error")
	}
}

func Test_upgradeNeeded_returns_error_for_no_config(t *testing.T) {
	t.Parallel()

	m := &Method{
		name:            "empty",
		prefix:          "{empty}",
		upgradeNeededFn: upgradeNeeded,
	}

	_, err := m.UpgradeNeeded("{empty}encoded")
	if err == nil {
		t.Fatal("expected error")
	}
}

// Test_saltEntropy_viaPublicAPI verifies that encoded passwords embed a salt
// section of the expected length and that two consecutive encodes of the same
// raw password produce distinct outputs (i.e. the salt entropy source is
// actually wired up). This replaces the previous unit test against the now-
// removed private generateSalt helper; entropy now flows from common/crypto/random.
func Test_saltEntropy_viaPublicAPI(t *testing.T) {
	t.Parallel()

	t.Run("two encodes produce distinct outputs", func(t *testing.T) {
		t.Parallel()

		first, err := Argon2id.Encode("same-password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		second, err := Argon2id.Encode("same-password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if first == second {
			t.Fatal("expected distinct encoded outputs from fresh salts; got identical")
		}
	})

	t.Run("argon2 embeds salt section of configured length", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2id.Encode("salt-length-check")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		decoded, err := argon2Decode(encoded)
		if err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}

		if len(decoded.salt) != Argon2SaltLength {
			t.Fatalf("expected salt length %d, got %d", Argon2SaltLength, len(decoded.salt))
		}

		allZero := true
		for _, b := range decoded.salt {
			if b != 0 {
				allZero = false
				break
			}
		}
		if allZero {
			t.Fatal("expected non-zero salt bytes")
		}
	})
}

// TestBackwardCompat_OldHashesVerify ensures that hashes encoded under the
// pre-OWASP-2024 defaults (BcryptDefaultCost=10, ScryptN=32768) still verify
// after bumping the package defaults to OWASP-2024 values (cost=12, N=131072).
//
// Both bcrypt and scrypt encode their parameters into the hash string, and
// Method.Verify reads the stored parameters when re-deriving the key — it
// never substitutes the current package defaults. This test pins that
// contract so the YA-0006 default bump cannot lock out existing users.
func TestBackwardCompat_OldHashesVerify(t *testing.T) {
	t.Parallel()

	t.Run("bcrypt hash encoded at old cost 10 still verifies under cost 12 default", func(t *testing.T) {
		t.Parallel()

		// Encode using a Method pinned to the pre-OWASP default cost.
		oldBcrypt := NewMethod("LegacyBcrypt", BcryptPrefixKey, WithBcryptParams(10))
		encoded, err := oldBcrypt.Encode("legacy-bcrypt-password")
		if err != nil {
			t.Fatalf("unexpected error encoding under old cost: %v", err)
		}

		// Verify using the package-default Bcrypt method (which now uses cost 12).
		ok, err := Bcrypt.Verify(encoded, "legacy-bcrypt-password")
		if err != nil {
			t.Fatalf("unexpected error verifying legacy hash: %v", err)
		}
		if !ok {
			t.Fatal("expected legacy bcrypt hash (cost 10) to verify under cost 12 default")
		}

		// Bcrypt should also flag the legacy hash as upgrade-needed.
		needed, err := Bcrypt.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error checking upgrade: %v", err)
		}
		if !needed {
			t.Fatal("expected upgrade needed for legacy bcrypt hash (cost 10 < 12)")
		}
	})

	t.Run("scrypt hash encoded at old N=32768 still verifies under N=131072 default", func(t *testing.T) {
		t.Parallel()

		// Encode using a Method pinned to the pre-OWASP default N. WithScryptParams
		// rejects values below the current ScryptN floor (now 131072), so we build
		// the Method directly with a hand-rolled config to simulate a legacy hash.
		oldScrypt := &Method{
			name:   "LegacyScrypt",
			prefix: ScryptPrefixKey,
			scryptParams: &scryptConfig{
				n:          32768,
				r:          ScryptR,
				p:          ScryptP,
				saltLength: ScryptSaltLength,
				keyLength:  ScryptKeyLength,
			},
			encodeFn:        encode,
			verifyFn:        verify,
			upgradeNeededFn: upgradeNeeded,
		}
		encoded, err := oldScrypt.Encode("legacy-scrypt-password")
		if err != nil {
			t.Fatalf("unexpected error encoding under old N: %v", err)
		}

		// Verify using the package-default Scrypt method (which now uses N=131072).
		ok, err := Scrypt.Verify(encoded, "legacy-scrypt-password")
		if err != nil {
			t.Fatalf("unexpected error verifying legacy hash: %v", err)
		}
		if !ok {
			t.Fatal("expected legacy scrypt hash (N=32768) to verify under N=131072 default")
		}

		// Scrypt should also flag the legacy hash as upgrade-needed.
		needed, err := Scrypt.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error checking upgrade: %v", err)
		}
		if !needed {
			t.Fatal("expected upgrade needed for legacy scrypt hash (N=32768 < 131072)")
		}
	})
}

// --- Helpers ---

func validB64() string {
	return base64.RawStdEncoding.EncodeToString([]byte("validdata12345678"))
}

func defaultHashFunc() HashFunc {
	return sha512.New
}
