package validation

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// ValidationType is the error type classification for validation failures.
const ValidationType = "validation"

var _ error = (*Error)(nil)

// Error is the domain error for validation failures. It embeds errs.TypedError
// so AsErrorInfo groups violations under the "validation" type.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted validation error message.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("validation %s error: %s", e.Type, e.Err)
}

// Sentinel errors for the validation leaves and reflection helpers.
var (
	ErrValidationFailed     = errors.New("validation failed")
	ErrFieldRequired        = errors.New("field is required")
	ErrFieldMustBeUndefined = errors.New("field must be undefined")
	ErrMinLen               = errors.New("value is shorter than the minimum length")
	ErrMaxLen               = errors.New("value is longer than the maximum length")
	ErrRegexInvalid         = errors.New("regex pattern is invalid")
	ErrRegexMismatch        = errors.New("value does not match the required pattern")
	ErrEmailInvalid         = errors.New("value is not a valid email")
	ErrURLInvalid           = errors.New("value is not a valid URL")
	ErrMinValue             = errors.New("value is below the minimum")
	ErrMaxValue             = errors.New("value is above the maximum")
	ErrOutOfRange           = errors.New("value is out of the allowed range")
	ErrInvalidRange         = errors.New("min must be less than or equal to max")
	ErrUIDInvalid           = errors.New("value is not a valid unique identifier")
	ErrJWTInvalid           = errors.New("value is not a valid JWT")
	ErrSemverInvalid        = errors.New("value is not a valid semver string")
	ErrCollectionEmpty      = errors.New("collection is empty")
	ErrEachFailed           = errors.New("collection element failed validation")
	ErrPathInvalid          = errors.New("field path is invalid")
	ErrPathNotFound         = errors.New("field path not found")
	ErrPathTypeMismatch     = errors.New("field path traversal hit incompatible type")
	ErrIndexOutOfRange      = errors.New("slice index is out of range")
	ErrObjectNil            = errors.New("target object is nil")
	ErrContainsMissing      = errors.New("value does not contain the required substring")
	ErrPrefixMissing        = errors.New("value does not have the required prefix")
	ErrSuffixMissing        = errors.New("value does not have the required suffix")
	ErrNotLowercase         = errors.New("value is not lowercase")
	ErrNotUppercase         = errors.New("value is not uppercase")
	ErrNotAlpha             = errors.New("value is not alphabetic")
	ErrNotAlphanumeric      = errors.New("value is not alphanumeric")
	ErrNotNumeric           = errors.New("value is not numeric")
	ErrNotASCII             = errors.New("value is not ASCII")
	ErrNotHex               = errors.New("value is not hexadecimal")
	ErrBase64Invalid        = errors.New("value is not valid base64")
	ErrNotTrimmed           = errors.New("value has leading or trailing whitespace")
	ErrNotPositive          = errors.New("value is not positive")
	ErrNotNegative          = errors.New("value is not negative")
	ErrZero                 = errors.New("value must be non-zero")
	ErrNotMultipleOf        = errors.New("value is not a multiple of the required factor")
	ErrIntegerStringInvalid = errors.New("value is not a valid integer string")
	ErrFloatStringInvalid   = errors.New("value is not a valid float string")
	ErrNotEqual             = errors.New("value is not equal to the expected value")
	ErrMustNotEqual         = errors.New("value must not equal the forbidden value")
	ErrNotInAllowed         = errors.New("value is not in the allowed set")
	ErrInForbidden          = errors.New("value is in the forbidden set")
	ErrEmptyAllowed         = errors.New("the allowed set is empty")
	ErrIPInvalid            = errors.New("value is not a valid IP address")
	ErrIPv4Invalid          = errors.New("value is not a valid IPv4 address")
	ErrIPv6Invalid          = errors.New("value is not a valid IPv6 address")
	ErrCIDRInvalid          = errors.New("value is not a valid CIDR")
	ErrMACInvalid           = errors.New("value is not a valid MAC address")
	ErrHostnameInvalid      = errors.New("value is not a valid hostname")
	ErrFQDNInvalid          = errors.New("value is not a valid fully qualified domain name")
	ErrPortInvalid          = errors.New("value is not a valid port number")
	ErrDateInvalid          = errors.New("value does not match the date layout")
	ErrLayoutInvalid        = errors.New("date layout is invalid")
	ErrTimeBefore           = errors.New("time is not before the reference")
	ErrTimeAfter            = errors.New("time is not after the reference")
	ErrTimeOutOfRange       = errors.New("time is out of the allowed range")
	ErrInvalidTimeRange     = errors.New("time range lower bound must not exceed upper bound")
	ErrCountBelowMin        = errors.New("collection has fewer elements than the minimum")
	ErrCountAboveMax        = errors.New("collection has more elements than the maximum")
	ErrCountOutOfRange      = errors.New("collection size is out of the allowed range")
	ErrInvalidCountRange    = errors.New("count range lower bound must not exceed upper bound")
	ErrDuplicate            = errors.New("collection contains a duplicate element")
	ErrNotSortedAsc         = errors.New("collection is not sorted in ascending order")
	ErrNotSortedDesc        = errors.New("collection is not sorted in descending order")
	ErrKeyMissing           = errors.New("map does not contain the required key")
	ErrMinKeys              = errors.New("map has fewer keys than the minimum")
	ErrMaxKeys              = errors.New("map has more keys than the maximum")
	ErrAssertionInverted    = errors.New("inverted assertion was unexpectedly satisfied")
)

// ErrValidation creates a validation domain error joining the given causes with
// ErrValidationFailed.
func ErrValidation(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: ValidationType,
			Err:  errors.Join(append(causes, ErrValidationFailed)...),
		},
	}
}
