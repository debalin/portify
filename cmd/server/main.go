package main

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/debalin/portify/gen/go/converter/v1/converterv1connect"
	"github.com/debalin/portify/internal/adapters/mock"
	"github.com/debalin/portify/internal/domain"
	"github.com/debalin/portify/internal/server"
	"github.com/rs/cors"
)

func main() {
	mux := http.NewServeMux()

	// 0. Initialize the Provider Registry
	registry := domain.NewProviderRegistry()

	// Register Mock Providers for now
	registry.RegisterSource(&mock.MockSource{})
	registry.RegisterDestination(&mock.MockDestination{})

	// 1. Create our server logic
	converterHelper := server.NewConverterServer(registry)

	// 2. "Mount" the server onto the Mux (Router)
	// The generated code gives us a valid path and handler.
	path, handler := converterv1connect.NewConverterServiceHandler(converterHelper)
	mux.Handle(path, handler)

	log.Printf("Bount handler on path: %s", path)

	// 3. Setup CORS (Cross-Origin Resource Sharing)
	// This is required so our React Frontend (running on port 3000 or 5173) can talk to this Backend (port 8080).
	// For dev, we allow all origins.
	corsHandler := cors.AllowAll().Handler(mux)

	// 4. Start the Server
	// We use h2c to allow HTTP/2 over cleartext (no TLS) which is great for local dev.
	port := 8080
	addr := fmt.Sprintf("localhost:%d", port)
	log.Printf("Server listening on http://%s", addr)

	err := http.ListenAndServe(
		addr,
		h2c.NewHandler(corsHandler, &http2.Server{}),
	)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
