package aead

import (
	"encoding/binary"
	"errors"
	"io"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
	ctypes "github.com/guidomantilla/yarumo/core/common/types"
)

// StreamFrameSize is the plaintext chunk size used by EncryptStream and
// DecryptStream. 64 KiB is a common tradeoff between per-frame AEAD tag
// overhead (16 bytes per frame for the predefined ciphers) and memory
// footprint. Callers cannot override this constant because the on-wire
// frame length is encoded as uint32 and both sides must agree on the
// segmentation.
const StreamFrameSize = 64 * 1024

// streamFrameLengthSize is the size of the big-endian uint32 length prefix
// that precedes each frame on the wire.
const streamFrameLengthSize = 4

// streamFrameCounterSize is the size of the big-endian uint64 frame counter
// that is bound into every frame's AAD.
const streamFrameCounterSize = 8

// streamCipherMaxFrameSize is the maximum size in bytes a single ciphertext
// frame (nonce || sealed || tag) is permitted to occupy on the wire. It is
// computed from StreamFrameSize plus generous nonce + tag headroom; any
// length-prefix on the wire larger than this is rejected eagerly as a
// truncation or corruption signal before any allocation occurs.
const streamCipherMaxFrameSize = StreamFrameSize + 256

// Sentinel errors for the streaming AEAD API.
//
// These are unwrapped by the wrap-layer in EncryptStream / DecryptStream
// before being surfaced via ErrEncryption / ErrDecryption.
var (
	// ErrStreamFrameTooLarge indicates a length-prefix on the wire exceeded
	// streamCipherMaxFrameSize and was rejected before any allocation.
	ErrStreamFrameTooLarge = errors.New("aead stream frame too large")
	// ErrStreamTruncated indicates the input ended before the zero-length
	// end-of-stream sentinel was seen.
	ErrStreamTruncated = errors.New("aead stream truncated (missing end-of-stream sentinel)")
)

// EncryptStream reads plaintext from src and writes a chunked AEAD ciphertext
// to dst. Each plaintext chunk of at most StreamFrameSize bytes is sealed
// independently with method.Encrypt; the resulting ciphertext (which carries
// its own random nonce) is preceded on the wire by a 4-byte big-endian
// uint32 length prefix. A zero-length frame closes the stream.
//
// # Frame format
//
//	[ 4-byte BE uint32 frame length ][ ciphertext = nonce || enc(plain) || tag ]
//	[ 4-byte BE uint32 frame length ][ ciphertext                              ]
//	...
//	[ 4-byte BE uint32 = 0 ]   ← end-of-stream sentinel
//
// # Ordering / truncation protection
//
// The 64-bit frame counter (starting at 0, big-endian) is appended to the
// caller-supplied aad before sealing. Reordering, dropping, or duplicating
// frames therefore fails AEAD authentication on the receiver. The
// end-of-stream zero-length sentinel detects truncation of the final
// frame: a DecryptStream that hits io.EOF without seeing the sentinel
// returns ErrStreamTruncated.
//
// # Nonces
//
// EncryptStream relies on the per-frame random nonce produced by
// method.Encrypt; nonces are embedded inside each ciphertext frame. This
// avoids a stateful counter-based nonce derivation and keeps the
// streaming API a thin wrapper over the existing one-shot primitive.
func (m *Method) EncryptStream(key ctypes.Bytes, src io.Reader, dst io.Writer, aad ctypes.Bytes) error {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.encryptFn, "method encryptFn is nil")

	if src == nil {
		return ErrEncryption(ErrStreamSrcNil)
	}

	if dst == nil {
		return ErrEncryption(ErrStreamDstNil)
	}

	buf := make([]byte, StreamFrameSize)
	frameAAD := make([]byte, 0, len(aad)+streamFrameCounterSize)

	var counter uint64

	for {
		n, readErr := io.ReadFull(src, buf)
		if n > 0 {
			err := m.sealAndWriteFrame(key, buf[:n], aad, frameAAD, dst, counter)
			if err != nil {
				return err
			}

			counter++
		}

		if readErr == nil {
			continue
		}

		if errors.Is(readErr, io.EOF) || errors.Is(readErr, io.ErrUnexpectedEOF) {
			break
		}

		return ErrEncryption(cerrs.Wrap(ErrEncryptFailed, readErr))
	}

	// End-of-stream sentinel: a zero-length frame.
	writeErr := writeFrameLength(dst, 0)
	if writeErr != nil {
		return ErrEncryption(writeErr)
	}

	return nil
}

// sealAndWriteFrame seals a single plaintext frame and writes it to dst with
// its length prefix. frameAADBuf is a reusable scratch buffer.
func (m *Method) sealAndWriteFrame(key ctypes.Bytes, plain []byte, aad ctypes.Bytes, frameAADBuf []byte, dst io.Writer, counter uint64) error {
	frameAAD := appendFrameAAD(frameAADBuf[:0], aad, counter)

	ciphered, err := m.encryptFn(m, key, plain, frameAAD)
	if err != nil {
		return ErrEncryption(err)
	}

	writeErr := writeFrame(dst, ciphered)
	if writeErr != nil {
		return ErrEncryption(writeErr)
	}

	return nil
}

// DecryptStream reads a chunked AEAD ciphertext produced by EncryptStream
// from src and writes the recovered plaintext to dst. Each frame is opened
// independently with method.Decrypt; the caller-supplied aad is augmented
// with the 64-bit frame counter before authentication, matching
// EncryptStream.
//
// Truncation detection: the final zero-length sentinel frame is mandatory.
// If src reaches EOF without it, DecryptStream returns an error wrapping
// ErrStreamTruncated.
//
// Tampering detection: any single byte flipped inside a ciphertext frame
// triggers an AEAD authentication failure on that frame and aborts the
// stream with ErrDecryption.
//
// Reordering detection: the per-frame counter inside the AAD ensures that
// a swapped frame fails authentication. Replay of a previously-seen frame
// at the wrong index likewise fails.
func (m *Method) DecryptStream(key ctypes.Bytes, src io.Reader, dst io.Writer, aad ctypes.Bytes) error {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.decryptFn, "method decryptFn is nil")

	if src == nil {
		return ErrDecryption(ErrStreamSrcNil)
	}

	if dst == nil {
		return ErrDecryption(ErrStreamDstNil)
	}

	frameAAD := make([]byte, 0, len(aad)+streamFrameCounterSize)

	var counter uint64

	for {
		length, lengthErr := readFrameLength(src)
		if lengthErr != nil {
			return wrapDecryptReadError(lengthErr)
		}

		if length == 0 {
			return nil
		}

		if length > streamCipherMaxFrameSize {
			return ErrDecryption(cerrs.Wrap(ErrStreamFrameTooLarge, ErrDecryptFailed))
		}

		err := m.openAndWriteFrame(key, src, dst, aad, frameAAD, length, counter)
		if err != nil {
			return err
		}

		counter++
	}
}

// openAndWriteFrame reads a single ciphertext frame, opens it, and writes
// the recovered plaintext to dst. frameAADBuf is a reusable scratch buffer.
func (m *Method) openAndWriteFrame(key ctypes.Bytes, src io.Reader, dst io.Writer, aad ctypes.Bytes, frameAADBuf []byte, length uint32, counter uint64) error {
	ciphered := make([]byte, length)

	_, readErr := io.ReadFull(src, ciphered)
	if readErr != nil {
		return wrapDecryptReadError(readErr)
	}

	frameAAD := appendFrameAAD(frameAADBuf[:0], aad, counter)

	plain, decErr := m.decryptFn(m, key, ciphered, frameAAD)
	if decErr != nil {
		return ErrDecryption(decErr)
	}

	_, writeErr := dst.Write(plain)
	if writeErr != nil {
		return ErrDecryption(cerrs.Wrap(ErrDecryptFailed, writeErr))
	}

	return nil
}

// wrapDecryptReadError categorises a read error encountered during
// DecryptStream. io.EOF / io.ErrUnexpectedEOF before the end-of-stream
// sentinel are surfaced as ErrStreamTruncated wrapped in ErrDecryption.
func wrapDecryptReadError(err error) error {
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return ErrDecryption(cerrs.Wrap(ErrStreamTruncated, err))
	}

	return ErrDecryption(cerrs.Wrap(ErrDecryptFailed, err))
}

// appendFrameAAD appends the 64-bit big-endian frame counter to the
// caller-supplied AAD. The result is the per-frame AAD that
// EncryptStream / DecryptStream bind into AEAD authentication.
func appendFrameAAD(dst []byte, aad []byte, counter uint64) []byte {
	dst = append(dst, aad...)
	dst = binary.BigEndian.AppendUint64(dst, counter)

	return dst
}

// writeFrame writes a single length-prefixed frame to dst.
func writeFrame(dst io.Writer, ciphered []byte) error {
	err := writeFrameLength(dst, uint32(len(ciphered))) //nolint:gosec // frame length is bounded by StreamFrameSize + tag/nonce overhead, well below MaxUint32
	if err != nil {
		return err
	}

	_, writeErr := dst.Write(ciphered)
	if writeErr != nil {
		return cerrs.Wrap(ErrEncryptFailed, writeErr)
	}

	return nil
}

// writeFrameLength writes a 4-byte big-endian uint32 length prefix.
func writeFrameLength(dst io.Writer, length uint32) error {
	var buf [streamFrameLengthSize]byte
	binary.BigEndian.PutUint32(buf[:], length)

	_, err := dst.Write(buf[:])
	if err != nil {
		return cerrs.Wrap(ErrEncryptFailed, err)
	}

	return nil
}

// readFrameLength reads a 4-byte big-endian uint32 length prefix from src.
func readFrameLength(src io.Reader) (uint32, error) {
	var buf [streamFrameLengthSize]byte

	_, err := io.ReadFull(src, buf[:])
	if err != nil {
		return 0, err //nolint:wrapcheck // surfaced unwrapped so callers can match io.EOF / io.ErrUnexpectedEOF
	}

	return binary.BigEndian.Uint32(buf[:]), nil
}
