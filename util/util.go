package util

import "math"

func Round(val float64, n int) float64 {
	return math.Round(val*float64(n)) / float64(n)
}