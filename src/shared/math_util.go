package shared

import "math"

// Common square root values.
var squareRootTable = map[uint64]uint64{
	4:       2,
	16:      4,
	64:      8,
	256:     16,
	1024:    32,
	4096:    64,
	16384:   128,
	65536:   256,
	262144:  512,
	1048576: 1024,
	4194304: 2048,
}

// Max returns the larger integer of the two
// given ones.This is used over the Max function
// in the standard math library because that max function
// has to check for some special floating point cases
// making it slower by a magnitude of 10.
func Max(a uint64, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

// Min returns the smaller integer of the two
// given ones. This is used over the Min function
// in the standard math library because that min function
// has to check for some special floating point cases
// making it slower by a magnitude of 10.
func Min(a uint64, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// IntegerSquareRoot defines a function that returns the
// largest possible integer root of a number using go's standard library.
func IntegerSquareRoot(n uint64) uint64 {
	if v, ok := squareRootTable[n]; ok {
		return v
	}

	return uint64(math.Sqrt(float64(n)))
}

// PowerOf2 returns an integer that is the provided
// exponent of 2. Can only return powers of 2 till 63,
// after that it overflows
func PowerOf2(n uint64) uint64 {
	if n >= 64 {
		panic("integer overflow")
	}
	return 1 << n
}