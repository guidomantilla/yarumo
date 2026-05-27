package validation

// adapters.go bridges typed-generic leaves from common/validation into the
// RuleFn shape the engine expects. Each helper handles the same boilerplate
// the manual adapters used to write: coerce the runtime value, coerce the
// params, invoke the typed function. Coercion failures bubble up as the
// same ErrBadParams the manual adapters produced.

// AdaptString turns a string-only check (CheckFn[string]) into a RuleFn.
func AdaptString(f func(s string) error) RuleFn {
	return func(value any, _ []any) error {
		s, err := asString(value)
		if err != nil {
			return err
		}

		return f(s)
	}
}

// AdaptStringWithParam turns a (s, param string) check into a RuleFn that
// reads param from params[0].
func AdaptStringWithParam(f func(s, p string) error) RuleFn {
	return func(value any, params []any) error {
		s, err := asString(value)
		if err != nil {
			return err
		}

		p, err := asStringParam(params, 0)
		if err != nil {
			return err
		}

		return f(s, p)
	}
}

// AdaptStringWithInt turns a (s string, n int) check into a RuleFn that
// reads n from params[0].
func AdaptStringWithInt(f func(s string, n int) error) RuleFn {
	return func(value any, params []any) error {
		s, err := asString(value)
		if err != nil {
			return err
		}

		n, err := asInt(params, 0)
		if err != nil {
			return err
		}

		return f(s, n)
	}
}

// AdaptNumeric turns a unary numeric check (v T) into a RuleFn over float64
// since the engine carries numerics as floats after coercion.
func AdaptNumeric(f func(v float64) error) RuleFn {
	return func(value any, _ []any) error {
		v, err := asFloat(value)
		if err != nil {
			return err
		}

		return f(v)
	}
}

// AdaptNumericBinary turns a (v, p T) check into a RuleFn that reads p
// from params[0].
func AdaptNumericBinary(f func(v, p float64) error) RuleFn {
	return func(value any, params []any) error {
		v, err := asFloat(value)
		if err != nil {
			return err
		}

		p, err := asFloatParam(params, 0)
		if err != nil {
			return err
		}

		return f(v, p)
	}
}

// AdaptNumericRange turns a (v, lo, hi T) check into a RuleFn that reads
// lo from params[0] and hi from params[1].
func AdaptNumericRange(f func(v, lo, hi float64) error) RuleFn {
	return func(value any, params []any) error {
		v, err := asFloat(value)
		if err != nil {
			return err
		}

		lo, err := asFloatParam(params, 0)
		if err != nil {
			return err
		}

		hi, err := asFloatParam(params, 1)
		if err != nil {
			return err
		}

		return f(v, lo, hi)
	}
}

// AdaptStringSet turns a (v string, set []string) check into a RuleFn that
// reads the set from params (variadic, all strings).
func AdaptStringSet(f func(v string, set []string) error) RuleFn {
	return func(value any, params []any) error {
		s, err := asString(value)
		if err != nil {
			return err
		}

		set := make([]string, 0, len(params))
		for i := range params {
			p, ok := params[i].(string)
			if !ok {
				return ErrEngine(ErrBadParams)
			}

			set = append(set, p)
		}

		return f(s, set)
	}
}

// AdaptCollection turns a (xs []any) check into a RuleFn. The runtime
// value is coerced via asSlice so typed slices and arrays decompose
// transparently.
func AdaptCollection(f func(xs []any) error) RuleFn {
	return func(value any, _ []any) error {
		xs, err := asSlice(value)
		if err != nil {
			return err
		}

		return f(xs)
	}
}

// AdaptCollectionWithInt turns a (xs []any, n int) check into a RuleFn.
func AdaptCollectionWithInt(f func(xs []any, n int) error) RuleFn {
	return func(value any, params []any) error {
		xs, err := asSlice(value)
		if err != nil {
			return err
		}

		n, err := asInt(params, 0)
		if err != nil {
			return err
		}

		return f(xs, n)
	}
}
