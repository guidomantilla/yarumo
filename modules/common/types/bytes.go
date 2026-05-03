// Package types provides common type definitions shared across the project.
package types

import (
	"encoding/base64"
	"encoding/hex"
)

// Bytes is a byte slice with convenience methods for encoding.
type Bytes []byte

// ToHex converts the Bytes receiver into its hexadecimal string representation.
func (b Bytes) ToHex() string {
	return hex.EncodeToString(b)
}

// ToBase64Std converts the Bytes receiver to a standard Base64-encoded string using padding.
func (b Bytes) ToBase64Std() string {
	return base64.StdEncoding.EncodeToString(b)
}

// ToBase64RawStd converts the Bytes receiver to a standard Base64-encoded string without padding.
func (b Bytes) ToBase64RawStd() string {
	return base64.RawStdEncoding.EncodeToString(b)
}

// ToBase64Url converts the Bytes receiver to a URL-safe Base64-encoded string using padding.
func (b Bytes) ToBase64Url() string {
	return base64.URLEncoding.EncodeToString(b)
}

// ToBase64RawUrl converts the Bytes receiver to a URL-safe Base64-encoded string without padding.
func (b Bytes) ToBase64RawUrl() string {
	return base64.RawURLEncoding.EncodeToString(b)
}
