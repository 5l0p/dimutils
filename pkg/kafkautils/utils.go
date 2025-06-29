package kafkautils

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Mechanism string // PLAIN, SCRAM-SHA-256, SCRAM-SHA-512, GSSAPI
	Username  string
	Password  string
	SASLSSL   bool
	TLSConfig *tls.Config
}

// ConnectionConfig holds Kafka connection configuration
type ConnectionConfig struct {
	Brokers     []string
	Timeout     time.Duration
	RetryBackoff time.Duration
	Auth        *AuthConfig
	TLS         *TLSConfig
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled            bool
	InsecureSkipVerify bool
	CertFile           string
	KeyFile            string
	CAFile             string
}

// ConfigureAuthentication sets up SASL authentication
func ConfigureAuthentication(config *sarama.Config, auth *AuthConfig) error {
	if auth == nil {
		return nil
	}

	config.Net.SASL.Enable = true
	config.Net.SASL.User = auth.Username
	config.Net.SASL.Password = auth.Password

	switch auth.Mechanism {
	case "PLAIN":
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	case "SCRAM-SHA-256":
		config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
		config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &XDGSCRAMClient{HashGeneratorFcn: SHA256}
		}
	case "SCRAM-SHA-512":
		config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &XDGSCRAMClient{HashGeneratorFcn: SHA512}
		}
	case "GSSAPI":
		config.Net.SASL.Mechanism = sarama.SASLTypeGSSAPI
		// Additional GSSAPI configuration would go here
	default:
		return fmt.Errorf("unsupported SASL mechanism: %s", auth.Mechanism)
	}

	if auth.SASLSSL {
		config.Net.TLS.Enable = true
		if auth.TLSConfig != nil {
			config.Net.TLS.Config = auth.TLSConfig
		}
	}

	return nil
}

// ConfigureTLS sets up TLS configuration
func ConfigureTLS(config *sarama.Config, tlsConf *TLSConfig) error {
	if tlsConf == nil || !tlsConf.Enabled {
		return nil
	}

	config.Net.TLS.Enable = true

	if tlsConf.InsecureSkipVerify {
		config.Net.TLS.Config = &tls.Config{
			InsecureSkipVerify: true,
		}
		return nil
	}

	if tlsConf.CertFile != "" && tlsConf.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(tlsConf.CertFile, tlsConf.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to load client certificates: %w", err)
		}

		config.Net.TLS.Config = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	return nil
}

// CreateBaseConfig creates a base Sarama configuration with common settings
func CreateBaseConfig(connConfig ConnectionConfig) (*sarama.Config, error) {
	config := sarama.NewConfig()
	
	// Set version
	config.Version = sarama.V2_6_0_0
	
	// Set timeouts
	config.Net.DialTimeout = connConfig.Timeout
	config.Net.ReadTimeout = connConfig.Timeout
	config.Net.WriteTimeout = connConfig.Timeout
	
	// Set retry configuration
	config.Metadata.Retry.Backoff = connConfig.RetryBackoff
	config.Producer.Retry.Backoff = connConfig.RetryBackoff
	config.Consumer.Retry.Backoff = connConfig.RetryBackoff

	// Configure authentication
	if err := ConfigureAuthentication(config, connConfig.Auth); err != nil {
		return nil, fmt.Errorf("failed to configure authentication: %w", err)
	}

	// Configure TLS
	if err := ConfigureTLS(config, connConfig.TLS); err != nil {
		return nil, fmt.Errorf("failed to configure TLS: %w", err)
	}

	return config, nil
}

// DefaultConnectionConfig returns a default connection configuration
func DefaultConnectionConfig() ConnectionConfig {
	return ConnectionConfig{
		Brokers:      []string{"localhost:9092"},
		Timeout:      30 * time.Second,
		RetryBackoff: 250 * time.Millisecond,
	}
}

// ValidateTopicName validates a Kafka topic name
func ValidateTopicName(name string) error {
	if name == "" {
		return fmt.Errorf("topic name cannot be empty")
	}

	if len(name) > 249 {
		return fmt.Errorf("topic name cannot exceed 249 characters")
	}

	// Check for invalid characters
	for _, char := range name {
		if char == '/' || char == '\\' || char == ',' || char == '\u0000' ||
		   char == ':' || char == '"' || char == '\'' || char == ';' ||
		   char == '*' || char == '?' || char == ' ' || char == '\t' ||
		   char == '\n' || char == '\r' {
			return fmt.Errorf("topic name contains invalid character: %c", char)
		}
	}

	// Check for reserved names
	if name == "." || name == ".." {
		return fmt.Errorf("topic name cannot be '.' or '..'")
	}

	return nil
}

// FormatByteSize formats byte sizes in human readable format
func FormatByteSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration formats duration in human readable format
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}

// CalculateLag calculates consumer lag
func CalculateLag(highWaterMark, currentOffset int64) int64 {
	if highWaterMark >= currentOffset {
		return highWaterMark - currentOffset
	}
	return 0
}

// IsRetriableError checks if an error is retriable
func IsRetriableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific Kafka errors that are retriable
	switch err {
	case sarama.ErrOutOfBrokers,
		 sarama.ErrShuttingDown,
		 sarama.ErrControllerNotAvailable,
		 sarama.ErrLeaderNotAvailable,
		 sarama.ErrNetworkException,
		 sarama.ErrNotEnoughReplicas,
		 sarama.ErrNotEnoughReplicasAfterAppend,
		 sarama.ErrRequestTimedOut:
		return true
	}

	return false
}

// GetErrorMessage returns a user-friendly error message
func GetErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	switch err {
	case sarama.ErrNoTopicsToUpdateMetadata:
		return "No topics available to update metadata"
	case sarama.ErrOutOfBrokers:
		return "No brokers available - check your broker list and network connectivity"
	case sarama.ErrShuttingDown:
		return "Kafka client is shutting down"
	case sarama.ErrControllerNotAvailable:
		return "Kafka controller is not available"
	case sarama.ErrLeaderNotAvailable:
		return "Topic partition leader is not available"
	case sarama.ErrNetworkException:
		return "Network error occurred - check connectivity to Kafka brokers"
	case sarama.ErrNotEnoughReplicas:
		return "Not enough in-sync replicas available"
	case sarama.ErrRequestTimedOut:
		return "Request timed out - consider increasing timeout values"
	case sarama.ErrBrokerNotAvailable:
		return "Broker is not available"
	case sarama.ErrReplicaNotAvailable:
		return "Replica is not available"
	case sarama.ErrMessageTooLarge:
		return "Message is too large - check max.message.bytes configuration"
	case sarama.ErrInvalidMessage:
		return "Message format is invalid"
	case sarama.ErrOffsetOutOfRange:
		return "Requested offset is out of range"
	case sarama.ErrInvalidTopicException:
		return "Topic name is invalid"
	case sarama.ErrRecordListTooLarge:
		return "Record batch is too large"
	case sarama.ErrNotLeaderForPartition:
		return "Broker is not the leader for this partition"
	case sarama.ErrOffsetMetadataTooLarge:
		return "Offset metadata is too large"
	case sarama.ErrOffsetsLoadInProgress:
		return "Offset loading is in progress - please retry"
	case sarama.ErrConsumerCoordinatorNotAvailable:
		return "Consumer coordinator is not available"
	case sarama.ErrNotCoordinatorForConsumer:
		return "Broker is not the coordinator for this consumer group"
	default:
		return err.Error()
	}
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level   string // debug, info, warn, error
	Enabled bool
}

// SetupLogging configures Sarama logging
func SetupLogging(config LoggingConfig) {
	if !config.Enabled {
		sarama.Logger = &NoOpLogger{}
		return
	}

	// Use default Go logger for now
	// In production, you might want to integrate with structured logging
}

// NoOpLogger is a no-op logger that discards all log messages
type NoOpLogger struct{}

func (NoOpLogger) Print(v ...interface{})                 {}
func (NoOpLogger) Printf(format string, v ...interface{}) {}
func (NoOpLogger) Println(v ...interface{})               {}

// BrokerInfo represents information about a Kafka broker
type BrokerInfo struct {
	ID   int32
	Host string
	Port int32
}

// TopicInfo represents basic topic information
type TopicInfo struct {
	Name       string
	Partitions int32
	Replicas   int16
}

// ConsumerGroupInfo represents basic consumer group information
type ConsumerGroupInfo struct {
	GroupID      string
	State        string
	MemberCount  int
	Protocol     string
	ProtocolType string
}

// HealthCheck performs a basic health check on Kafka cluster
func HealthCheck(brokers []string, timeout time.Duration) error {
	config := sarama.NewConfig()
	config.Net.DialTimeout = timeout
	config.Version = sarama.V2_6_0_0

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer client.Close()

	// Try to get metadata
	_, err = client.RefreshMetadata()
	if err != nil {
		return fmt.Errorf("failed to refresh metadata: %w", err)
	}

	return nil
}

// GetBrokerInfo retrieves information about all brokers
func GetBrokerInfo(brokers []string) ([]BrokerInfo, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	var brokerInfos []BrokerInfo
	for _, broker := range client.Brokers() {
		brokerInfos = append(brokerInfos, BrokerInfo{
			ID:   broker.ID(),
			Host: broker.Addr(),
			Port: 0, // Port is included in Addr()
		})
	}

	return brokerInfos, nil
}