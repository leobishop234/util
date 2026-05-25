package ptr

func Ptr[T any](t T) *T {
	return &t
}

func PtrDefaultToNil[T comparable](t T) *T {
	var check T
	if t == check {
		return nil
	}
	return &t
}

func Deref[T any](t *T, d T) T {
	if t == nil {
		return d
	}
	return *t
}

func DerefDefault[T any](t *T) T {
	var d T
	if t == nil {
		return d
	}
	return *t
}
