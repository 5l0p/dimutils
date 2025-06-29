package produce

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

// Config holds configuration for Kafka producer
type Config struct {
	Brokers         []string
	Topic           string
	Key             string
	KeyField        string // JSON field to use as message key
	Partition       int32
	Headers         map[string]string
	Async           bool
	BatchSize       int
	LingerMs        int
	Compression     string // none, gzip, snappy, lz4, zstd
	Acks            string // 0, 1, all
	Retries         int
	TimeoutMs       int
	Verbose         bool
	DryRun          bool
	InputFile       string
	MessageFormat   string // raw, json
	ValueField      string // JSON field to use as message value
}

// DefaultConfig returns default producer configuration
func DefaultConfig() Config {
	return Config{
		Brokers:       []string{"localhost:9092"},
		Partition:     -1, // let Kafka decide
		Async:         false,
		BatchSize:     16384,
		LingerMs:      0,
		Compression:   "none",
		Acks:          "1",
		Retries:       3,
		TimeoutMs:     10000,
		MessageFormat: "raw",
		Headers:       make(map[string]string),
	}
}

// MessageInput represents a message to be produced
type MessageInput struct {
	Key     string            `json:"key,omitempty"`
	Value   string            `json:"value"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Run is the main entry point for produce functionality
func Run(args []string) error {
	config := DefaultConfig()
	
	if err := parseArgs(args, &config); err != nil {
		return err
	}

	if config.Topic == "" {
		printHelp()
		return fmt.Errorf("topic is required")
	}

	return startProducer(config)
}

func parseArgs(args []string, config *Config) error {
	// Check for help first
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return printHelp()
		}
	}
	
	for i, arg := range args {
		switch arg {
		case "--brokers", "-b":
			if i+1 < len(args) {
				config.Brokers = strings.Split(args[i+1], ",")
			}
		case "--topic", "-t":
			if i+1 < len(args) {
				config.Topic = args[i+1]
			}
		case "--key", "-k":
			if i+1 < len(args) {
				config.Key = args[i+1]
			}
		case "--key-field":
			if i+1 < len(args) {
				config.KeyField = args[i+1]
			}
		case "--value-field":
			if i+1 < len(args) {
				config.ValueField = args[i+1]
			}
		case "--partition", "-p":
			if i+1 < len(args) {
				if partition, err := strconv.ParseInt(args[i+1], 10, 32); err == nil {
					config.Partition = int32(partition)
				}
			}
		case "--header", "-H":
			if i+1 < len(args) {
				parts := strings.SplitN(args[i+1], ":", 2)
				if len(parts) == 2 {
					config.Headers[parts[0]] = parts[1]
				}
			}
		case "--async", "-a":
			config.Async = true
		case "--batch-size":
			if i+1 < len(args) {
				if size, err := strconv.Atoi(args[i+1]); err == nil {
					config.BatchSize = size
				}
			}
		case "--linger-ms":
			if i+1 < len(args) {
				if linger, err := strconv.Atoi(args[i+1]); err == nil {
					config.LingerMs = linger
				}
			}
		case "--compression", "-c":
			if i+1 < len(args) {
				config.Compression = args[i+1]
			}
		case "--acks":
			if i+1 < len(args) {
				config.Acks = args[i+1]
			}
		case "--retries":
			if i+1 < len(args) {
				if retries, err := strconv.Atoi(args[i+1]); err == nil {
					config.Retries = retries
				}
			}
		case "--timeout":
			if i+1 < len(args) {
				if timeout, err := strconv.Atoi(args[i+1]); err == nil {
					config.TimeoutMs = timeout
				}
			}
		case "--input", "-i":
			if i+1 < len(args) {
				config.InputFile = args[i+1]
			}
		case "--format", "-f":
			if i+1 < len(args) {
				config.MessageFormat = args[i+1]
			}
		case "--verbose", "-v":
			config.Verbose = true
		case "--dry-run":
			config.DryRun = true
		case "-h", "--help":
			return printHelp()
		default:
			// If it doesn't start with -, treat as topic name
			if !strings.HasPrefix(arg, "-") && config.Topic == "" {
				config.Topic = arg
			}
		}
	}
	return nil
}

func printHelp() error {
	help := `Usage: produce [options] <topic>

Produce messages to a Kafka topic from stdin or file.

Options:
  --brokers, -b BROKERS     Comma-separated list of brokers (default: localhost:9092)
  --topic, -t TOPIC         Topic to produce to
  --key, -k KEY             Message key (same for all messages)
  --key-field FIELD         JSON field to use as message key
  --value-field FIELD       JSON field to use as message value (default: entire message)
  --partition, -p PARTITION Specific partition to send to (default: let Kafka decide)
  --header, -H KEY:VALUE    Add header to messages (can be used multiple times)
  --async, -a               Use async producer for better throughput
  --batch-size SIZE         Producer batch size in bytes (default: 16384)
  --linger-ms MS            Time to wait for batching (default: 0)
  --compression, -c TYPE    Compression: none, gzip, snappy, lz4, zstd (default: none)
  --acks LEVEL              Acknowledgment level: 0, 1, all (default: 1)
  --retries COUNT           Number of retries (default: 3)
  --timeout MS              Producer timeout in milliseconds (default: 10000)
  --input, -i FILE          Input file (default: stdin)
  --format, -f FORMAT       Input format: raw, json (default: raw)
  --verbose, -v             Verbose output
  --dry-run                 Show what would be sent without actually sending
  -h, --help                Show this help message

Examples:
  echo "hello world" | produce my-topic
  produce --brokers broker1:9092 --key mykey my-topic < messages.txt
  produce --format json --key-field id --value-field data my-topic < data.json
  produce --async --compression gzip --batch-size 32768 my-topic < large-file.txt`

	fmt.Println(help)
	return nil
}

func startProducer(config Config) error {
	if config.Verbose {
		log.Printf("Starting producer for topic %s", config.Topic)
	}

	// Skip producer creation in dry-run mode
	var producer sarama.SyncProducer
	var asyncProducer sarama.AsyncProducer
	var err error

	if !config.DryRun {
		// Create Sarama config
		saramaConfig := sarama.NewConfig()
		saramaConfig.Producer.Return.Successes = true
		saramaConfig.Producer.Return.Errors = true
		saramaConfig.Producer.Retry.Max = config.Retries
		saramaConfig.Producer.Timeout = time.Duration(config.TimeoutMs) * time.Millisecond

		// Set acknowledgment level
		switch config.Acks {
		case "0":
			saramaConfig.Producer.RequiredAcks = sarama.NoResponse
		case "1":
			saramaConfig.Producer.RequiredAcks = sarama.WaitForLocal
		case "all":
			saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
		}

		// Set compression
		switch config.Compression {
		case "gzip":
			saramaConfig.Producer.Compression = sarama.CompressionGZIP
		case "snappy":
			saramaConfig.Producer.Compression = sarama.CompressionSnappy
		case "lz4":
			saramaConfig.Producer.Compression = sarama.CompressionLZ4
		case "zstd":
			saramaConfig.Producer.Compression = sarama.CompressionZSTD
		default:
			saramaConfig.Producer.Compression = sarama.CompressionNone
		}

		// Set batch settings
		saramaConfig.Producer.Flush.Bytes = config.BatchSize
		saramaConfig.Producer.Flush.Frequency = time.Duration(config.LingerMs) * time.Millisecond

		// Create producer
		if config.Async {
			asyncProducer, err = sarama.NewAsyncProducer(config.Brokers, saramaConfig)
			if err != nil {
				return fmt.Errorf("error creating async producer: %w", err)
			}
			defer asyncProducer.Close()

			// Handle async responses
			go func() {
				for success := range asyncProducer.Successes() {
					if config.Verbose {
						log.Printf("Message sent to partition %d offset %d", success.Partition, success.Offset)
					}
				}
			}()

			go func() {
				for err := range asyncProducer.Errors() {
					log.Printf("Failed to send message: %v", err)
				}
			}()
		} else {
			producer, err = sarama.NewSyncProducer(config.Brokers, saramaConfig)
			if err != nil {
				return fmt.Errorf("error creating sync producer: %w", err)
			}
			defer producer.Close()
		}
	}

	// Set up input source
	var input *os.File
	if config.InputFile != "" {
		input, err = os.Open(config.InputFile)
		if err != nil {
			return fmt.Errorf("error opening input file: %w", err)
		}
		defer input.Close()
	} else {
		input = os.Stdin
	}

	// Process messages
	scanner := bufio.NewScanner(input)
	messageCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		message, err := prepareMessage(line, config)
		if err != nil {
			log.Printf("Error preparing message: %v", err)
			continue
		}

		if config.DryRun {
			var keyStr, valueStr string
			if message.Key != nil {
				keyBytes, _ := message.Key.Encode()
				keyStr = string(keyBytes)
			}
			if message.Value != nil {
				valueBytes, _ := message.Value.Encode()
				valueStr = string(valueBytes)
			}
			fmt.Printf("Would send: Topic=%s, Key=%s, Value=%s\n", 
				config.Topic, keyStr, valueStr)
			continue
		}

		if config.Async {
			asyncProducer.Input() <- message
		} else {
			partition, offset, err := producer.SendMessage(message)
			if err != nil {
				log.Printf("Failed to send message: %v", err)
				continue
			}
			if config.Verbose {
				log.Printf("Message sent to partition %d offset %d", partition, offset)
			}
		}

		messageCount++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	if config.Verbose {
		log.Printf("Sent %d messages", messageCount)
	}

	return nil
}

func prepareMessage(line string, config Config) (*sarama.ProducerMessage, error) {
	message := &sarama.ProducerMessage{
		Topic: config.Topic,
	}

	// Set partition if specified
	if config.Partition >= 0 {
		message.Partition = config.Partition
	}

	// Add headers
	for key, value := range config.Headers {
		message.Headers = append(message.Headers, sarama.RecordHeader{
			Key:   []byte(key),
			Value: []byte(value),
		})
	}

	// Process message based on format
	switch config.MessageFormat {
	case "json":
		return prepareJSONMessage(message, line, config)
	default:
		return prepareRawMessage(message, line, config)
	}
}

func prepareJSONMessage(message *sarama.ProducerMessage, line string, config Config) (*sarama.ProducerMessage, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Extract key from JSON field
	if config.KeyField != "" {
		if keyValue, exists := data[config.KeyField]; exists {
			if keyStr, ok := keyValue.(string); ok {
				message.Key = sarama.StringEncoder(keyStr)
			} else {
				// Convert non-string values to string
				keyBytes, _ := json.Marshal(keyValue)
				message.Key = sarama.StringEncoder(string(keyBytes))
			}
		}
	} else if config.Key != "" {
		message.Key = sarama.StringEncoder(config.Key)
	}

	// Extract value from JSON field or use entire message
	if config.ValueField != "" {
		if valueData, exists := data[config.ValueField]; exists {
			if valueStr, ok := valueData.(string); ok {
				message.Value = sarama.StringEncoder(valueStr)
			} else {
				// Convert non-string values to JSON
				valueBytes, _ := json.Marshal(valueData)
				message.Value = sarama.StringEncoder(string(valueBytes))
			}
		} else {
			return nil, fmt.Errorf("value field '%s' not found in JSON", config.ValueField)
		}
	} else {
		message.Value = sarama.StringEncoder(line)
	}

	return message, nil
}

func prepareRawMessage(message *sarama.ProducerMessage, line string, config Config) (*sarama.ProducerMessage, error) {
	// Use provided key or extract from line if it contains key:value format
	if config.Key != "" {
		message.Key = sarama.StringEncoder(config.Key)
		message.Value = sarama.StringEncoder(line)
	} else if strings.Contains(line, ":") && config.KeyField == "" {
		// Try to parse as key:value
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			message.Key = sarama.StringEncoder(strings.TrimSpace(parts[0]))
			message.Value = sarama.StringEncoder(strings.TrimSpace(parts[1]))
		} else {
			message.Value = sarama.StringEncoder(line)
		}
	} else {
		message.Value = sarama.StringEncoder(line)
	}

	return message, nil
}