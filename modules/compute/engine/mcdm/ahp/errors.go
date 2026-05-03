package ahp

import "github.com/guidomantilla/yarumo/compute/engine/mcdm"

// ErrAnalyze creates an MCDM domain error for AHP analysis failures.
func ErrAnalyze(errs ...error) error {
	return mcdm.ErrMCDM(errs...)
}

// ErrRank creates an MCDM domain error for AHP ranking failures.
func ErrRank(errs ...error) error {
	return mcdm.ErrMCDM(errs...)
}
