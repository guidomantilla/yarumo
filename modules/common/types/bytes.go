package types

import (
	"encoding/base64"
	"encoding/hex"
)

type Bytes []byte

func (b Bytes) ToHex() string {
	return hex.EncodeToString(b)
}

func (b Bytes) ToBase64Std() string {
	return base64.StdEncoding.EncodeToString(b)
}

func (b Bytes) ToBase64RawStd() string {
	return base64.RawStdEncoding.EncodeToString(b)
}

func (b Bytes) ToBase64Url() string {
	return base64.URLEncoding.EncodeToString(b)
}

func (b Bytes) ToBase64RawUrl() string {
	return base64.RawURLEncoding.EncodeToString(b)
}
