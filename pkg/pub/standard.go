package pub

func defaultValue[T comparable](val, def, nilValue T) T {
	if val != nilValue {
		return val
	}
	return def
}

func stringInSlice(s string, ss []string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
