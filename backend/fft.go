package main

import (
	"math"
	"math/cmplx"
)

// FFT computes the forward discrete Fourier transform using the
// iterative Cooley-Tukey radix-2 decimation-in-time algorithm.
// len(x) MUST be a power of 2; panics otherwise.
func FFT(x []complex128) []complex128 {
	n := len(x)
	if n == 0 {
		return nil
	}
	if !isPowerOf2(n) {
		panic("fft: length must be a power of 2")
	}

	// Copy input so we don't mutate the caller's slice.
	out := make([]complex128, n)
	copy(out, x)

	// Bit-reversal permutation.
	bitReverse(out)

	// Butterfly stages.
	for s := 1; s <= int(math.Log2(float64(n))); s++ {
		m := 1 << s                                          // butterfly span
		wm := cmplx.Exp(complex(0, -2*math.Pi/float64(m)))  // twiddle factor (negative for forward)

		for k := 0; k < n; k += m {
			w := complex(1, 0)
			for j := 0; j < m/2; j++ {
				t := w * out[k+j+m/2]
				u := out[k+j]
				out[k+j] = u + t
				out[k+j+m/2] = u - t
				w *= wm
			}
		}
	}

	return out
}

// IFFT computes the inverse discrete Fourier transform.
// Uses the conjugate-FFT-conjugate-scale identity:
//   IFFT(X) = conj(FFT(conj(X))) / N
// len(X) MUST be a power of 2; panics otherwise.
func IFFT(X []complex128) []complex128 {
	n := len(X)
	if n == 0 {
		return nil
	}

	conj := make([]complex128, n)
	for i, v := range X {
		conj[i] = cmplx.Conj(v)
	}

	result := FFT(conj)

	scale := complex(float64(n), 0)
	for i := range result {
		result[i] = cmplx.Conj(result[i]) / scale
	}

	return result
}

// NextPowerOf2 returns the smallest power of 2 that is >= n.
func NextPowerOf2(n int) int {
	if n <= 1 {
		return 1
	}
	p := 1
	for p < n {
		p <<= 1
	}
	return p
}

// isPowerOf2 reports whether n is a positive power of 2.
func isPowerOf2(n int) bool {
	return n > 0 && (n&(n-1)) == 0
}

// bitReverse reorders elements of x by bit-reversing their indices.
func bitReverse(x []complex128) {
	n := len(x)
	bits := int(math.Log2(float64(n)))

	for i := 0; i < n; i++ {
		j := reverseBits(i, bits)
		if j > i {
			x[i], x[j] = x[j], x[i]
		}
	}
}

// reverseBits reverses the lowest `bits` bits of v.
func reverseBits(v, bits int) int {
	r := 0
	for i := 0; i < bits; i++ {
		r = (r << 1) | (v & 1)
		v >>= 1
	}
	return r
}
