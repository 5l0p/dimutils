package datagen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// ShadowConfig holds configuration for shadow traffic generation
type ShadowConfig struct {
	Targets     []ShadowTarget `json:"targets"`
	Rate        int            `json:"rate"`
	Duration    time.Duration  `json:"duration"`
	Concurrent  int            `json:"concurrent"`
	Template    *DataTemplate  `json:"template"`
	Delay       time.Duration  `json:"delay"`
	Jitter      time.Duration  `json:"jitter"`
	Timeout     time.Duration  `json:"timeout"`
	RetryCount  int            `json:"retry_count"`
	FailOnError bool           `json:"fail_on_error"`
}

// ShadowTarget defines where to send shadow traffic
type ShadowTarget struct {
	Name     string            `json:"name"`
	URL      string            `json:"url"`
	Method   string            `json:"method"`
	Headers  map[string]string `json:"headers"`
	Weight   int               `json:"weight"`
	Enabled  bool              `json:"enabled"`
	Validate bool              `json:"validate"`
}

// ShadowMetrics tracks shadow traffic performance
type ShadowMetrics struct {
	TotalRequests    int64         `json:"total_requests"`
	SuccessRequests  int64         `json:"success_requests"`
	FailedRequests   int64         `json:"failed_requests"`
	AverageLatency   time.Duration `json:"average_latency"`
	MinLatency       time.Duration `json:"min_latency"`
	MaxLatency       time.Duration `json:"max_latency"`
	ThroughputPerSec float64       `json:"throughput_per_sec"`
	ErrorRate        float64       `json:"error_rate"`
	StartTime        time.Time     `json:"start_time"`
	EndTime          time.Time     `json:"end_time"`
	TargetMetrics    map[string]*TargetMetrics `json:"target_metrics"`
}

// TargetMetrics tracks metrics per target
type TargetMetrics struct {
	Requests       int64         `json:"requests"`
	Successes      int64         `json:"successes"`
	Failures       int64         `json:"failures"`
	AverageLatency time.Duration `json:"average_latency"`
	LastError      string        `json:"last_error,omitempty"`
}

// ShadowGenerator manages shadow traffic generation
type ShadowGenerator struct {
	Config  ShadowConfig
	Metrics *ShadowMetrics
	Client  *http.Client
	mu      sync.RWMutex
}

// NewShadowGenerator creates a new shadow traffic generator
func NewShadowGenerator(config ShadowConfig) *ShadowGenerator {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.RetryCount == 0 {
		config.RetryCount = 3
	}

	return &ShadowGenerator{
		Config: config,
		Metrics: &ShadowMetrics{
			StartTime:     time.Now(),
			TargetMetrics: make(map[string]*TargetMetrics),
		},
		Client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Start begins shadow traffic generation
func (sg *ShadowGenerator) Start(ctx context.Context, generator *Generator) error {
	fmt.Printf("Starting shadow traffic generation:\n")
	fmt.Printf("  Targets: %d\n", len(sg.Config.Targets))
	fmt.Printf("  Rate: %d requests/sec\n", sg.Config.Rate)
	fmt.Printf("  Duration: %v\n", sg.Config.Duration)
	fmt.Printf("  Concurrent workers: %d\n", sg.Config.Concurrent)

	// Initialize target metrics
	for _, target := range sg.Config.Targets {
		if target.Enabled {
			sg.Metrics.TargetMetrics[target.Name] = &TargetMetrics{}
		}
	}

	// Create work channel
	workChan := make(chan ShadowRequest, sg.Config.Rate*2)
	
	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < sg.Config.Concurrent; i++ {
		wg.Add(1)
		go sg.worker(ctx, workChan, &wg, generator)
	}

	// Generate requests based on rate
	ticker := time.NewTicker(time.Second / time.Duration(sg.Config.Rate))
	defer ticker.Stop()

	timeout := time.After(sg.Config.Duration)
	requestCount := 0

	go sg.printProgress()

	for {
		select {
		case <-ticker.C:
			target := sg.selectTarget()
			if target != nil {
				request := ShadowRequest{
					Target:    *target,
					Data:      generator.generateRecord(sg.Config.Template),
					Timestamp: time.Now(),
					ID:        requestCount,
				}
				
				select {
				case workChan <- request:
					requestCount++
				default:
					// Channel full, skip this request
					fmt.Printf("Warning: Work channel full, skipping request\n")
				}
			}

		case <-timeout:
			close(workChan)
			wg.Wait()
			sg.Metrics.EndTime = time.Now()
			return sg.printFinalReport()

		case <-ctx.Done():
			close(workChan)
			wg.Wait()
			return ctx.Err()
		}
	}
}

// ShadowRequest represents a single shadow traffic request
type ShadowRequest struct {
	Target    ShadowTarget           `json:"target"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	ID        int                    `json:"id"`
}

func (sg *ShadowGenerator) worker(ctx context.Context, workChan <-chan ShadowRequest, wg *sync.WaitGroup, generator *Generator) {
	defer wg.Done()

	for {
		select {
		case request, ok := <-workChan:
			if !ok {
				return
			}
			sg.processRequest(request, generator)

		case <-ctx.Done():
			return
		}
	}
}

func (sg *ShadowGenerator) processRequest(request ShadowRequest, generator *Generator) {
	start := time.Now()
	
	// Add jitter delay
	if sg.Config.Jitter > 0 {
		jitter := time.Duration(generator.Rand.Int63n(int64(sg.Config.Jitter)))
		time.Sleep(jitter)
	}
	
	if sg.Config.Delay > 0 {
		time.Sleep(sg.Config.Delay)
	}

	success := false
	var lastError string

	// Retry logic
	for attempt := 0; attempt <= sg.Config.RetryCount; attempt++ {
		err := sg.sendRequest(request)
		if err == nil {
			success = true
			break
		}
		
		lastError = err.Error()
		if attempt < sg.Config.RetryCount {
			backoff := time.Duration(attempt+1) * 100 * time.Millisecond
			time.Sleep(backoff)
		}
	}

	latency := time.Since(start)
	
	// Update metrics
	sg.updateMetrics(request.Target.Name, success, latency, lastError)
}

func (sg *ShadowGenerator) sendRequest(request ShadowRequest) error {
	// Prepare request body
	var body []byte
	var err error
	
	if request.Target.Method != "GET" {
		body, err = json.Marshal(request.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal request data: %w", err)
		}
	}

	// Create HTTP request
	req, err := http.NewRequest(request.Target.Method, request.Target.URL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "dimutils-datagen-shadow/1.0")
	
	for key, value := range request.Target.Headers {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := sg.Client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Validate response if enabled
	if request.Target.Validate {
		if resp.StatusCode >= 400 {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}
	}

	return nil
}

func (sg *ShadowGenerator) selectTarget() *ShadowTarget {
	var enabledTargets []ShadowTarget
	totalWeight := 0

	for _, target := range sg.Config.Targets {
		if target.Enabled {
			enabledTargets = append(enabledTargets, target)
			weight := target.Weight
			if weight == 0 {
				weight = 1
			}
			totalWeight += weight
		}
	}

	if len(enabledTargets) == 0 {
		return nil
	}

	if totalWeight == 0 {
		// Equal distribution
		return &enabledTargets[0] // Simplified for now
	}

	// Weighted selection (simplified)
	return &enabledTargets[0]
}

func (sg *ShadowGenerator) updateMetrics(targetName string, success bool, latency time.Duration, lastError string) {
	sg.mu.Lock()
	defer sg.mu.Unlock()

	// Update overall metrics
	sg.Metrics.TotalRequests++
	if success {
		sg.Metrics.SuccessRequests++
	} else {
		sg.Metrics.FailedRequests++
	}

	// Update latency metrics
	if sg.Metrics.MinLatency == 0 || latency < sg.Metrics.MinLatency {
		sg.Metrics.MinLatency = latency
	}
	if latency > sg.Metrics.MaxLatency {
		sg.Metrics.MaxLatency = latency
	}

	// Simple moving average for latency
	if sg.Metrics.TotalRequests == 1 {
		sg.Metrics.AverageLatency = latency
	} else {
		sg.Metrics.AverageLatency = time.Duration(
			(int64(sg.Metrics.AverageLatency)*int64(sg.Metrics.TotalRequests-1) + int64(latency)) / int64(sg.Metrics.TotalRequests),
		)
	}

	// Update target-specific metrics
	targetMetrics, exists := sg.Metrics.TargetMetrics[targetName]
	if !exists {
		targetMetrics = &TargetMetrics{}
		sg.Metrics.TargetMetrics[targetName] = targetMetrics
	}

	targetMetrics.Requests++
	if success {
		targetMetrics.Successes++
	} else {
		targetMetrics.Failures++
		targetMetrics.LastError = lastError
	}

	// Update target average latency
	if targetMetrics.Requests == 1 {
		targetMetrics.AverageLatency = latency
	} else {
		targetMetrics.AverageLatency = time.Duration(
			(int64(targetMetrics.AverageLatency)*int64(targetMetrics.Requests-1) + int64(latency)) / int64(targetMetrics.Requests),
		)
	}
}

func (sg *ShadowGenerator) printProgress() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		sg.mu.RLock()
		elapsed := time.Since(sg.Metrics.StartTime)
		throughput := float64(sg.Metrics.TotalRequests) / elapsed.Seconds()
		errorRate := 0.0
		if sg.Metrics.TotalRequests > 0 {
			errorRate = float64(sg.Metrics.FailedRequests) / float64(sg.Metrics.TotalRequests) * 100
		}

		fmt.Printf("[%v] Requests: %d, Success: %d, Failed: %d, Throughput: %.1f req/s, Error Rate: %.1f%%\n",
			elapsed.Truncate(time.Second),
			sg.Metrics.TotalRequests,
			sg.Metrics.SuccessRequests,
			sg.Metrics.FailedRequests,
			throughput,
			errorRate,
		)
		sg.mu.RUnlock()
	}
}

func (sg *ShadowGenerator) printFinalReport() error {
	sg.mu.RLock()
	defer sg.mu.RUnlock()

	duration := sg.Metrics.EndTime.Sub(sg.Metrics.StartTime)
	throughput := float64(sg.Metrics.TotalRequests) / duration.Seconds()
	errorRate := 0.0
	if sg.Metrics.TotalRequests > 0 {
		errorRate = float64(sg.Metrics.FailedRequests) / float64(sg.Metrics.TotalRequests) * 100
	}

	fmt.Printf("\n=== Shadow Traffic Generation Report ===\n")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Total Requests: %d\n", sg.Metrics.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", sg.Metrics.SuccessRequests)
	fmt.Printf("Failed Requests: %d\n", sg.Metrics.FailedRequests)
	fmt.Printf("Throughput: %.2f requests/second\n", throughput)
	fmt.Printf("Error Rate: %.2f%%\n", errorRate)
	fmt.Printf("Average Latency: %v\n", sg.Metrics.AverageLatency)
	fmt.Printf("Min Latency: %v\n", sg.Metrics.MinLatency)
	fmt.Printf("Max Latency: %v\n", sg.Metrics.MaxLatency)

	fmt.Printf("\n=== Per-Target Metrics ===\n")
	for name, metrics := range sg.Metrics.TargetMetrics {
		targetErrorRate := 0.0
		if metrics.Requests > 0 {
			targetErrorRate = float64(metrics.Failures) / float64(metrics.Requests) * 100
		}

		fmt.Printf("Target: %s\n", name)
		fmt.Printf("  Requests: %d\n", metrics.Requests)
		fmt.Printf("  Success Rate: %.2f%%\n", 100-targetErrorRate)
		fmt.Printf("  Average Latency: %v\n", metrics.AverageLatency)
		if metrics.LastError != "" {
			fmt.Printf("  Last Error: %s\n", metrics.LastError)
		}
		fmt.Printf("\n")
	}

	return nil
}

// GetMetrics returns current metrics (thread-safe)
func (sg *ShadowGenerator) GetMetrics() ShadowMetrics {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	
	// Deep copy to avoid race conditions
	metrics := *sg.Metrics
	metrics.TargetMetrics = make(map[string]*TargetMetrics)
	for k, v := range sg.Metrics.TargetMetrics {
		targetMetric := *v
		metrics.TargetMetrics[k] = &targetMetric
	}
	
	return metrics
}

// DefaultShadowConfig returns a default shadow traffic configuration
func DefaultShadowConfig() ShadowConfig {
	return ShadowConfig{
		Targets: []ShadowTarget{
			{
				Name:     "local",
				URL:      "http://localhost:8080/api/test",
				Method:   "POST",
				Headers:  map[string]string{"X-Test": "shadow-traffic"},
				Weight:   1,
				Enabled:  true,
				Validate: true,
			},
		},
		Rate:        10,
		Duration:    1 * time.Minute,
		Concurrent:  4,
		Delay:       0,
		Jitter:      100 * time.Millisecond,
		Timeout:     10 * time.Second,
		RetryCount:  2,
		FailOnError: false,
	}
}