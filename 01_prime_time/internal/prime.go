package internal

import "math"

// IsPrime checks if the given integer is a prime number. Returns true if it's a prime, false otherwise
func IsPrime(n int64) bool {
	if n <= 1 {
		return false
	}
	if n == 2 || n == 3 {
		return true
	}
	// Use property that all primes are of the form 6k +- 1 (except 2 & 3)
	rem := n % 6
	if rem == 0 || rem == 2 || rem == 3 || rem == 4 {
		return false
	}

	m := int64(math.Sqrt(float64(n)))

	// Only iterate over odd numbers, cutting down the number of iterations in half
	for i := int64(3); i <= m; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}
