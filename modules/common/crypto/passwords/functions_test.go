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

		encoded, err := Argon2.Encode("argon2-password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.HasPrefix(encoded, Argon2PrefixKey) {
			t.Fatalf("expected prefix %q, got %q", Argon2PrefixKey, encoded)
		}
	})

	t.Run("verify matches correct password", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2.Encode("argon2-verify")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := Argon2.Verify(encoded, "argon2-verify")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected password to match")
		}
	})

	t.Run("verify rejects wrong password", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2.Encode("argon2-correct")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := Argon2.Verify(encoded, "argon2-wrong")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatal("expected password not to match")
		}
	})

	t.Run("verify returns error for wrong prefix", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2.Verify("{bcrypt}$something", "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("upgrade needed returns false for same params", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2.Encode("argon2-upgrade")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		needed, err := Argon2.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if needed {
			t.Fatal("expected no upgrade needed")
		}
	})

	t.Run("upgrade needed returns true for higher iterations", func(t *testing.T) {
		t.Parallel()

		encoded, err := Argon2.Encode("argon2-upgrade-iter")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongIter", Argon2PrefixKey,
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

		encoded, err := Argon2.Encode("argon2-upgrade-mem")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongMem", Argon2PrefixKey,
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

		encoded, err := Argon2.Encode("argon2-upgrade-threads")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongThreads", Argon2PrefixKey,
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

		encoded, err := Argon2.Encode("argon2-upgrade-salt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongSalt", Argon2PrefixKey,
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

		encoded, err := Argon2.Encode("argon2-upgrade-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stronger := NewMethod("StrongKey", Argon2PrefixKey,
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

		_, err := Argon2.UpgradeNeeded("{bcrypt}$something")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for wrong number of fields", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2.Verify("{argon2}$not$enough", "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for non-numeric version", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2.Verify("{argon2}$abc$1$1$1$"+validB64()+"$"+validB64(), "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for non-numeric iterations", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2.Verify("{argon2}$19$abc$1$1$"+validB64()+"$"+validB64(), "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for non-numeric memory", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2.Verify("{argon2}$19$1$abc$1$"+validB64()+"$"+validB64(), "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for non-numeric threads", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2.Verify("{argon2}$19$1$1$abc$"+validB64()+"$"+validB64(), "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for invalid base64 salt", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2.Verify("{argon2}$19$1$1$1$!!!invalid!!!$"+validB64(), "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("decode returns error for invalid base64 key", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2.Verify("{argon2}$19$1$1$1$"+validB64()+"$!!!invalid!!!", "password")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("upgrade decode returns error for malformed password", func(t *testing.T) {
		t.Parallel()

		_, err := Argon2.UpgradeNeeded("{argon2}$not$enough")
		if err == nil {
			t.Fatal("expected error")
		}
	})
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

func Test_generateSalt(t *testing.T) {
	t.Parallel()

	t.Run("generates salt of expected size", func(t *testing.T) {
		t.Parallel()

		salt, err := generateSalt(16)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(salt) == 0 {
			t.Fatal("expected non-empty salt")
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
