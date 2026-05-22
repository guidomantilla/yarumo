package nanoid

import (
	"regexp"

	nanoid "github.com/devmiek/nanoid-go"

	"github.com/guidomantilla/yarumo/uids"
)

// NanoID is the preconfigured NanoID generator singleton. Consumers that
// want registry-based lookup must register it explicitly via
// uids.Register(NanoID) at startup.
var NanoID = uids.NewUID(Name, NANOID)

// nanoIDRegex matches the default NanoID format: 21 characters from the
// URL-safe alphabet (A-Z, a-z, 0-9, _, -). The upstream
// github.com/devmiek/nanoid-go library does not expose a parser, so the
// canonical default alphabet and length are encoded here.
var nanoIDRegex = regexp.MustCompile(`^[A-Za-z0-9_-]{21}$`)

// NANOID generates a tiny, secure, URL-friendly unique string ID.
func NANOID() (string, error) {
	id, err := nanoid.New()
	if err != nil {
		return "", uids.ErrGeneration(err)
	}

	return id, nil
}

// IsNanoID reports whether s matches the default NanoID format: 21
// characters from the URL-safe alphabet (A-Z, a-z, 0-9, underscore, hyphen).
// Custom alphabets or sizes are intentionally rejected.
func IsNanoID(s string) bool {
	return nanoIDRegex.MatchString(s)
}
