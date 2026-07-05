package main

import "math/rand"

// randIntn returns a value in [0, n).
func randIntn(n int) int {
	if n <= 0 {
		return 0
	}
	return rand.Intn(n)
}

// randInt returns a value in [a, b] inclusive, like Python's random.randint.
func randInt(a, b int) int {
	if b < a {
		a, b = b, a
	}
	return a + rand.Intn(b-a+1)
}

// randFloat returns a value in [0, 1), like Python's random.random.
func randFloat() float64 { return rand.Float64() }

// uniform returns a value in [a, b), like Python's random.uniform.
func uniform(a, b float64) float64 { return a + rand.Float64()*(b-a) }

// choiceInt returns a random element of s, like Python's random.choice.
func choiceInt(s []int) int {
	if len(s) == 0 {
		return 0
	}
	return s[rand.Intn(len(s))]
}
