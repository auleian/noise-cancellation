package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// WAVHeader holds metadata extracted from a WAV file.
type WAVHeader struct {
	SampleRate    int
	NumChannels   int
	BitsPerSample int
}

// ReadWAV parses a 16-bit PCM WAV file from raw bytes.
// Returns samples normalized to [-1.0, +1.0] and the sample rate.
// Stereo inputs are mixed down to mono by averaging left and right channels.
func ReadWAV(data []byte) ([]float64, int, error) {
	if len(data) < 12 {
		return nil, 0, errors.New("wav: file too short")
	}

	// Validate RIFF header.
	if string(data[0:4]) != "RIFF" {
		return nil, 0, errors.New("wav: missing RIFF header")
	}
	if string(data[8:12]) != "WAVE" {
		return nil, 0, errors.New("wav: missing WAVE identifier")
	}

	var header *WAVHeader
	var pcmData []byte

	// Walk through chunks.
	pos := 12
	for pos+8 <= len(data) {
		chunkID := string(data[pos : pos+4])
		chunkSize := int(binary.LittleEndian.Uint32(data[pos+4 : pos+8]))
		chunkStart := pos + 8

		switch chunkID {
		case "fmt ":
			if chunkSize < 16 {
				return nil, 0, errors.New("wav: fmt chunk too small")
			}
			if chunkStart+16 > len(data) {
				return nil, 0, errors.New("wav: fmt chunk truncated")
			}
			audioFormat := binary.LittleEndian.Uint16(data[chunkStart : chunkStart+2])
			if audioFormat != 1 {
				return nil, 0, fmt.Errorf("wav: unsupported audio format %d (only PCM/1 supported)", audioFormat)
			}
			header = &WAVHeader{
				NumChannels:   int(binary.LittleEndian.Uint16(data[chunkStart+2 : chunkStart+4])),
				SampleRate:    int(binary.LittleEndian.Uint32(data[chunkStart+4 : chunkStart+8])),
				BitsPerSample: int(binary.LittleEndian.Uint16(data[chunkStart+14 : chunkStart+16])),
			}
			if header.BitsPerSample != 16 {
				return nil, 0, fmt.Errorf("wav: unsupported bits per sample %d (only 16 supported)", header.BitsPerSample)
			}

		case "data":
			end := chunkStart + chunkSize
			if end > len(data) {
				end = len(data) // allow truncated data chunks
			}
			pcmData = data[chunkStart:end]
		}

		// Advance to next chunk (chunks are word-aligned).
		pos = chunkStart + chunkSize
		if chunkSize%2 != 0 {
			pos++ // padding byte
		}
	}

	if header == nil {
		return nil, 0, errors.New("wav: no fmt chunk found")
	}
	if pcmData == nil {
		return nil, 0, errors.New("wav: no data chunk found")
	}

	// Parse int16 samples.
	numSamples := len(pcmData) / 2
	rawSamples := make([]float64, numSamples)
	for i := 0; i < numSamples; i++ {
		s := int16(binary.LittleEndian.Uint16(pcmData[i*2 : i*2+2]))
		rawSamples[i] = float64(s) / 32768.0
	}

	// Mix to mono if stereo.
	if header.NumChannels == 2 {
		monoLen := numSamples / 2
		mono := make([]float64, monoLen)
		for i := 0; i < monoLen; i++ {
			mono[i] = (rawSamples[i*2] + rawSamples[i*2+1]) / 2.0
		}
		return mono, header.SampleRate, nil
	}

	return rawSamples, header.SampleRate, nil
}

// WriteWAV encodes mono float64 samples (in [-1.0, +1.0]) as a 16-bit PCM WAV file.
func WriteWAV(samples []float64, sampleRate int) []byte {
	numSamples := len(samples)
	dataSize := numSamples * 2 // 16-bit = 2 bytes per sample
	fileSize := 36 + dataSize  // total file size minus 8 bytes for RIFF header

	buf := &bytes.Buffer{}
	buf.Grow(44 + dataSize)

	// RIFF header.
	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, uint32(fileSize))
	buf.WriteString("WAVE")

	// fmt chunk.
	buf.WriteString("fmt ")
	binary.Write(buf, binary.LittleEndian, uint32(16)) // chunk size
	binary.Write(buf, binary.LittleEndian, uint16(1))  // PCM format
	binary.Write(buf, binary.LittleEndian, uint16(1))  // mono
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate*2)) // byte rate
	binary.Write(buf, binary.LittleEndian, uint16(2))            // block align
	binary.Write(buf, binary.LittleEndian, uint16(16))           // bits per sample

	// data chunk.
	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, uint32(dataSize))

	for _, s := range samples {
		// Clamp to [-1, 1].
		if s > 1.0 {
			s = 1.0
		} else if s < -1.0 {
			s = -1.0
		}
		// Convert to int16.
		var i16 int16
		if s >= 0 {
			i16 = int16(math.Round(s * 32767))
		} else {
			i16 = int16(math.Round(s * 32768))
		}
		binary.Write(buf, binary.LittleEndian, i16)
	}

	return buf.Bytes()
}
