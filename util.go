package main

import (
	"math"
	"strconv"
)

func itoa(i int) string { return strconv.Itoa(i) }

// floorToInt returns floor(v) as an int.
func floorToInt(v float64) int { return int(math.Floor(v)) }

// hexDigit returns the lowercase hex digit for n (0-15), matching Python's
// hex(n)[2:] for the single-digit values the game uses.
func hexDigit(n int) string {
	const digits = "0123456789abcdef"
	if n < 0 || n >= len(digits) {
		return "0"
	}
	return string(digits[n])
}

// parseHexDigit converts a single hex character to its value, or -1 if invalid.
func parseHexDigit(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minf(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxf(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func absf(x float64) float64 { return math.Abs(x) }
