package topsis

import "github.com/guidomantilla/yarumo/compute/engine/mcdm"

// ErrRank creates an MCDM domain error for TOPSIS ranking failures.
func ErrRank(errs ...error) error {
	return mcdm.ErrMCDM(errs...)
}
