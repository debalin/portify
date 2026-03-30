package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/debalin/portify/gen/go/converter/v1/converterv1connect"
	"github.com/debalin/portify/internal/adapters/mock"
	"github.com/debalin/portify/internal/adapters/spotify"
	"github.com/debalin/portify/internal/adapters/youtube"
	"github.com/debalin/portify/internal/domain"
	"github.com/debalin/portify/internal/server"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	// Try loading .env from the current running directory
	// (usually project root if run via "go run ./cmd/server" from the root)
	_ = godotenv.Load(".env")

	mux := http.NewServeMux()

	// 0. Initialize the Provider Registry
	registry := domain.NewProviderRegistry()

	// Register Providers
	if os.Getenv("PORTIFY_MOCK_MODE") == "true" {
		log.Println("⚠️  WARNING: Starting server in MOCK MODE (PORTIFY_MOCK_MODE=true)")
		registry.RegisterSource(&mock.MockSourceWithTracks{})
		registry.RegisterDestination(&mock.MockDestination{})
	} else {
		registry.RegisterSource(spotify.NewAdapter())
		registry.RegisterDestination(youtube.NewAdapter())
	}

	// 1. Create our server logic
	converterHelper := server.NewConverterServer(registry)

	// 2. "Mount" the server onto the Mux (Router)
	// The generated code gives us a valid path and handler.
	path, handler := converterv1connect.NewConverterServiceHandler(converterHelper)
	mux.Handle(path, handler)

	// Add a simple healthcheck endpoint for testing/orchestration tools like Playwright
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

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
