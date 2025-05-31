package storetests

func MockReadReturn[T any](found bool, value T, err error) func(_, _ string, v any) (bool, error) {
	return func(_, _ string, v any) (bool, error) {
		if out, ok := v.(*T); ok {
			*out = value
		}
		return found, err
	}
}

func MockReadNotFound() func(_, _ string, v any) (bool, error) {
	return func(_, _ string, _ any) (bool, error) {
		return false, nil
	}
}
