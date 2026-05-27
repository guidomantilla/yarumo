// Copyright 2026 Guido Mauricio Mantilla Tarazona
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package crypto re-exports every subpackage so the module dependency-graph
// image tooling picks them up. It defines no symbols of its own.
package crypto

import (
	_ "github.com/guidomantilla/yarumo/core/crypto/certs"
	_ "github.com/guidomantilla/yarumo/core/crypto/ciphers/aead"
	_ "github.com/guidomantilla/yarumo/core/crypto/ciphers/hybrid"
	_ "github.com/guidomantilla/yarumo/core/crypto/ciphers/rsaoaep"
	_ "github.com/guidomantilla/yarumo/core/crypto/hashes"
	_ "github.com/guidomantilla/yarumo/core/crypto/kdfs"
	_ "github.com/guidomantilla/yarumo/core/crypto/passwords"
	_ "github.com/guidomantilla/yarumo/core/crypto/passwords/generator"
	_ "github.com/guidomantilla/yarumo/core/crypto/random"
	_ "github.com/guidomantilla/yarumo/core/crypto/signers/ecdsas"
	_ "github.com/guidomantilla/yarumo/core/crypto/signers/ed25519"
	_ "github.com/guidomantilla/yarumo/core/crypto/signers/hmacs"
	_ "github.com/guidomantilla/yarumo/core/crypto/signers/rsassas"
	_ "github.com/guidomantilla/yarumo/core/crypto/tokens"
)
