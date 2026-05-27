// Package main demonstrates common/types: the Bytes alias type and its
// hex / base64 encoding helpers used throughout the workspace for
// crypto outputs, REST raw bodies, and cache codecs.
package main

import (
	"fmt"

	ctypes "github.com/guidomantilla/yarumo/core/common/types"
)

func main() {
	demoHex()
	demoBase64Std()
	demoBase64Url()
	demoRoundTrip()
	demoConversion()
}

// demoHex encodes a byte slice as lowercase hex.
func demoHex() {
	fmt.Println("=== ToHex ===")

	b := ctypes.Bytes("yarumo")
	fmt.Printf("  Bytes(%q).ToHex() -> %s\n", string(b), b.ToHex())
}

// demoBase64Std exercises the padded and unpadded standard Base64 variants.
func demoBase64Std() {
	fmt.Println("=== ToBase64Std / ToBase64RawStd ===")

	b := ctypes.Bytes("hello world")
	fmt.Printf("  ToBase64Std()    -> %s\n", b.ToBase64Std())
	fmt.Printf("  ToBase64RawStd() -> %s\n", b.ToBase64RawStd())
}

// demoBase64Url exercises the URL-safe variants.
func demoBase64Url() {
	fmt.Println("=== ToBase64Url / ToBase64RawUrl ===")

	b := ctypes.Bytes{0xfb, 0xef, 0xff} // would contain '+' and '/' in std base64
	fmt.Printf("  ToBase64Url()    -> %s\n", b.ToBase64Url())
	fmt.Printf("  ToBase64RawUrl() -> %s\n", b.ToBase64RawUrl())
}

// demoRoundTrip shows hex encoding is stable for ASCII input.
func demoRoundTrip() {
	fmt.Println("=== Round trip ===")

	b := ctypes.Bytes("abc")
	fmt.Printf("  hex of \"abc\" -> %s\n", b.ToHex())
	fmt.Printf("  len(b)        -> %d\n", len(b))
}

// demoConversion shows that ctypes.Bytes is a []byte under the hood and
// composes with stdlib helpers naturally.
func demoConversion() {
	fmt.Println("=== Conversion ===")

	plain := []byte{0x00, 0x10, 0x20, 0x30}
	wrapped := ctypes.Bytes(plain)

	fmt.Printf("  []byte -> ctypes.Bytes -> hex: %s\n", wrapped.ToHex())
	fmt.Printf("  back to []byte len:           %d\n", len([]byte(wrapped)))
}
