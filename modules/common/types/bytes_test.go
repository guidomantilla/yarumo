package types

import "testing"

func TestBytes_Encodings(t *testing.T) {
	tests := []struct {
		name      string
		in        Bytes
		hex       string
		b64Std    string
		b64RawStd string
		b64URL    string
		b64RawURL string
	}{
		{
			name:      "empty",
			in:        Bytes{},
			hex:       "",
			b64Std:    "",
			b64RawStd: "",
			b64URL:    "",
			b64RawURL: "",
		},
		{
			name:      "hello",
			in:        Bytes([]byte("hello")),
			hex:       "68656c6c6f",
			b64Std:    "aGVsbG8=",
			b64RawStd: "aGVsbG8",
			b64URL:    "aGVsbG8=",
			b64RawURL: "aGVsbG8",
		},
		{
			name: "binary_with_plus_and_slash_base64",
			// input chosen so that standard base64 includes '+' and '/'
			in:        Bytes([]byte{251, 255, 239}),
			hex:       "fbffef",
			b64Std:    "+//v",
			b64RawStd: "+//v",
			b64URL:    "-__v",
			b64RawURL: "-__v",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.ToHex(); got != tt.hex {
				t.Fatalf("ToHex() = %q, want %q", got, tt.hex)
			}

			if got := tt.in.ToBase64Std(); got != tt.b64Std {
				t.Fatalf("ToBase64Std() = %q, want %q", got, tt.b64Std)
			}

			if got := tt.in.ToBase64RawStd(); got != tt.b64RawStd {
				t.Fatalf("ToBase64RawStd() = %q, want %q", got, tt.b64RawStd)
			}

			if got := tt.in.ToBase64Url(); got != tt.b64URL {
				t.Fatalf("ToBase64Url() = %q, want %q", got, tt.b64URL)
			}

			if got := tt.in.ToBase64RawUrl(); got != tt.b64RawURL {
				t.Fatalf("ToBase64RawUrl() = %q, want %q", got, tt.b64RawURL)
			}
		})
	}
}
