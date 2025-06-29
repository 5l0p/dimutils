package apigen

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// TopicTransformer handles transformation of topic data to blob storage
type TopicTransformer struct {
	Config TransformConfig
}

// TransformConfig holds configuration for topic-to-blob transformation
type TransformConfig struct {
	InputTopic   string
	OutputDir    string
	Format       string
	Compression  bool
	BatchSize    int
	ScheduleRate time.Duration
}

// BlobMetadata contains metadata about transformed blob data
type BlobMetadata struct {
	SourceTopic   string    `json:"source_topic"`
	CreatedAt     time.Time `json:"created_at"`
	RecordCount   int       `json:"record_count"`
	Format        string    `json:"format"`
	Size          int64     `json:"size_bytes"`
	Checksum      string    `json:"checksum"`
	Schema        string    `json:"schema,omitempty"`
}

// TopicRecord represents a single record from a topic
type TopicRecord struct {
	Key       string                 `json:"key"`
	Value     interface{}            `json:"value"`
	Timestamp time.Time              `json:"timestamp"`
	Offset    int64                  `json:"offset"`
	Partition int                    `json:"partition"`
	Headers   map[string]interface{} `json:"headers,omitempty"`
}

// DefaultTransformConfig returns default transformation configuration
func DefaultTransformConfig() TransformConfig {
	return TransformConfig{
		InputTopic:   "",
		OutputDir:    "./blobs",
		Format:       "json",
		Compression:  true,
		BatchSize:    1000,
		ScheduleRate: 1 * time.Hour,
	}
}

// NewTopicTransformer creates a new topic transformer
func NewTopicTransformer(config TransformConfig) *TopicTransformer {
	return &TopicTransformer{
		Config: config,
	}
}

// TransformToBlob converts topic data to blob storage format
func (tt *TopicTransformer) TransformToBlob(records []TopicRecord) error {
	if len(records) == 0 {
		return fmt.Errorf("no records to transform")
	}

	// Create output directory
	if err := os.MkdirAll(tt.Config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now()
	filename := tt.generateFilename(timestamp)
	
	switch strings.ToLower(tt.Config.Format) {
	case "json":
		return tt.writeJSONBlob(filename, records)
	case "csv":
		return tt.writeCSVBlob(filename, records)
	case "parquet":
		return tt.writeParquetBlob(filename, records)
	default:
		return fmt.Errorf("unsupported format: %s", tt.Config.Format)
	}
}

func (tt *TopicTransformer) generateFilename(timestamp time.Time) string {
	dateStr := timestamp.Format("2006-01-02")
	timeStr := timestamp.Format("15-04-05")
	
	filename := fmt.Sprintf("%s_%s_%s.%s", 
		tt.Config.InputTopic, dateStr, timeStr, tt.Config.Format)
	
	return filepath.Join(tt.Config.OutputDir, filename)
}

func (tt *TopicTransformer) writeJSONBlob(filename string, records []TopicRecord) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create blob structure
	blob := struct {
		Metadata BlobMetadata  `json:"metadata"`
		Records  []TopicRecord `json:"records"`
	}{
		Metadata: BlobMetadata{
			SourceTopic: tt.Config.InputTopic,
			CreatedAt:   time.Now(),
			RecordCount: len(records),
			Format:      "json",
		},
		Records: records,
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(blob); err != nil {
		return err
	}

	// Write metadata file
	return tt.writeMetadata(filename, blob.Metadata)
}

func (tt *TopicTransformer) writeCSVBlob(filename string, records []TopicRecord) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	headers := []string{"key", "value", "timestamp", "offset", "partition"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write records
	for _, record := range records {
		valueStr, _ := json.Marshal(record.Value)
		row := []string{
			record.Key,
			string(valueStr),
			record.Timestamp.Format(time.RFC3339),
			strconv.FormatInt(record.Offset, 10),
			strconv.Itoa(record.Partition),
		}
		
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	// Write metadata
	metadata := BlobMetadata{
		SourceTopic: tt.Config.InputTopic,
		CreatedAt:   time.Now(),
		RecordCount: len(records),
		Format:      "csv",
	}
	
	return tt.writeMetadata(filename, metadata)
}

func (tt *TopicTransformer) writeParquetBlob(filename string, records []TopicRecord) error {
	// Placeholder for Parquet implementation
	// Would require Apache Arrow or similar library
	return fmt.Errorf("parquet format not yet implemented")
}

func (tt *TopicTransformer) writeMetadata(dataFilename string, metadata BlobMetadata) error {
	// Get file stats
	if stat, err := os.Stat(dataFilename); err == nil {
		metadata.Size = stat.Size()
	}

	metadataFilename := strings.TrimSuffix(dataFilename, filepath.Ext(dataFilename)) + ".metadata.json"
	
	file, err := os.Create(metadataFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(metadata)
}

// GenerateAPIFromBlob creates API endpoints from existing blob files
func (tt *TopicTransformer) GenerateAPIFromBlob(blobDir string) error {
	// Scan blob directory for files
	entries, err := os.ReadDir(blobDir)
	if err != nil {
		return fmt.Errorf("failed to read blob directory: %w", err)
	}

	apiDir := filepath.Join(blobDir, "api")
	if err := os.MkdirAll(apiDir, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		if strings.Contains(entry.Name(), ".metadata.") {
			continue
		}

		// Create API endpoint for this blob
		if err := tt.createBlobEndpoint(blobDir, entry.Name(), apiDir); err != nil {
			fmt.Printf("Warning: failed to create endpoint for %s: %v\n", entry.Name(), err)
		}
	}

	// Generate index of all endpoints
	return tt.generateAPIIndex(apiDir)
}

func (tt *TopicTransformer) createBlobEndpoint(blobDir, filename, apiDir string) error {
	blobPath := filepath.Join(blobDir, filename)
	
	// Read blob data
	file, err := os.Open(blobPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var blob struct {
		Metadata BlobMetadata  `json:"metadata"`
		Records  []TopicRecord `json:"records"`
	}

	if err := json.NewDecoder(file).Decode(&blob); err != nil {
		return err
	}

	// Create endpoint directory structure
	endpointName := strings.TrimSuffix(filename, ".json")
	endpointDir := filepath.Join(apiDir, endpointName)
	if err := os.MkdirAll(endpointDir, 0755); err != nil {
		return err
	}

	// Write records endpoint
	recordsFile := filepath.Join(endpointDir, "records.json")
	if err := writeJSONFile(recordsFile, blob.Records); err != nil {
		return err
	}

	// Write metadata endpoint
	metadataFile := filepath.Join(endpointDir, "metadata.json")
	if err := writeJSONFile(metadataFile, blob.Metadata); err != nil {
		return err
	}

	// Write summary endpoint
	summary := map[string]interface{}{
		"name":         endpointName,
		"record_count": len(blob.Records),
		"created_at":   blob.Metadata.CreatedAt,
		"endpoints": map[string]string{
			"records":  fmt.Sprintf("/api/%s/records", endpointName),
			"metadata": fmt.Sprintf("/api/%s/metadata", endpointName),
		},
	}
	
	summaryFile := filepath.Join(endpointDir, "index.json")
	return writeJSONFile(summaryFile, summary)
}

func (tt *TopicTransformer) generateAPIIndex(apiDir string) error {
	entries, err := os.ReadDir(apiDir)
	if err != nil {
		return err
	}

	var endpoints []map[string]interface{}
	
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Read endpoint summary
		summaryPath := filepath.Join(apiDir, entry.Name(), "index.json")
		if summaryData, err := os.ReadFile(summaryPath); err == nil {
			var summary map[string]interface{}
			if json.Unmarshal(summaryData, &summary) == nil {
				endpoints = append(endpoints, summary)
			}
		}
	}

	index := map[string]interface{}{
		"api_version": "1.0",
		"generated_at": time.Now(),
		"endpoints": endpoints,
		"base_url": "/api",
	}

	indexFile := filepath.Join(apiDir, "index.json")
	return writeJSONFile(indexFile, index)
}

func writeJSONFile(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// ScheduledTransform runs transformation on a schedule
func (tt *TopicTransformer) ScheduledTransform(recordSource func() ([]TopicRecord, error)) error {
	ticker := time.NewTicker(tt.Config.ScheduleRate)
	defer ticker.Stop()

	fmt.Printf("Starting scheduled transformation every %v\n", tt.Config.ScheduleRate)

	for {
		select {
		case <-ticker.C:
			records, err := recordSource()
			if err != nil {
				fmt.Printf("Error fetching records: %v\n", err)
				continue
			}

			if len(records) > 0 {
				if err := tt.TransformToBlob(records); err != nil {
					fmt.Printf("Error transforming records: %v\n", err)
				} else {
					fmt.Printf("Transformed %d records to blob storage\n", len(records))
				}
			}
		}
	}
}