package utils

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Number interface {
	Integer | ~float32 | ~float64
}

// Clamp a value to be within a range (inclusive).
func Clamp[T Number](value T, lower T, upper T) T {
	return min(max(value, lower), upper)
}
