# Fuzzing crypto parsers

The `tokens` and `certs` packages accept attacker-controlled bytes (JWT
strings and PEM blocks respectively). Both wrap mature standard libraries
(`golang-jwt/v5`, `crypto/x509`), but the wrapping layer has its own surface
area: length checks, key resolution, block-type assertions, claim envelope
unmarshalling. Fuzz tests guard the no-panic contract on that wrapping
layer.

## Available fuzz harnesses

### `modules/common/crypto/tokens`

| Harness | Target |
| --- | --- |
| `FuzzValidate` | `Method.Validate` with a keyed HS256 method (reaches `jwt.ParseWithClaims`). |
| `FuzzDecodeUnsafe` | `Method.DecodeUnsafe` (no key, unverified `jwt.NewParser` path). |

### `modules/common/crypto/certs`

| Harness | Target |
| --- | --- |
| `FuzzClientTlsFromPEM` | `ClientTlsFromPEM(serverName, ca, cert, key, insecure)`. |
| `FuzzServerTlsFromPEM` | `ServerTlsFromPEM(cert, key, clientCA)`. |
| `FuzzParseCertificatePEM` | `ParseCertificatePEM(pem)`. |
| `FuzzParsePEMChain` | `ParsePEMChain(pem)`. |
| `FuzzParseCSR` | `ParseCSR(pem)`. |

## Running fuzz tests locally

From any worktree, anchored at the `modules/common` go.mod:

```bash
# Tokens
go test -run=^$ -fuzz=^FuzzValidate$         -fuzztime=30s ./crypto/tokens/
go test -run=^$ -fuzz=^FuzzDecodeUnsafe$     -fuzztime=30s ./crypto/tokens/

# Certs
go test -run=^$ -fuzz=^FuzzClientTlsFromPEM$    -fuzztime=30s ./crypto/certs/
go test -run=^$ -fuzz=^FuzzServerTlsFromPEM$    -fuzztime=30s ./crypto/certs/
go test -run=^$ -fuzz=^FuzzParseCertificatePEM$ -fuzztime=30s ./crypto/certs/
go test -run=^$ -fuzz=^FuzzParsePEMChain$       -fuzztime=30s ./crypto/certs/
go test -run=^$ -fuzz=^FuzzParseCSR$            -fuzztime=30s ./crypto/certs/
```

`-run=^$` disables the regular `Test*` set so that only the fuzz harness
runs. `-fuzz=^<Name>$` picks a single harness. Bump `-fuzztime` (e.g. `5m`,
`1h`) for longer campaigns when triaging a suspect surface; 30s is enough
to clear the acceptance criteria for YA-0032 but not enough to claim a
package is bug-free.

To replay only the committed seed corpus without engaging the mutator,
drop `-fuzz` and use `-run`:

```bash
go test -run=^FuzzValidate$ ./crypto/tokens/
```

## Seed corpus

Seeds live under `<package>/testdata/fuzz/<HarnessName>/`. Each file is a
plain-text `go test fuzz v1` blob that the runtime loads at startup. We
keep the seeds focused on known-edge inputs rather than coverage padding:

- `tokens/FuzzValidate/`: alg=none JWT, missing signature, oversized
  10KB subject claim, embedded null byte, empty string.
- `tokens/FuzzDecodeUnsafe/`: alg=none JWT, missing signature.
- `certs/FuzzClientTlsFromPEM/`: malformed PEM headers, truncated
  certificate body, embedded null bytes.
- `certs/FuzzServerTlsFromPEM/`: malformed PEM headers, truncated.
- `certs/FuzzParseCertificatePEM/`: wrong block type, truncated.
- `certs/FuzzParsePEMChain/`: multi-block CERTIFICATE, mixed block type.
- `certs/FuzzParseCSR/`: wrong block type, truncated.

## When fuzz finds a panic

The Go fuzz runtime writes new failure inputs to `$GOCACHE/fuzz/<pkg>/<Harness>/`
(not `testdata/`). If a panic shows up:

1. Copy the offending corpus file into `testdata/fuzz/<Harness>/` so the
   crash is reproducible without the fuzz cache.
2. Open a yarumo issue with the panic stack and the corpus blob.
3. Fix the panic. Re-run the same harness; the seed copy now serves as a
   permanent regression test.

Do not silently delete a panic seed. Per the workspace defensive-validation
standard, any panic on a runtime input is a bug — `cassert.True` /
`cassert.NotNil` belong only to struct invariants, never to caller bytes.
