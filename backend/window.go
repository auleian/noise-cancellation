package main

import "math"

// HannWindow returns a Hann (raised-cosine) window of length n.
//
//	w[i] = 0.5 * (1 - cos(2*pi*i / (n-1)))
//
// When used with 50% overlap, adjacent Hann windows sum to 1.0 (COLA property),
// enabling artifact-free overlap-add reconstruction.
func HannWindow(n int) []float64 {
	if n <= 1 {
		return []float64{1.0}
	}
	w := make([]float64, n)
	for i := 0; i < n; i++ {
		w[i] = 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(n-1)))
	}
	return w
}
