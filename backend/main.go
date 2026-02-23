package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("port", 8080, "server port")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/denoise", handleDenoise)

	handler := corsMiddleware(mux)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("noise cancellation server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
