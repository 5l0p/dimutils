package schemaregistry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client represents a Confluent Schema Registry client
type Client struct {
	baseURL    string
	httpClient *http.Client
	auth       *AuthConfig
}

// AuthConfig holds authentication configuration for Schema Registry
type AuthConfig struct {
	Username string
	Password string
}

// Schema represents a schema in the registry
type Schema struct {
	ID       int    `json:"id"`
	Version  int    `json:"version"`
	Schema   string `json:"schema"`
	Type     string `json:"schemaType,omitempty"`
	Subject  string `json:"subject,omitempty"`
}

// Subject represents a subject in the registry
type Subject struct {
	Name     string   `json:"subject"`
	Versions []int    `json:"versions"`
	Latest   *Schema  `json:"latest,omitempty"`
}

// CompatibilityLevel represents schema compatibility settings
type CompatibilityLevel struct {
	Compatibility string `json:"compatibility"`
}

// Config holds schema registry configuration
type Config struct {
	URL     string
	Timeout time.Duration
	Auth    *AuthConfig
}

// DefaultConfig returns default schema registry configuration
func DefaultConfig() Config {
	return Config{
		URL:     "http://localhost:8081",
		Timeout: 30 * time.Second,
	}
}

// NewClient creates a new Schema Registry client
func NewClient(config Config) *Client {
	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	return &Client{
		baseURL:    strings.TrimSuffix(config.URL, "/"),
		httpClient: httpClient,
		auth:       config.Auth,
	}
}

// GetSubjects returns all subjects in the registry
func (c *Client) GetSubjects() ([]string, error) {
	resp, err := c.makeRequest("GET", "/subjects", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get subjects: HTTP %d", resp.StatusCode)
	}

	var subjects []string
	if err := json.NewDecoder(resp.Body).Decode(&subjects); err != nil {
		return nil, fmt.Errorf("failed to decode subjects response: %w", err)
	}

	return subjects, nil
}

// GetSubjectVersions returns all versions for a subject
func (c *Client) GetSubjectVersions(subject string) ([]int, error) {
	path := fmt.Sprintf("/subjects/%s/versions", url.PathEscape(subject))
	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get subject versions: HTTP %d", resp.StatusCode)
	}

	var versions []int
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("failed to decode versions response: %w", err)
	}

	return versions, nil
}

// GetSchema returns a schema by subject and version
func (c *Client) GetSchema(subject string, version interface{}) (*Schema, error) {
	var versionStr string
	switch v := version.(type) {
	case int:
		versionStr = strconv.Itoa(v)
	case string:
		versionStr = v
	default:
		return nil, fmt.Errorf("version must be int or string")
	}

	path := fmt.Sprintf("/subjects/%s/versions/%s", 
		url.PathEscape(subject), url.PathEscape(versionStr))
	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get schema: HTTP %d", resp.StatusCode)
	}

	var schema Schema
	if err := json.NewDecoder(resp.Body).Decode(&schema); err != nil {
		return nil, fmt.Errorf("failed to decode schema response: %w", err)
	}

	schema.Subject = subject
	return &schema, nil
}

// GetSchemaByID returns a schema by its ID
func (c *Client) GetSchemaByID(id int) (*Schema, error) {
	path := fmt.Sprintf("/schemas/ids/%d", id)
	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get schema by ID: HTTP %d", resp.StatusCode)
	}

	var schema Schema
	if err := json.NewDecoder(resp.Body).Decode(&schema); err != nil {
		return nil, fmt.Errorf("failed to decode schema response: %w", err)
	}

	schema.ID = id
	return &schema, nil
}

// RegisterSchema registers a new schema for a subject
func (c *Client) RegisterSchema(subject, schema, schemaType string) (*Schema, error) {
	if schemaType == "" {
		schemaType = "AVRO"
	}

	payload := map[string]interface{}{
		"schema":     schema,
		"schemaType": schemaType,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	path := fmt.Sprintf("/subjects/%s/versions", url.PathEscape(subject))
	resp, err := c.makeRequest("POST", path, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to register schema: HTTP %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode registration response: %w", err)
	}

	id, ok := result["id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid response format: missing or invalid id")
	}

	return &Schema{
		ID:      int(id),
		Schema:  schema,
		Type:    schemaType,
		Subject: subject,
	}, nil
}

// DeleteSubject deletes a subject and all its versions
func (c *Client) DeleteSubject(subject string, permanent bool) ([]int, error) {
	path := fmt.Sprintf("/subjects/%s", url.PathEscape(subject))
	if permanent {
		path += "?permanent=true"
	}

	resp, err := c.makeRequest("DELETE", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to delete subject: HTTP %d", resp.StatusCode)
	}

	var versions []int
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("failed to decode delete response: %w", err)
	}

	return versions, nil
}

// DeleteSubjectVersion deletes a specific version of a subject
func (c *Client) DeleteSubjectVersion(subject string, version int, permanent bool) error {
	path := fmt.Sprintf("/subjects/%s/versions/%d", 
		url.PathEscape(subject), version)
	if permanent {
		path += "?permanent=true"
	}

	resp, err := c.makeRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete subject version: HTTP %d", resp.StatusCode)
	}

	return nil
}

// GetCompatibility returns the compatibility level for a subject
func (c *Client) GetCompatibility(subject string) (*CompatibilityLevel, error) {
	path := fmt.Sprintf("/config/%s", url.PathEscape(subject))
	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get compatibility: HTTP %d", resp.StatusCode)
	}

	var compat CompatibilityLevel
	if err := json.NewDecoder(resp.Body).Decode(&compat); err != nil {
		return nil, fmt.Errorf("failed to decode compatibility response: %w", err)
	}

	return &compat, nil
}

// SetCompatibility sets the compatibility level for a subject
func (c *Client) SetCompatibility(subject, level string) error {
	payload := map[string]string{
		"compatibility": level,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	path := fmt.Sprintf("/config/%s", url.PathEscape(subject))
	resp, err := c.makeRequest("PUT", path, bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to set compatibility: HTTP %d", resp.StatusCode)
	}

	return nil
}

// TestCompatibility tests if a schema is compatible with the latest version
func (c *Client) TestCompatibility(subject, schema, schemaType string) (bool, error) {
	if schemaType == "" {
		schemaType = "AVRO"
	}

	payload := map[string]interface{}{
		"schema":     schema,
		"schemaType": schemaType,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal payload: %w", err)
	}

	path := fmt.Sprintf("/compatibility/subjects/%s/versions/latest", 
		url.PathEscape(subject))
	resp, err := c.makeRequest("POST", path, bytes.NewReader(payloadBytes))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			if compatible, ok := result["is_compatible"].(bool); ok {
				return compatible, nil
			}
		}
	}

	return false, fmt.Errorf("failed to test compatibility: HTTP %d", resp.StatusCode)
}

// GetGlobalCompatibility returns the global compatibility level
func (c *Client) GetGlobalCompatibility() (*CompatibilityLevel, error) {
	resp, err := c.makeRequest("GET", "/config", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get global compatibility: HTTP %d", resp.StatusCode)
	}

	var compat CompatibilityLevel
	if err := json.NewDecoder(resp.Body).Decode(&compat); err != nil {
		return nil, fmt.Errorf("failed to decode compatibility response: %w", err)
	}

	return &compat, nil
}

// SetGlobalCompatibility sets the global compatibility level
func (c *Client) SetGlobalCompatibility(level string) error {
	payload := map[string]string{
		"compatibility": level,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := c.makeRequest("PUT", "/config", bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to set global compatibility: HTTP %d", resp.StatusCode)
	}

	return nil
}

// HealthCheck performs a health check on the schema registry
func (c *Client) HealthCheck() error {
	resp, err := c.makeRequest("GET", "/", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("schema registry health check failed: HTTP %d", resp.StatusCode)
	}

	return nil
}

// makeRequest makes an HTTP request to the schema registry
func (c *Client) makeRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/vnd.schemaregistry.v1+json")
	req.Header.Set("Accept", "application/vnd.schemaregistry.v1+json")

	// Add authentication if configured
	if c.auth != nil {
		req.SetBasicAuth(c.auth.Username, c.auth.Password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// ValidateSchemaType validates if the schema type is supported
func ValidateSchemaType(schemaType string) error {
	supportedTypes := []string{"AVRO", "JSON", "PROTOBUF"}
	
	for _, supported := range supportedTypes {
		if strings.EqualFold(schemaType, supported) {
			return nil
		}
	}
	
	return fmt.Errorf("unsupported schema type: %s. Supported types: %v", 
		schemaType, supportedTypes)
}

// ValidateCompatibilityLevel validates if the compatibility level is valid
func ValidateCompatibilityLevel(level string) error {
	validLevels := []string{
		"NONE", "BACKWARD", "BACKWARD_TRANSITIVE", 
		"FORWARD", "FORWARD_TRANSITIVE", "FULL", "FULL_TRANSITIVE",
	}
	
	for _, valid := range validLevels {
		if strings.EqualFold(level, valid) {
			return nil
		}
	}
	
	return fmt.Errorf("invalid compatibility level: %s. Valid levels: %v", 
		level, validLevels)
}