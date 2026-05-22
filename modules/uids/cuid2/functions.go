package cuid2

import (
	"regexp"

	"github.com/akshayvadher/cuid2"

	"github.com/guidomantilla/yarumo/uids"
)

// Cuid2 is the preconfigured CUID2 generator singleton. Consumers that
// want registry-based lookup must register it explicitly via
// uids.Register(Cuid2) at startup.
var Cuid2 = uids.NewUID(Name, CUID2)

// cuid2Regex matches the canonical CUID2 format: 24 characters, lowercase
// alphanumeric, starting with a letter. Although
// github.com/akshayvadher/cuid2 exposes IsCuid, that helper accepts any
// length between 2 and 32 — too permissive for the default-length CUID2
// emitted by Generate. Length 24 is anchored here to match the canonical
// output of cuid2.CreateId.
var cuid2Regex = regexp.MustCompile(`^[a-z][a-z0-9]{23}$`)

// CUID2 generates a collision-resistant unique identifier.
func CUID2() (string, error) {
	return cuid2.CreateId(), nil
}

// IsCUID2 reports whether s is a syntactically valid CUID2: exactly 24
// characters, lowercase alphanumeric, starting with a letter. Non-default
// lengths produced by CreateIdOf are intentionally rejected.
func IsCUID2(s string) bool {
	return cuid2Regex.MatchString(s)
}
