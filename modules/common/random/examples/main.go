// Package main demonstrates common/random: non-secure helpers backed by
// math/rand/v2 for generating bytes, bounded numbers, arbitrary-charset
// strings, and the TextXxx convenience aliases. Use common/crypto/random
// for any value that must be unpredictable (tokens, secrets, IDs that
// face the internet).
package main

import (
	"fmt"

	crandom "github.com/guidomantilla/yarumo/common/random"
)

func main() {
	demoBytes()
	demoNumber()
	demoString()
	demoTextHelpers()
}

// demoBytes generates a short random byte slice and renders it as hex.
// The returned value is a common/types.Bytes — methods like ToHex are
// available without importing common/types directly.
func demoBytes() {
	fmt.Println("=== Bytes ===")

	out := crandom.Bytes(8)
	fmt.Printf("  Bytes(8)        -> %d bytes\n", len(out))
	fmt.Printf("  .ToHex()        -> %s\n", out.ToHex())
	fmt.Printf("  .ToBase64Std()  -> %s\n", out.ToBase64Std())
}

// demoNumber draws a few values in [0, limit).
func demoNumber() {
	fmt.Println("=== Number ===")

	for i := 0; i < 3; i++ {
		fmt.Printf("  Number(100) -> %d\n", crandom.Number(100))
	}

	fmt.Printf("  Number(0)   -> %d (zero limit returns 0)\n", crandom.Number(0))
}

// demoString generates a random string from an arbitrary charset.
func demoString() {
	fmt.Println("=== String ===")

	fmt.Printf("  String(10, AlphaSet)   -> %q\n", crandom.String(10, crandom.AlphaSet))
	fmt.Printf("  String(8, \"01\")        -> %q\n", crandom.String(8, "01"))
	fmt.Printf("  String(5, \"\")          -> %q (empty charset -> empty result)\n", crandom.String(5, ""))
}

// demoTextHelpers exercises the convenience TextXxx aliases.
func demoTextHelpers() {
	fmt.Println("=== TextXxx ===")

	fmt.Printf("  TextLower(6)     -> %q\n", crandom.TextLower(6))
	fmt.Printf("  TextUpper(6)     -> %q\n", crandom.TextUpper(6))
	fmt.Printf("  TextNumber(6)    -> %q\n", crandom.TextNumber(6))
	fmt.Printf("  TextAlphaNum(10) -> %q\n", crandom.TextAlphaNum(10))
}
