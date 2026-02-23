package main

import (
	"math"
	"math/cmplx"
	"testing"
)

func TestFFTRoundtrip(t *testing.T) {
	// Generate a known signal: sum of two sinusoids.
	n := 1024
	input := make([]complex128, n)
	for i := 0; i < n; i++ {
		v := math.Sin(2*math.Pi*3*float64(i)/float64(n)) +
			0.5*math.Cos(2*math.Pi*7*float64(i)/float64(n))
		input[i] = complex(v, 0)
	}

	// Forward then inverse should recover original signal.
	spectrum := FFT(input)
	recovered := IFFT(spectrum)

	for i := 0; i < n; i++ {
		diff := cmplx.Abs(input[i] - recovered[i])
		if diff > 1e-9 {
			t.Fatalf("sample %d: expected %v, got %v (diff=%e)", i, input[i], recovered[i], diff)
		}
	}
}

func TestFFTParseval(t *testing.T) {
	// Parseval's theorem: sum(|x|^2) == (1/N) * sum(|X|^2)
	n := 512
	input := make([]complex128, n)
	for i := 0; i < n; i++ {
		input[i] = complex(math.Sin(2*math.Pi*float64(i)/float64(n)), 0)
	}

	spectrum := FFT(input)

	var timeEnergy, freqEnergy float64
	for i := 0; i < n; i++ {
		timeEnergy += cmplx.Abs(input[i]) * cmplx.Abs(input[i])
		freqEnergy += cmplx.Abs(spectrum[i]) * cmplx.Abs(spectrum[i])
	}
	freqEnergy /= float64(n)

	if math.Abs(timeEnergy-freqEnergy) > 1e-6 {
		t.Fatalf("Parseval violated: time=%f, freq=%f", timeEnergy, freqEnergy)
	}
}

func TestDenoiseReducesNoise(t *testing.T) {
	sampleRate := 44100
	duration := 2.0 // seconds
	n := int(duration * float64(sampleRate))

	// Generate pure white noise.
	samples := make([]float64, n)
	// Use a simple deterministic pseudo-noise (not rand, for reproducibility).
	state := uint32(12345)
	for i := range samples {
		state ^= state << 13
		state ^= state >> 17
		state ^= state << 5
		samples[i] = (float64(int32(state)) / float64(math.MaxInt32)) * 0.5
	}

	inputRMS := rms(samples)
	cleaned := Denoise(samples, sampleRate)
	outputRMS := rms(cleaned)

	// Noise should be significantly reduced.
	reduction := 20 * math.Log10(outputRMS/inputRMS)
	t.Logf("input RMS=%.6f, output RMS=%.6f, reduction=%.1f dB", inputRMS, outputRMS, reduction)

	if reduction > -3 {
		t.Fatalf("expected at least 3 dB noise reduction, got %.1f dB", reduction)
	}
}

func TestDenoisePreservesSignal(t *testing.T) {
	sampleRate := 44100
	n := sampleRate * 2 // 2 seconds

	samples := make([]float64, n)

	// First 0.5s: silence (noise estimation region).
	// Remaining 1.5s: 440 Hz tone.
	toneStart := sampleRate / 2
	for i := toneStart; i < n; i++ {
		samples[i] = 0.8 * math.Sin(2*math.Pi*440*float64(i)/float64(sampleRate))
	}

	cleaned := Denoise(samples, sampleRate)

	// Measure energy of the tone region in input and output.
	inputToneRMS := rms(samples[toneStart:])
	outputToneRMS := rms(cleaned[toneStart:])

	// The tone should retain most of its energy (within 6 dB).
	ratio := outputToneRMS / inputToneRMS
	t.Logf("tone input RMS=%.6f, output RMS=%.6f, ratio=%.3f", inputToneRMS, outputToneRMS, ratio)

	if ratio < 0.25 {
		t.Fatalf("tone was attenuated too much: ratio=%.3f", ratio)
	}
}

func TestWAVRoundtrip(t *testing.T) {
	samples := make([]float64, 1000)
	for i := range samples {
		samples[i] = math.Sin(2 * math.Pi * float64(i) / 100)
	}

	data := WriteWAV(samples, 44100)
	recovered, sr, err := ReadWAV(data)
	if err != nil {
		t.Fatalf("ReadWAV failed: %v", err)
	}
	if sr != 44100 {
		t.Fatalf("expected sample rate 44100, got %d", sr)
	}
	if len(recovered) != len(samples) {
		t.Fatalf("expected %d samples, got %d", len(samples), len(recovered))
	}

	// 16-bit quantization gives ~1/32768 precision.
	for i := range samples {
		diff := math.Abs(samples[i] - recovered[i])
		if diff > 0.001 {
			t.Fatalf("sample %d: expected %.6f, got %.6f (diff=%.6f)", i, samples[i], recovered[i], diff)
		}
	}
}

func TestFullPipeline(t *testing.T) {
	// Simulate exactly what the HTTP handler does: ReadWAV -> Denoise -> WriteWAV.
	sampleRate := 48000
	n := sampleRate * 3 // 3 seconds

	// Generate noisy speech: sine wave + noise.
	samples := make([]float64, n)
	state := uint32(99999)
	for i := range samples {
		state ^= state << 13
		state ^= state >> 17
		state ^= state << 5
		noise := (float64(int32(state)) / float64(math.MaxInt32)) * 0.1
		tone := 0.5 * math.Sin(2*math.Pi*440*float64(i)/float64(sampleRate))
		samples[i] = tone + noise
	}

	// Encode to WAV.
	wavBytes := WriteWAV(samples, sampleRate)

	// Decode WAV.
	decoded, sr, err := ReadWAV(wavBytes)
	if err != nil {
		t.Fatalf("ReadWAV: %v", err)
	}
	if sr != sampleRate {
		t.Fatalf("sample rate mismatch: %d vs %d", sr, sampleRate)
	}

	// Denoise.
	cleaned := Denoise(decoded, sr)
	if len(cleaned) != len(decoded) {
		t.Fatalf("length mismatch: input=%d, cleaned=%d", len(decoded), len(cleaned))
	}

	// Re-encode.
	outputWAV := WriteWAV(cleaned, sr)

	// Verify output is valid WAV.
	finalSamples, finalSR, err := ReadWAV(outputWAV)
	if err != nil {
		t.Fatalf("output ReadWAV: %v", err)
	}
	if finalSR != sampleRate {
		t.Fatalf("output sample rate mismatch: %d", finalSR)
	}
	if len(finalSamples) != len(cleaned) {
		t.Fatalf("output length mismatch: %d vs %d", len(finalSamples), len(cleaned))
	}

	t.Logf("pipeline OK: %d input samples -> %d bytes WAV -> %d decoded -> %d cleaned -> %d bytes output",
		len(samples), len(wavBytes), len(decoded), len(cleaned), len(outputWAV))
}
