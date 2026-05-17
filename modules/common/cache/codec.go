package cache

import "encoding/json"

// JSONCodec is the default Codec, backed by encoding/json.
type JSONCodec struct{}

// Encode marshals v using encoding/json.Marshal.
func (JSONCodec) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Decode unmarshals data into v using encoding/json.Unmarshal. v must be a
// non-nil pointer.
func (JSONCodec) Decode(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
