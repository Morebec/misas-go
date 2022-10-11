package maputils

func MapHasKey[K string | int, T any](m map[K]T, k K) bool {
	if _, found := m[k]; found {
		return true
	}

	return false
}

func MapGetOrDefault[K string | int, T any](m map[K]T, k K, defaultValue T) T {
	if MapHasKey[K, T](m, k) {
		return m[k]
	}

	return defaultValue
}

func MapValues[K string | int, T any](m map[K]T) []T {
	var out []T
	for _, v := range m {
		out = append(out, v)
	}

	return out
}

func MapKeys[K string | int, T any](m map[K]T) []K {
	var keys []K
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
