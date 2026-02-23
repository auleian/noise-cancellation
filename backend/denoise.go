package main

import (
	"math"
	"math/cmplx"
)

const (
	// FrameSize is the number of samples per FFT frame.
	// Must be a power of 2. At 44.1 kHz this is ~46 ms per frame,
	// giving 21.5 Hz frequency resolution — good for voice.
	FrameSize = 2048

	// HopSize is the step between consecutive frames.
	// 50% overlap with Hann window satisfies the COLA condition.
	HopSize = FrameSize / 2

	// NoiseFrames is the number of initial frames used to estimate
	// the noise profile. The beginning of the recording is assumed
	// to contain only background noise / silence.
	// 10 frames * 1024 hop ≈ 230 ms at 44.1 kHz.
	NoiseFrames = 10

	// SpectralFloor prevents magnitude bins from being driven to zero,
	// which would cause "musical noise" (isolated tonal artifacts).
	// Each bin retains at least this fraction of its original magnitude.
	SpectralFloor = 0.02

	// OverSubtract is the over-subtraction factor (alpha).
	// Subtracting more than the estimated noise compensates for
	// estimation variance. Typical range: 1.0–4.0.
	OverSubtract = 2.0
)

// Denoise performs spectral-subtraction noise cancellation on mono audio samples.
// samples should be normalized to [-1.0, +1.0]. sampleRate is preserved for
// potential future use but the algorithm is rate-independent.
func Denoise(samples []float64, sampleRate int) []float64 {
	n := len(samples)
	if n == 0 {
		return nil
	}

	// If the audio is shorter than one frame, zero-pad it.
	if n < FrameSize {
		padded := make([]float64, FrameSize)
		copy(padded, samples)
		samples = padded
		n = FrameSize
	}

	// How many frames fit?
	totalFrames := (n-FrameSize)/HopSize + 1
	if totalFrames < 1 {
		totalFrames = 1
	}

	// Cap noise frames to available frames.
	noiseFrames := NoiseFrames
	if noiseFrames > totalFrames {
		noiseFrames = totalFrames
	}

	// Generate window once.
	window := HannWindow(FrameSize)

	// ---------------------------------------------------------------
	// Step 1: Estimate noise magnitude spectrum from initial frames.
	// ---------------------------------------------------------------
	noiseMag := make([]float64, FrameSize)

	for fi := 0; fi < noiseFrames; fi++ {
		start := fi * HopSize
		frame := extractFrame(samples, start, FrameSize)
		applyWindow(frame, window)

		cx := realToComplex(frame)
		spectrum := FFT(cx)

		for k := 0; k < FrameSize; k++ {
			noiseMag[k] += cmplx.Abs(spectrum[k])
		}
	}

	// Average.
	for k := range noiseMag {
		noiseMag[k] /= float64(noiseFrames)
	}

	// ---------------------------------------------------------------
	// Step 2: Process every frame via spectral subtraction.
	// ---------------------------------------------------------------
	output := make([]float64, n)
	windowSum := make([]float64, n) // for overlap-add normalization

	for fi := 0; fi < totalFrames; fi++ {
		start := fi * HopSize

		// Extract and window the frame.
		frame := extractFrame(samples, start, FrameSize)
		applyWindow(frame, window)

		// Forward FFT.
		cx := realToComplex(frame)
		spectrum := FFT(cx)

		// Spectral subtraction.
		for k := 0; k < FrameSize; k++ {
			mag := cmplx.Abs(spectrum[k])
			phase := cmplx.Phase(spectrum[k])

			// Subtract over-estimated noise.
			cleanMag := mag - OverSubtract*noiseMag[k]

			// Gain floor: keep at least SpectralFloor * original magnitude.
			floor := SpectralFloor * mag
			if cleanMag < floor {
				cleanMag = floor
			}

			// Reconstruct with original phase.
			spectrum[k] = cmplx.Rect(cleanMag, phase)
		}

		// Inverse FFT.
		cleaned := IFFT(spectrum)

		// Overlap-add with synthesis window.
		for j := 0; j < FrameSize; j++ {
			idx := start + j
			if idx < n {
				output[idx] += real(cleaned[j]) * window[j]
				windowSum[idx] += window[j] * window[j]
			}
		}
	}

	// ---------------------------------------------------------------
	// Step 3: Normalize by the accumulated window energy.
	// ---------------------------------------------------------------
	for i := 0; i < n; i++ {
		if windowSum[i] > 1e-8 {
			output[i] /= windowSum[i]
		}
	}

	// ---------------------------------------------------------------
	// Step 4: Peak normalization — scale so the loudest sample hits
	// the target level, maximizing voice volume without clipping.
	// ---------------------------------------------------------------
	normalize(output, 0.95)

	return output
}

// extractFrame copies FrameSize samples starting at `start` from src.
// If the frame extends past the end of src, the remainder is zero-padded.
func extractFrame(src []float64, start, size int) []float64 {
	frame := make([]float64, size)
	end := start + size
	if end > len(src) {
		end = len(src)
	}
	copy(frame, src[start:end])
	return frame
}

// applyWindow multiplies each element of frame by the corresponding window value.
func applyWindow(frame, window []float64) {
	for i := range frame {
		frame[i] *= window[i]
	}
}

// realToComplex converts a float64 slice to complex128 (imaginary part = 0).
func realToComplex(x []float64) []complex128 {
	cx := make([]complex128, len(x))
	for i, v := range x {
		cx[i] = complex(v, 0)
	}
	return cx
}

// magnitude returns the magnitude spectrum of a complex slice.
func magnitude(x []complex128) []float64 {
	m := make([]float64, len(x))
	for i, v := range x {
		m[i] = cmplx.Abs(v)
	}
	return m
}

// normalize scales samples so the peak amplitude equals targetLevel.
// If the signal is silent (all zeros), it does nothing.
func normalize(samples []float64, targetLevel float64) {
	// Find peak absolute value.
	var peak float64
	for _, s := range samples {
		a := math.Abs(s)
		if a > peak {
			peak = a
		}
	}

	if peak < 1e-10 {
		return // silence — nothing to amplify
	}

	gain := targetLevel / peak
	for i := range samples {
		samples[i] *= gain
	}
}

// rms returns the root mean square of a float64 slice.
func rms(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	var sum float64
	for _, v := range x {
		sum += v * v
	}
	return math.Sqrt(sum / float64(len(x)))
}
