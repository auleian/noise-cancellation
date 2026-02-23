package main

import (
	"io"
	"log"
	"net/http"
)

const maxUploadSize = 50 << 20 // 50 MB

// corsMiddleware adds CORS headers so the Vite dev server (or any origin)
// can make requests to this backend.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleDenoise handles POST /denoise.
// Expects a multipart form with a "file" field containing a WAV file.
// Returns the denoised audio as a WAV response.
func handleDenoise(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form.
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		log.Printf("denoise: failed to parse form: %v", err)
		http.Error(w, "failed to parse upload", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		log.Printf("denoise: no file in request: %v", err)
		http.Error(w, "no file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read the entire file into memory.
	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("denoise: failed to read file: %v", err)
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	// Decode WAV.
	samples, sampleRate, err := ReadWAV(data)
	if err != nil {
		log.Printf("denoise: invalid WAV: %v", err)
		http.Error(w, "invalid WAV file: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("denoise: received %d samples at %d Hz (%.2f seconds)",
		len(samples), sampleRate, float64(len(samples))/float64(sampleRate))

	// Run noise cancellation.
	cleaned := Denoise(samples, sampleRate)

	// Encode result as WAV.
	result := WriteWAV(cleaned, sampleRate)

	log.Printf("denoise: returning %d bytes of cleaned audio", len(result))

	// Send response.
	w.Header().Set("Content-Type", "audio/wav")
	w.Header().Set("Content-Disposition", "attachment; filename=\"cleaned.wav\"")
	w.Write(result)
}
