package consume

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

// Config holds configuration for Kafka consumer
type Config struct {
	Brokers       []string
	Topic         string
	ConsumerGroup string
	Offset        string // earliest, latest, or specific offset
	MaxMessages   int
	Timeout       time.Duration
	Format        string // json, raw, kv (key:value)
	ShowKey       bool
	ShowHeaders   bool
	ShowPartition bool
	ShowOffset    bool
	ShowTimestamp bool
	Verbose       bool
}

// DefaultConfig returns default consumer configuration
func DefaultConfig() Config {
	return Config{
		Brokers:       []string{"localhost:9092"},
		ConsumerGroup: "dimutils-consumer",
		Offset:        "latest",
		MaxMessages:   -1, // unlimited
		Timeout:       30 * time.Second,
		Format:        "raw",
		ShowKey:       false,
		ShowHeaders:   false,
		ShowPartition: false,
		ShowOffset:    false,
		ShowTimestamp: false,
		Verbose:       false,
	}
}

// MessageOutput represents a formatted message for output
type MessageOutput struct {
	Topic     string            `json:"topic,omitempty"`
	Partition int32             `json:"partition,omitempty"`
	Offset    int64             `json:"offset,omitempty"`
	Timestamp time.Time         `json:"timestamp,omitempty"`
	Key       string            `json:"key,omitempty"`
	Value     string            `json:"value"`
	Headers   map[string]string `json:"headers,omitempty"`
}

// Consumer represents a Kafka consumer
type Consumer struct {
	config Config
	client sarama.ConsumerGroup
	ready  chan bool
}

// Run is the main entry point for consume functionality
func Run(args []string) error {
	config := DefaultConfig()
	
	if err := parseArgs(args, &config); err != nil {
		return err
	}

	if config.Topic == "" {
		printHelp()
		return fmt.Errorf("topic is required")
	}

	return startConsumer(config)
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
		case "--group", "-g":
			if i+1 < len(args) {
				config.ConsumerGroup = args[i+1]
			}
		case "--offset", "-o":
			if i+1 < len(args) {
				config.Offset = args[i+1]
			}
		case "--max-messages", "-m":
			if i+1 < len(args) {
				if count, err := strconv.Atoi(args[i+1]); err == nil {
					config.MaxMessages = count
				}
			}
		case "--timeout":
			if i+1 < len(args) {
				if duration, err := time.ParseDuration(args[i+1]); err == nil {
					config.Timeout = duration
				}
			}
		case "--format", "-f":
			if i+1 < len(args) {
				config.Format = args[i+1]
			}
		case "--show-key", "-k":
			config.ShowKey = true
		case "--show-headers":
			config.ShowHeaders = true
		case "--show-partition", "-p":
			config.ShowPartition = true
		case "--show-offset":
			config.ShowOffset = true
		case "--show-timestamp":
			config.ShowTimestamp = true
		case "--verbose", "-v":
			config.Verbose = true
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
	help := `Usage: consume [options] <topic>

Consume messages from a Kafka topic and output to stdout.

Options:
  --brokers, -b BROKERS     Comma-separated list of brokers (default: localhost:9092)
  --topic, -t TOPIC         Topic to consume from
  --group, -g GROUP         Consumer group ID (default: dimutils-consumer)
  --offset, -o OFFSET       Start offset: earliest, latest, or number (default: latest)
  --max-messages, -m COUNT  Maximum messages to consume (default: unlimited)
  --timeout DURATION        Consumer timeout (default: 30s)
  --format, -f FORMAT       Output format: raw, json, kv (default: raw)
  --show-key, -k            Show message key
  --show-headers            Show message headers
  --show-partition, -p      Show partition number
  --show-offset             Show message offset
  --show-timestamp          Show message timestamp
  --verbose, -v             Verbose output
  -h, --help                Show this help message

Examples:
  consume my-topic
  consume --brokers broker1:9092,broker2:9092 --group my-group my-topic
  consume --format json --show-key --show-offset my-topic
  consume --offset earliest --max-messages 100 my-topic`

	fmt.Println(help)
	return nil
}

func startConsumer(config Config) error {
	if config.Verbose {
		log.Printf("Starting consumer for topic %s with group %s", config.Topic, config.ConsumerGroup)
	}

	// Create Sarama config
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaConfig.Consumer.Offsets.Initial = getOffsetMode(config.Offset)
	saramaConfig.Consumer.Group.Session.Timeout = config.Timeout
	saramaConfig.Consumer.Return.Errors = true

	// Create consumer group
	client, err := sarama.NewConsumerGroup(config.Brokers, config.ConsumerGroup, saramaConfig)
	if err != nil {
		return fmt.Errorf("error creating consumer group: %w", err)
	}
	defer client.Close()

	consumer := &Consumer{
		config: config,
		client: client,
		ready:  make(chan bool),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle consumer group errors
	go func() {
		for err := range client.Errors() {
			log.Printf("Consumer error: %v", err)
		}
	}()

	// Set up signal handling
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	// Start consuming
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := client.Consume(ctx, []string{config.Topic}, consumer); err != nil {
				log.Printf("Error from consumer: %v", err)
				return
			}
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready
	if config.Verbose {
		log.Println("Consumer started, waiting for messages...")
	}

	// Wait for termination signal
	select {
	case <-sigterm:
		if config.Verbose {
			log.Println("Terminating consumer...")
		}
	case <-ctx.Done():
	}

	cancel()
	wg.Wait()

	return nil
}

// Setup implements sarama.ConsumerGroupHandler
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	return nil
}

// Cleanup implements sarama.ConsumerGroupHandler
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim implements sarama.ConsumerGroupHandler
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	messageCount := 0

	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			if err := consumer.outputMessage(message); err != nil {
				log.Printf("Error outputting message: %v", err)
				continue
			}

			session.MarkMessage(message, "")
			messageCount++

			if consumer.config.MaxMessages > 0 && messageCount >= consumer.config.MaxMessages {
				if consumer.config.Verbose {
					log.Printf("Reached max messages limit (%d)", consumer.config.MaxMessages)
				}
				return nil
			}

		case <-session.Context().Done():
			return nil
		}
	}
}

func (consumer *Consumer) outputMessage(message *sarama.ConsumerMessage) error {
	switch consumer.config.Format {
	case "json":
		return consumer.outputJSON(message)
	case "kv":
		return consumer.outputKeyValue(message)
	default: // raw
		return consumer.outputRaw(message)
	}
}

func (consumer *Consumer) outputJSON(message *sarama.ConsumerMessage) error {
	output := MessageOutput{
		Value: string(message.Value),
	}

	if consumer.config.ShowKey && message.Key != nil {
		output.Key = string(message.Key)
	}

	if consumer.config.ShowPartition {
		output.Topic = message.Topic
		output.Partition = message.Partition
	}

	if consumer.config.ShowOffset {
		output.Offset = message.Offset
	}

	if consumer.config.ShowTimestamp {
		output.Timestamp = message.Timestamp
	}

	if consumer.config.ShowHeaders && len(message.Headers) > 0 {
		output.Headers = make(map[string]string)
		for _, header := range message.Headers {
			output.Headers[string(header.Key)] = string(header.Value)
		}
	}

	jsonData, err := json.Marshal(output)
	if err != nil {
		return err
	}

	fmt.Println(string(jsonData))
	return nil
}

func (consumer *Consumer) outputKeyValue(message *sarama.ConsumerMessage) error {
	var parts []string

	if consumer.config.ShowTimestamp {
		parts = append(parts, message.Timestamp.Format(time.RFC3339))
	}

	if consumer.config.ShowPartition {
		parts = append(parts, fmt.Sprintf("partition=%d", message.Partition))
	}

	if consumer.config.ShowOffset {
		parts = append(parts, fmt.Sprintf("offset=%d", message.Offset))
	}

	if consumer.config.ShowKey && message.Key != nil {
		parts = append(parts, fmt.Sprintf("key=%s", string(message.Key)))
	}

	parts = append(parts, fmt.Sprintf("value=%s", string(message.Value)))

	fmt.Println(strings.Join(parts, " "))
	return nil
}

func (consumer *Consumer) outputRaw(message *sarama.ConsumerMessage) error {
	var output strings.Builder

	if consumer.config.ShowKey && message.Key != nil {
		output.WriteString(string(message.Key))
		output.WriteString(":")
	}

	output.WriteString(string(message.Value))
	fmt.Println(output.String())
	return nil
}

func getOffsetMode(offset string) int64 {
	switch strings.ToLower(offset) {
	case "earliest":
		return sarama.OffsetOldest
	case "latest":
		return sarama.OffsetNewest
	default:
		// Try to parse as specific offset
		if offsetNum, err := strconv.ParseInt(offset, 10, 64); err == nil {
			return offsetNum
		}
		return sarama.OffsetNewest
	}
}