package parser

// ParseOptions control how the parser behaves.
// Strict=false (default) enables synonyms and Unicode operators/keywords.
// Strict=true accepts only the canonical ASCII grammar: !, &, |, =>, <=>, (, ), identifiers.
// In strict mode, keywords like AND/OR/NOT/THEN/IFF and TRUE/FALSE are treated as identifiers.
type ParseOptions struct {
	Strict bool
}
