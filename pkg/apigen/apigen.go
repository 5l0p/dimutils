package apigen

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config holds configuration for API generation
type Config struct {
	InputFile      string
	OutputDir      string
	OutputFormat   string
	ServerPort     int
	RefreshRate    int
	EnableHTML     bool
	EnableMetaRefresh bool
}

// APIEndpoint represents a generated API endpoint
type APIEndpoint struct {
	Path     string
	Method   string
	Handler  http.HandlerFunc
	Data     interface{}
	Metadata EndpointMetadata
}

// EndpointMetadata holds metadata about an API endpoint
type EndpointMetadata struct {
	Description string
	ContentType string
	LastUpdated time.Time
	Schema      map[string]interface{}
}

// Generator manages API generation and serving
type Generator struct {
	Config    Config
	Endpoints []APIEndpoint
	Server    *http.Server
}

// DefaultConfig returns default API generator configuration
func DefaultConfig() Config {
	return Config{
		InputFile:         "-",
		OutputDir:         "./api",
		OutputFormat:      "json",
		ServerPort:        8080,
		RefreshRate:       30,
		EnableHTML:        true,
		EnableMetaRefresh: false,
	}
}

// Run executes the API generator
func Run(args []string) error {
	config := DefaultConfig()
	
	// Parse arguments
	for i, arg := range args {
		switch arg {
		case "--input", "-i":
			if i+1 < len(args) {
				config.InputFile = args[i+1]
			}
		case "--output", "-o":
			if i+1 < len(args) {
				config.OutputDir = args[i+1]
			}
		case "--format", "-f":
			if i+1 < len(args) {
				config.OutputFormat = args[i+1]
			}
		case "--port", "-p":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &config.ServerPort)
			}
		case "--refresh", "-r":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &config.RefreshRate)
			}
		case "--html":
			config.EnableHTML = true
		case "--meta-refresh":
			config.EnableMetaRefresh = true
		case "--serve":
			return serveAPI(config)
		case "--help", "-h":
			return showHelp()
		}
	}

	generator := NewGenerator(config)
	return generator.Generate()
}

func showHelp() error {
	fmt.Printf(`apigen - API Generator for read-only data APIs

Usage: apigen [options]

Options:
  -i, --input       Input data file or topic (default: stdin)
  -o, --output      Output directory for generated files (default: ./api)
  -f, --format      Output format (json, xml, csv) (default: json)
  -p, --port        Server port for API serving (default: 8080)
  -r, --refresh     Refresh rate in seconds for meta-refresh (default: 30)
      --html        Enable HTML page generation
      --meta-refresh Enable meta-refresh HTML pages
      --serve       Start API server instead of generating static files
  -h, --help        Show this help message

Examples:
  apigen -i data.json -o api --html
  apigen --serve -p 8080 --meta-refresh
  cat topic.json | apigen -f json --html
`)
	return nil
}

// NewGenerator creates a new API generator instance
func NewGenerator(config Config) *Generator {
	return &Generator{
		Config:    config,
		Endpoints: make([]APIEndpoint, 0),
	}
}

// Generate creates API files and HTML pages
func (g *Generator) Generate() error {
	// Read input data
	data, err := g.readInputData()
	if err != nil {
		return fmt.Errorf("failed to read input data: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(g.Config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate API endpoints
	if err := g.generateEndpoints(data); err != nil {
		return fmt.Errorf("failed to generate endpoints: %w", err)
	}

	// Generate HTML pages if enabled
	if g.Config.EnableHTML {
		if err := g.generateHTMLPages(data); err != nil {
			return fmt.Errorf("failed to generate HTML pages: %w", err)
		}
	}

	// Generate meta-refresh pages if enabled
	if g.Config.EnableMetaRefresh {
		if err := g.generateMetaRefreshPages(data); err != nil {
			return fmt.Errorf("failed to generate meta-refresh pages: %w", err)
		}
	}

	fmt.Printf("API generation completed. Files written to: %s\n", g.Config.OutputDir)
	return nil
}

func (g *Generator) readInputData() (interface{}, error) {
	var reader io.Reader
	
	if g.Config.InputFile == "-" {
		reader = os.Stdin
	} else {
		file, err := os.Open(g.Config.InputFile)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader = file
	}

	var data interface{}
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func (g *Generator) generateEndpoints(data interface{}) error {
	// Create main data endpoint
	endpoint := APIEndpoint{
		Path:   "/api/data",
		Method: "GET",
		Data:   data,
		Metadata: EndpointMetadata{
			Description: "Main data endpoint",
			ContentType: "application/json",
			LastUpdated: time.Now(),
		},
	}
	
	g.Endpoints = append(g.Endpoints, endpoint)

	// Write JSON data file
	jsonFile := filepath.Join(g.Config.OutputDir, "data.json")
	return g.writeJSONFile(jsonFile, data)
}

func (g *Generator) generateHTMLPages(data interface{}) error {
	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API Data Viewer</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .container { max-width: 1200px; margin: 0 auto; }
        .data-table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        .data-table th, .data-table td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        .data-table th { background-color: #f2f2f2; }
        .json-view { background: #f5f5f5; padding: 15px; border-radius: 5px; overflow-x: auto; }
        pre { margin: 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>API Data Viewer</h1>
        <p>Last updated: {{.LastUpdated}}</p>
        
        <h2>Raw JSON Data</h2>
        <div class="json-view">
            <pre>{{.JSONData}}</pre>
        </div>
        
        <h2>API Endpoints</h2>
        <ul>
            <li><a href="data.json">GET /api/data</a> - Main data endpoint</li>
        </ul>
    </div>
</body>
</html>`

	tmpl, err := template.New("index").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	templateData := struct {
		LastUpdated string
		JSONData    string
	}{
		LastUpdated: time.Now().Format(time.RFC3339),
		JSONData:    string(jsonData),
	}

	htmlFile := filepath.Join(g.Config.OutputDir, "index.html")
	file, err := os.Create(htmlFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, templateData)
}

func (g *Generator) generateMetaRefreshPages(data interface{}) error {
	refreshTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="refresh" content="{{.RefreshRate}}">
    <title>Live Data Feed</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 20px; 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            min-height: 100vh;
        }
        .live-indicator { 
            color: #28a745; 
            font-weight: bold;
            animation: pulse 2s infinite;
        }
        @keyframes pulse {
            0% { opacity: 1; }
            50% { opacity: 0.5; }
            100% { opacity: 1; }
        }
        .timestamp { color: #e0e0e0; font-size: 0.9em; }
        .data-container { 
            background: rgba(255,255,255,0.1); 
            padding: 20px; 
            border-radius: 10px; 
            box-shadow: 0 4px 8px rgba(0,0,0,0.2);
            backdrop-filter: blur(10px);
        }
        .json-view { 
            background: rgba(0,0,0,0.3); 
            padding: 15px; 
            border-radius: 5px; 
            overflow-x: auto;
            border-left: 4px solid #28a745;
        }
        .status-bar {
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            background: rgba(0,0,0,0.8);
            padding: 10px;
            text-align: center;
            z-index: 1000;
        }
        .metrics {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin: 20px 0;
        }
        .metric-card {
            background: rgba(255,255,255,0.1);
            padding: 15px;
            border-radius: 8px;
            text-align: center;
        }
        .metric-value {
            font-size: 2em;
            font-weight: bold;
            color: #28a745;
        }
        pre { 
            margin: 0; 
            white-space: pre-wrap; 
            font-family: 'Courier New', monospace;
            color: #f8f9fa;
        }
    </style>
</head>
<body>
    <div class="status-bar">
        🔴 <span class="live-indicator">LIVE</span> | Next refresh in <span id="countdown">{{.RefreshRate}}</span>s
    </div>

    <div style="margin-top: 60px;">
        <div class="data-container">
            <h1>📡 Live Data Feed</h1>
            <p class="timestamp">Last updated: {{.LastUpdated}}</p>
            <p>Auto-refresh every {{.RefreshRate}} seconds</p>
            
            <div class="metrics">
                <div class="metric-card">
                    <div class="metric-value" id="record-count">{{.RecordCount}}</div>
                    <div>Records</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value" id="update-time">{{.UpdateTime}}</div>
                    <div>Last Update</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value" id="status">●</div>
                    <div>Status</div>
                </div>
            </div>
            
            <div class="json-view">
                <pre>{{.JSONData}}</pre>
            </div>
        </div>
    </div>

    <script>
        let countdownTime = {{.RefreshRate}};
        
        function updateCountdown() {
            const countdownElement = document.getElementById('countdown');
            if (countdownElement) {
                countdownElement.textContent = countdownTime;
                countdownTime--;
                
                if (countdownTime < 0) {
                    countdownTime = {{.RefreshRate}};
                }
            }
        }
        
        // Update countdown every second
        setInterval(updateCountdown, 1000);
        
        // Add some visual feedback
        document.addEventListener('DOMContentLoaded', function() {
            const statusElement = document.getElementById('status');
            if (statusElement) {
                statusElement.style.color = '#28a745';
            }
        });
    </script>
</body>
</html>`

	tmpl, err := template.New("live").Parse(refreshTemplate)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	templateData := struct {
		RefreshRate int
		LastUpdated string
		JSONData    string
		RecordCount int
		UpdateTime  string
	}{
		RefreshRate: g.Config.RefreshRate,
		LastUpdated: time.Now().Format(time.RFC3339),
		JSONData:    string(jsonData),
		RecordCount: g.getRecordCountSimple(data),
		UpdateTime:  time.Now().Format("15:04:05"),
	}

	htmlFile := filepath.Join(g.Config.OutputDir, "live.html")
	file, err := os.Create(htmlFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, templateData)
}

func (g *Generator) writeJSONFile(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// serveAPI starts a web server to serve the generated API
func serveAPI(config Config) error {
	generator := NewGenerator(config)
	
	// Setup HTTP handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			generator.handleAPIRequest(w, r)
		} else {
			generator.handleStaticRequest(w, r)
		}
	})

	addr := fmt.Sprintf(":%d", config.ServerPort)
	fmt.Printf("Starting API server on http://localhost%s\n", addr)
	
	server := &http.Server{
		Addr:    addr,
		Handler: nil,
	}
	
	return server.ListenAndServe()
}

func (g *Generator) getRecordCountSimple(data interface{}) int {
	if slice, ok := data.([]interface{}); ok {
		return len(slice)
	}
	return 1
}

func (g *Generator) handleAPIRequest(w http.ResponseWriter, r *http.Request) {
	// Placeholder for API request handling
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	response := map[string]interface{}{
		"message":   "API endpoint not yet implemented",
		"path":      r.URL.Path,
		"method":    r.Method,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	json.NewEncoder(w).Encode(response)
}

func (g *Generator) handleStaticRequest(w http.ResponseWriter, r *http.Request) {
	// Serve static files from output directory
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}
	
	filePath := filepath.Join(g.Config.OutputDir, path)
	http.ServeFile(w, r, filePath)
}