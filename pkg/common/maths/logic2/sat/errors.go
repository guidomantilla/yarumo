package sat

import "errors"

// ErrNotImplemented is returned in Phase 0 for functions not yet available.
var ErrNotImplemented = errors.New("logic2/sat: not implemented (Phase 0)")
