package aead

import (
	"bytes"
	"errors"
	"io"
	"testing"

	crandom "github.com/guidomantilla/yarumo/common/random"
)

func TestMethod_EncryptStream_DecryptStream_RoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("multi-megabyte buffer round-trips", func(t *testing.T) {
		t.Parallel()

		// 3 MiB of random plaintext exercises many frames.
		plaintext := crandom.Bytes(3 * 1024 * 1024)

		key, err := AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		aad := []byte("multi-mb-stream-test")

		var encrypted bytes.Buffer

		encErr := AES_256_GCM.EncryptStream(key, bytes.NewReader(plaintext), &encrypted, aad)
		if encErr != nil {
			t.Fatalf("unexpected encrypt error: %v", encErr)
		}

		var decrypted bytes.Buffer

		decErr := AES_256_GCM.DecryptStream(key, &encrypted, &decrypted, aad)
		if decErr != nil {
			t.Fatalf("unexpected decrypt error: %v", decErr)
		}

		if !bytes.Equal(decrypted.Bytes(), plaintext) {
			t.Fatalf("plaintext mismatch: got %d bytes, want %d", decrypted.Len(), len(plaintext))
		}
	})

	t.Run("round-trips across all predefined ciphers", func(t *testing.T) {
		t.Parallel()

		methods := []*Method{AES_128_GCM, AES_256_GCM, CHACHA20_POLY1305, XCHACHA20_POLY1305}

		plaintext := bytes.Repeat([]byte("yarumo aead streaming"), 1024)
		aad := []byte("multi-cipher")

		for _, m := range methods {
			t.Run(m.Name(), func(t *testing.T) {
				t.Parallel()

				key, err := m.GenerateKey()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				var encrypted bytes.Buffer

				encErr := m.EncryptStream(key, bytes.NewReader(plaintext), &encrypted, aad)
				if encErr != nil {
					t.Fatalf("unexpected encrypt error: %v", encErr)
				}

				var decrypted bytes.Buffer

				decErr := m.DecryptStream(key, &encrypted, &decrypted, aad)
				if decErr != nil {
					t.Fatalf("unexpected decrypt error: %v", decErr)
				}

				if !bytes.Equal(decrypted.Bytes(), plaintext) {
					t.Fatalf("plaintext mismatch for %s", m.Name())
				}
			})
		}
	})
}

func TestMethod_EncryptStream_FrameBoundaries(t *testing.T) {
	t.Parallel()

	sizes := []struct {
		name string
		size int
	}{
		{"1 byte", 1},
		{"frame size minus one", StreamFrameSize - 1},
		{"exactly one frame", StreamFrameSize},
		{"one frame plus one byte", StreamFrameSize + 1},
		{"two frames", 2 * StreamFrameSize},
		{"two frames plus one byte", 2*StreamFrameSize + 1},
		{"three frames", 3 * StreamFrameSize},
	}

	for _, tc := range sizes {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			plaintext := crandom.Bytes(tc.size)

			key, err := AES_256_GCM.GenerateKey()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var encrypted bytes.Buffer

			encErr := AES_256_GCM.EncryptStream(key, bytes.NewReader(plaintext), &encrypted, nil)
			if encErr != nil {
				t.Fatalf("unexpected encrypt error: %v", encErr)
			}

			var decrypted bytes.Buffer

			decErr := AES_256_GCM.DecryptStream(key, &encrypted, &decrypted, nil)
			if decErr != nil {
				t.Fatalf("unexpected decrypt error: %v", decErr)
			}

			if !bytes.Equal(decrypted.Bytes(), plaintext) {
				t.Fatalf("plaintext mismatch for size %d: got %d bytes", tc.size, decrypted.Len())
			}
		})
	}
}

func TestMethod_EncryptStream_EmptyInput(t *testing.T) {
	t.Parallel()

	t.Run("empty plaintext emits only end-of-stream sentinel", func(t *testing.T) {
		t.Parallel()

		key, err := AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var encrypted bytes.Buffer

		encErr := AES_256_GCM.EncryptStream(key, bytes.NewReader(nil), &encrypted, nil)
		if encErr != nil {
			t.Fatalf("unexpected encrypt error: %v", encErr)
		}

		// Only the 4-byte zero-length sentinel is emitted.
		if encrypted.Len() != 4 {
			t.Fatalf("expected 4 bytes (sentinel only), got %d", encrypted.Len())
		}

		var decrypted bytes.Buffer

		decErr := AES_256_GCM.DecryptStream(key, &encrypted, &decrypted, nil)
		if decErr != nil {
			t.Fatalf("unexpected decrypt error: %v", decErr)
		}

		if decrypted.Len() != 0 {
			t.Fatalf("expected empty plaintext, got %d bytes", decrypted.Len())
		}
	})
}

func TestMethod_DecryptStream_Truncation(t *testing.T) {
	t.Parallel()

	t.Run("missing end-of-stream sentinel returns ErrStreamTruncated", func(t *testing.T) {
		t.Parallel()

		key, err := AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plaintext := bytes.Repeat([]byte("truncate-me"), 1024)

		var encrypted bytes.Buffer

		encErr := AES_256_GCM.EncryptStream(key, bytes.NewReader(plaintext), &encrypted, nil)
		if encErr != nil {
			t.Fatalf("unexpected encrypt error: %v", encErr)
		}

		// Drop the last 4-byte zero-length sentinel.
		truncated := encrypted.Bytes()[:encrypted.Len()-4]

		var decrypted bytes.Buffer

		decErr := AES_256_GCM.DecryptStream(key, bytes.NewReader(truncated), &decrypted, nil)
		if decErr == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(decErr, ErrStreamTruncated) {
			t.Fatalf("expected ErrStreamTruncated, got %v", decErr)
		}
	})

	t.Run("partial frame body returns ErrStreamTruncated", func(t *testing.T) {
		t.Parallel()

		key, err := AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plaintext := bytes.Repeat([]byte("partial-frame"), 1024)

		var encrypted bytes.Buffer

		encErr := AES_256_GCM.EncryptStream(key, bytes.NewReader(plaintext), &encrypted, nil)
		if encErr != nil {
			t.Fatalf("unexpected encrypt error: %v", encErr)
		}

		// Drop the trailing portion mid-frame.
		raw := encrypted.Bytes()
		truncated := raw[:len(raw)/2]

		var decrypted bytes.Buffer

		decErr := AES_256_GCM.DecryptStream(key, bytes.NewReader(truncated), &decrypted, nil)
		if decErr == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(decErr, ErrStreamTruncated) {
			t.Fatalf("expected ErrStreamTruncated, got %v", decErr)
		}
	})
}

func TestMethod_DecryptStream_Tampering(t *testing.T) {
	t.Parallel()

	t.Run("byte flip mid-frame triggers AEAD authentication failure", func(t *testing.T) {
		t.Parallel()

		key, err := AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Use a payload large enough to span multiple frames so the
		// tamper falls inside a known ciphertext region.
		plaintext := crandom.Bytes(3 * StreamFrameSize)

		var encrypted bytes.Buffer

		encErr := AES_256_GCM.EncryptStream(key, bytes.NewReader(plaintext), &encrypted, nil)
		if encErr != nil {
			t.Fatalf("unexpected encrypt error: %v", encErr)
		}

		raw := encrypted.Bytes()
		// Flip a byte well inside the first ciphertext frame
		// (past the 4-byte length prefix and the 12-byte nonce).
		raw[100] ^= 0xff

		var decrypted bytes.Buffer

		decErr := AES_256_GCM.DecryptStream(key, bytes.NewReader(raw), &decrypted, nil)
		if decErr == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(decErr, ErrDecryptFailed) {
			t.Fatalf("expected ErrDecryptFailed, got %v", decErr)
		}
	})

	t.Run("wrong key fails authentication", func(t *testing.T) {
		t.Parallel()

		keyA, err := AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		keyB, err := AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plaintext := bytes.Repeat([]byte("wrong-key"), 4096)

		var encrypted bytes.Buffer

		encErr := AES_256_GCM.EncryptStream(keyA, bytes.NewReader(plaintext), &encrypted, nil)
		if encErr != nil {
			t.Fatalf("unexpected encrypt error: %v", encErr)
		}

		var decrypted bytes.Buffer

		decErr := AES_256_GCM.DecryptStream(keyB, &encrypted, &decrypted, nil)
		if decErr == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(decErr, ErrDecryptFailed) {
			t.Fatalf("expected ErrDecryptFailed, got %v", decErr)
		}
	})

	t.Run("wrong AAD fails authentication", func(t *testing.T) {
		t.Parallel()

		key, err := AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plaintext := bytes.Repeat([]byte("wrong-aad"), 4096)

		var encrypted bytes.Buffer

		encErr := AES_256_GCM.EncryptStream(key, bytes.NewReader(plaintext), &encrypted, []byte("aadA"))
		if encErr != nil {
			t.Fatalf("unexpected encrypt error: %v", encErr)
		}

		var decrypted bytes.Buffer

		decErr := AES_256_GCM.DecryptStream(key, &encrypted, &decrypted, []byte("aadB"))
		if decErr == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(decErr, ErrDecryptFailed) {
			t.Fatalf("expected ErrDecryptFailed, got %v", decErr)
		}
	})

	t.Run("oversize length prefix is rejected", func(t *testing.T) {
		t.Parallel()

		key, err := AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Forge a stream that begins with a length prefix larger than allowed.
		var forged bytes.Buffer
		_, _ = forged.Write([]byte{0xff, 0xff, 0xff, 0xff}) // length = ~4 GiB

		var decrypted bytes.Buffer

		decErr := AES_256_GCM.DecryptStream(key, &forged, &decrypted, nil)
		if decErr == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(decErr, ErrStreamFrameTooLarge) {
			t.Fatalf("expected ErrStreamFrameTooLarge, got %v", decErr)
		}
	})
}

func TestMethod_DecryptStream_FrameReordering(t *testing.T) {
	t.Parallel()

	t.Run("swapping two frames triggers AEAD authentication failure", func(t *testing.T) {
		t.Parallel()

		key, err := AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Two distinct full-frame blocks so we can swap them on the wire.
		plaintext := crandom.Bytes(2 * StreamFrameSize)

		var encrypted bytes.Buffer

		encErr := AES_256_GCM.EncryptStream(key, bytes.NewReader(plaintext), &encrypted, nil)
		if encErr != nil {
			t.Fatalf("unexpected encrypt error: %v", encErr)
		}

		// Parse the two frames from the wire and emit them in reverse order.
		raw := encrypted.Bytes()
		buf := bytes.NewReader(raw)

		f1, readErr := readOneFrame(buf)
		if readErr != nil {
			t.Fatalf("unexpected read error: %v", readErr)
		}

		f2, readErr := readOneFrame(buf)
		if readErr != nil {
			t.Fatalf("unexpected read error: %v", readErr)
		}

		var swapped bytes.Buffer
		writeOneFrame(&swapped, f2)
		writeOneFrame(&swapped, f1)
		_, _ = swapped.Write([]byte{0, 0, 0, 0}) // end-of-stream sentinel

		var decrypted bytes.Buffer

		decErr := AES_256_GCM.DecryptStream(key, &swapped, &decrypted, nil)
		if decErr == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(decErr, ErrDecryptFailed) {
			t.Fatalf("expected ErrDecryptFailed, got %v", decErr)
		}
	})
}

// readOneFrame reads a length-prefixed frame from r and returns the
// ciphertext body (without the prefix). Test helper only.
func readOneFrame(r io.Reader) ([]byte, error) {
	length, err := readFrameLength(r)
	if err != nil {
		return nil, err
	}

	frame := make([]byte, length)

	_, readErr := io.ReadFull(r, frame)
	if readErr != nil {
		return nil, readErr
	}

	return frame, nil
}

// writeOneFrame writes a length-prefixed frame to w. Test helper only.
func writeOneFrame(w io.Writer, frame []byte) {
	_ = writeFrameLength(w, uint32(len(frame))) //nolint:gosec // test helper, bounded input
	_, _ = w.Write(frame)
}
