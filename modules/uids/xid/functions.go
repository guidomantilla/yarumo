package xid

import (
	"github.com/rs/xid"

	"github.com/guidomantilla/yarumo/uids"
)

// XId is the preconfigured XID generator singleton. Consumers that want
// registry-based lookup must register it explicitly via uids.Register(XId)
// at startup.
var XId = uids.NewUID(Name, XID)

// XID generates a globally unique ID inspired by MongoDB ObjectID.
func XID() (string, error) {
	return xid.New().String(), nil
}

// IsXID reports whether s is a syntactically valid XID: 20 characters in
// base32hex encoding as produced by github.com/rs/xid.
func IsXID(s string) bool {
	_, err := xid.FromString(s)
	return err == nil
}
