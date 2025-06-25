package serve

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Config holds configuration for the server
type Config struct {
	Port int
	Slow bool
	Dir  string
}

// DefaultConfig returns default server configuration
func DefaultConfig() Config {
	return Config{
		Port: 4321,
		Slow: false,
		Dir:  ".",
	}
}

// Run starts the HTTP server
func Run(args []string) error {
	config := DefaultConfig()
	
	// Parse arguments
	for i, arg := range args {
		switch arg {
		case "--slow":
			config.Slow = true
		case "--port", "-p":
			if i+1 < len(args) {
				if port, err := strconv.Atoi(args[i+1]); err == nil {
					config.Port = port
				}
			}
		case "--dir", "-d":
			if i+1 < len(args) {
				config.Dir = args[i+1]
			}
		case "--help", "-h":
			fmt.Println("Usage: serve [options]")
			fmt.Println("Options:")
			fmt.Println("  --slow       Simulate a slow server response")
			fmt.Println("  --port, -p  Specify the port to listen on (default: 4321)")
			fmt.Println("  --dir, -d   Specify the directory to serve (default: current directory)")
			fmt.Println("  --help, -h  Show this help message")
			return nil

		}

	}

	return startServer(config)
}

func startServer(config Config) error {
	// Create a custom logger
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Middleware function to log each page access
	logMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Printf("%s", r.URL.Path)
			if config.Slow {
				time.Sleep(2 * time.Second) // Simulate a slow server
			}
			next.ServeHTTP(w, r)
		})
	}

	// Use the middleware
	http.Handle("/", logMiddleware(http.FileServer(http.Dir(config.Dir))))

	// Start the server
	addr := fmt.Sprintf(":%d", config.Port)
	log.Printf("Server listening on port %d, serving directory: %s", config.Port, config.Dir)

	return http.ListenAndServe(addr, nil)
}