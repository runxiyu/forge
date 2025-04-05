package misc

func FirstOrPanic[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func NoneOrPanic(err error) {
	if err != nil {
		panic(err)
	}
}
