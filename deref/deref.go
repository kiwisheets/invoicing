package deref

func Int(i *int, or int) int {
	if i == nil {
		return or
	}
	return *i
}

func IntF(i *int, or func() int) int {
	if i == nil {
		return or()
	}
	return *i
}
