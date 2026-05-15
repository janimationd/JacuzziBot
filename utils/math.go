package utils

import (
	"fmt"
	"math/rand/v2"
)

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

// Choose count random distinct options. If count is larger than max you're gonna have problems.
func ChooseRandomNIntegers[T Integer](count int, max T) *Set[T] {
	if count > int(max) {
		panic(fmt.Sprintf("Cannot choose %d distinct options from %d possibilities!", count, max))
	}

	// Fisher-Yates partial shuffle gets us O(n)!
	pool := make([]T, max)
	for i := range pool {
		pool[i] = T(i)
	}

	for i := range count {
		j := i + rand.IntN(int(max)-i)
		pool[i], pool[j] = pool[j], pool[i]
	}

	result := NewSet[T]()
	for _, v := range pool[:count] {
		result.Add(v)
	}
	return result
}

// Get the sign of the numeric value.
func Sign[T Number](val T) int {
	switch {
	case val < 0:
		return -1
	case val > 0:
		return 1
	default:
		return 0
	}
}
