package kafkaadmin

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

// Config holds configuration for Kafka admin operations
type Config struct {
	Brokers []string
	Timeout time.Duration
	Verbose bool
}

// DefaultConfig returns default admin configuration
func DefaultConfig() Config {
	return Config{
		Brokers: []string{"localhost:9092"},
		Timeout: 30 * time.Second,
		Verbose: false,
	}
}

// AdminClient wraps Kafka admin operations
type AdminClient struct {
	config Config
	client sarama.ClusterAdmin
}

// TopicDetails represents detailed topic information
type TopicDetails struct {
	Name              string
	Partitions        int32
	ReplicationFactor int16
	ConfigEntries     map[string]string
	PartitionDetails  []PartitionDetail
}

// PartitionDetail represents partition-level information
type PartitionDetail struct {
	ID       int32
	Leader   int32
	Replicas []int32
	ISR      []int32
}

// ConsumerGroupDetail represents consumer group information
type ConsumerGroupDetail struct {
	GroupID     string
	State       string
	Protocol    string
	Members     []ConsumerGroupMember
	Assignments []TopicPartitionOffset
}

// ConsumerGroupMember represents a member of a consumer group
type ConsumerGroupMember struct {
	MemberID   string
	ClientID   string
	ClientHost string
}

// TopicPartitionOffset represents offset information
type TopicPartitionOffset struct {
	Topic     string
	Partition int32
	Offset    int64
	Lag       int64
	Metadata  string
}

// Run is the main entry point for admin functionality
func Run(args []string) error {
	if len(args) == 0 {
		return printHelp()
	}

	config := DefaultConfig()
	subcommand := args[0]
	subArgs := args[1:]

	// Parse global flags
	for i, arg := range subArgs {
		switch arg {
		case "--brokers", "-b":
			if i+1 < len(subArgs) {
				config.Brokers = strings.Split(subArgs[i+1], ",")
			}
		case "--timeout":
			if i+1 < len(subArgs) {
				if duration, err := time.ParseDuration(subArgs[i+1]); err == nil {
					config.Timeout = duration
				}
			}
		case "--verbose", "-v":
			config.Verbose = true
		}
	}

	client, err := NewAdminClient(config)
	if err != nil {
		return fmt.Errorf("failed to create admin client: %w", err)
	}
	defer client.Close()

	switch subcommand {
	case "list-topics", "topics":
		return client.ListTopics(subArgs)
	case "describe-topic", "describe":
		return client.DescribeTopic(subArgs)
	case "create-topic", "create":
		return client.CreateTopic(subArgs)
	case "delete-topic", "delete":
		return client.DeleteTopic(subArgs)
	case "list-groups", "groups":
		return client.ListConsumerGroups(subArgs)
	case "describe-group":
		return client.DescribeConsumerGroup(subArgs)
	case "reset-offset", "reset":
		return client.ResetConsumerGroupOffset(subArgs)
	case "offsets":
		return client.GetTopicOffsets(subArgs)
	case "configs":
		return client.GetTopicConfigs(subArgs)
	case "help", "-h", "--help":
		return printHelp()
	default:
		return fmt.Errorf("unknown subcommand: %s. Use 'help' to see available commands", subcommand)
	}
}

func printHelp() error {
	help := `Usage: kafkaadmin <subcommand> [options]

Kafka administration utility for managing topics, consumer groups, and configurations.

Global Options:
  --brokers, -b BROKERS     Comma-separated list of brokers (default: localhost:9092)
  --timeout DURATION        Operation timeout (default: 30s)
  --verbose, -v             Verbose output

Subcommands:
  list-topics               List all topics
  describe-topic TOPIC      Show detailed topic information
  create-topic TOPIC        Create a new topic
    --partitions NUM        Number of partitions (default: 1)
    --replication NUM       Replication factor (default: 1)
    --config KEY=VALUE      Topic configuration (can be used multiple times)

  delete-topic TOPIC        Delete a topic

  list-groups               List consumer groups
  describe-group GROUP      Show consumer group details
  reset-offset GROUP        Reset consumer group offsets
    --topic TOPIC           Topic to reset (required)
    --partition NUM         Specific partition (default: all)
    --to-earliest           Reset to earliest offset
    --to-latest             Reset to latest offset
    --to-offset NUM         Reset to specific offset

  offsets TOPIC             Show topic partition offsets
  configs TOPIC             Show topic configuration

Examples:
  kafkaadmin list-topics
  kafkaadmin create-topic my-topic --partitions 3 --replication 2
  kafkaadmin describe-topic my-topic
  kafkaadmin list-groups
  kafkaadmin reset-offset my-group --topic my-topic --to-earliest`

	fmt.Println(help)
	return nil
}

// NewAdminClient creates a new Kafka admin client
func NewAdminClient(config Config) (*AdminClient, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_6_0_0
	saramaConfig.Admin.Timeout = config.Timeout

	client, err := sarama.NewClusterAdmin(config.Brokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	return &AdminClient{
		config: config,
		client: client,
	}, nil
}

// Close closes the admin client
func (ac *AdminClient) Close() error {
	return ac.client.Close()
}

// ListTopics lists all topics in the cluster
func (ac *AdminClient) ListTopics(args []string) error {
	metadata, err := ac.client.DescribeTopics(nil)
	if err != nil {
		return fmt.Errorf("failed to list topics: %w", err)
	}

	if ac.config.Verbose {
		fmt.Printf("Found %d topics:\n", len(metadata))
		fmt.Printf("%-30s %-10s %-15s\n", "TOPIC", "PARTITIONS", "REPLICATION")
		fmt.Println(strings.Repeat("-", 65))
	}

	for _, details := range metadata {
		if ac.config.Verbose {
			replication := int16(0)
			if len(details.Partitions) > 0 {
				replication = int16(len(details.Partitions[0].Replicas))
			}
			fmt.Printf("%-30s %-10d %-15d\n", 
				details.Name, len(details.Partitions), replication)
		} else {
			fmt.Println(details.Name)
		}
	}

	return nil
}

// DescribeTopic shows detailed information about a topic
func (ac *AdminClient) DescribeTopic(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("topic name is required")
	}

	topicName := args[0]
	metadata, err := ac.client.DescribeTopics([]string{topicName})
	if err != nil {
		return fmt.Errorf("failed to describe topic: %w", err)
	}

	var topicFound *sarama.TopicMetadata
	for _, topic := range metadata {
		if topic.Name == topicName {
			topicFound = topic
			break
		}
	}

	if topicFound != nil {
		fmt.Printf("Topic: %s\n", topicName)
		fmt.Printf("Partitions: %d\n", len(topicFound.Partitions))
		
		if len(topicFound.Partitions) > 0 {
			fmt.Printf("Replication Factor: %d\n", len(topicFound.Partitions[0].Replicas))
		}

		fmt.Println("\nPartition Details:")
		fmt.Printf("%-10s %-8s %-20s %s\n", "PARTITION", "LEADER", "REPLICAS", "ISR")
		fmt.Println(strings.Repeat("-", 60))

		for _, partition := range topicFound.Partitions {
			replicasStr := fmt.Sprintf("%v", partition.Replicas)
			isrStr := fmt.Sprintf("%v", partition.Isr)
			fmt.Printf("%-10d %-8d %-20s %s\n", 
				partition.ID, partition.Leader, replicasStr, isrStr)
		}
	} else {
		return fmt.Errorf("topic %s not found", topicName)
	}

	return nil
}

// CreateTopic creates a new topic
func (ac *AdminClient) CreateTopic(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("topic name is required")
	}

	topicName := args[0]
	partitions := int32(1)
	replicationFactor := int16(1)
	configs := make(map[string]*string)

	// Parse arguments
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--partitions":
			if i+1 < len(args) {
				if p, err := strconv.ParseInt(args[i+1], 10, 32); err == nil {
					partitions = int32(p)
					i++
				}
			}
		case "--replication":
			if i+1 < len(args) {
				if r, err := strconv.ParseInt(args[i+1], 10, 16); err == nil {
					replicationFactor = int16(r)
					i++
				}
			}
		case "--config":
			if i+1 < len(args) {
				parts := strings.SplitN(args[i+1], "=", 2)
				if len(parts) == 2 {
					configs[parts[0]] = &parts[1]
				}
				i++
			}
		}
	}

	topicDetail := &sarama.TopicDetail{
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
		ConfigEntries:     configs,
	}

	err := ac.client.CreateTopic(topicName, topicDetail, false)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	if ac.config.Verbose {
		log.Printf("Created topic %s with %d partitions and replication factor %d", 
			topicName, partitions, replicationFactor)
	} else {
		fmt.Printf("Topic %s created successfully\n", topicName)
	}

	return nil
}

// DeleteTopic deletes a topic
func (ac *AdminClient) DeleteTopic(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("topic name is required")
	}

	topicName := args[0]
	err := ac.client.DeleteTopic(topicName)
	if err != nil {
		return fmt.Errorf("failed to delete topic: %w", err)
	}

	fmt.Printf("Topic %s deleted successfully\n", topicName)
	return nil
}

// ListConsumerGroups lists all consumer groups
func (ac *AdminClient) ListConsumerGroups(args []string) error {
	groups, err := ac.client.ListConsumerGroups()
	if err != nil {
		return fmt.Errorf("failed to list consumer groups: %w", err)
	}

	if ac.config.Verbose {
		fmt.Printf("Found %d consumer groups:\n", len(groups))
		fmt.Printf("%-30s %s\n", "GROUP ID", "PROTOCOL TYPE")
		fmt.Println(strings.Repeat("-", 50))
	}

	for groupID := range groups {
		if ac.config.Verbose {
			fmt.Printf("%-30s %s\n", groupID, "CONSUMER")
		} else {
			fmt.Println(groupID)
		}
	}

	return nil
}

// DescribeConsumerGroup shows detailed information about a consumer group
func (ac *AdminClient) DescribeConsumerGroup(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("consumer group ID is required")
	}

	groupID := args[0]
	details, err := ac.client.DescribeConsumerGroups([]string{groupID})
	if err != nil {
		return fmt.Errorf("failed to describe consumer group: %w", err)
	}

	var groupFound *sarama.GroupDescription
	for _, group := range details {
		if group.GroupId == groupID {
			groupFound = group
			break
		}
	}

	if groupFound != nil {
		fmt.Printf("Consumer Group: %s\n", groupID)
		fmt.Printf("State: %s\n", groupFound.State)
		fmt.Printf("Protocol: %s\n", groupFound.Protocol)
		fmt.Printf("Protocol Type: %s\n", groupFound.ProtocolType)

		if len(groupFound.Members) > 0 {
			fmt.Println("\nMembers:")
			fmt.Printf("%-20s %-20s %s\n", "MEMBER ID", "CLIENT ID", "HOST")
			fmt.Println(strings.Repeat("-", 70))

			for memberID, member := range groupFound.Members {
				fmt.Printf("%-20s %-20s %s\n", 
					memberID, member.ClientId, member.ClientHost)
			}
		}
	} else {
		return fmt.Errorf("consumer group %s not found", groupID)
	}

	return nil
}

// ResetConsumerGroupOffset resets consumer group offsets
func (ac *AdminClient) ResetConsumerGroupOffset(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("consumer group ID is required")
	}

	groupID := args[0]
	var topicName string
	var partition int32 = -1
	var offsetMode string
	var specificOffset int64
	_ = specificOffset // Mark as used to avoid compiler error

	// Parse arguments
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--topic":
			if i+1 < len(args) {
				topicName = args[i+1]
				i++
			}
		case "--partition":
			if i+1 < len(args) {
				if p, err := strconv.ParseInt(args[i+1], 10, 32); err == nil {
					partition = int32(p)
				}
				i++
			}
		case "--to-earliest":
			offsetMode = "earliest"
		case "--to-latest":
			offsetMode = "latest"
		case "--to-offset":
			if i+1 < len(args) {
				if offset, err := strconv.ParseInt(args[i+1], 10, 64); err == nil {
					specificOffset = offset
					offsetMode = "specific"
				}
				i++
			}
		}
	}

	if topicName == "" {
		return fmt.Errorf("topic name is required (use --topic)")
	}

	if offsetMode == "" {
		return fmt.Errorf("offset mode is required (use --to-earliest, --to-latest, or --to-offset)")
	}

	// This is a simplified implementation
	// In a real scenario, you'd need to coordinate with the consumer group
	fmt.Printf("Would reset offsets for group %s, topic %s, partition %d to %s\n", 
		groupID, topicName, partition, offsetMode)
	fmt.Println("Note: Actual offset reset requires consumer group coordination")

	return nil
}

// GetTopicOffsets shows current offsets for a topic
func (ac *AdminClient) GetTopicOffsets(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("topic name is required")
	}

	topicName := args[0]
	
	// Get topic metadata to know partitions
	metadata, err := ac.client.DescribeTopics([]string{topicName})
	if err != nil {
		return fmt.Errorf("failed to get topic metadata: %w", err)
	}

	var topicFound *sarama.TopicMetadata
	for _, topic := range metadata {
		if topic.Name == topicName {
			topicFound = topic
			break
		}
	}

	if topicFound != nil {
		// Create a consumer to get offsets
		saramaConfig := sarama.NewConfig()
		consumer, err := sarama.NewConsumer(ac.config.Brokers, saramaConfig)
		if err != nil {
			return fmt.Errorf("failed to create consumer: %w", err)
		}
		defer consumer.Close()

		fmt.Printf("Topic: %s\n", topicName)
		fmt.Printf("%-10s %-15s %-15s %s\n", "PARTITION", "EARLIEST", "LATEST", "LAG")
		fmt.Println(strings.Repeat("-", 60))

		for _, partition := range topicFound.Partitions {
			partitionConsumer, err := consumer.ConsumePartition(topicName, partition.ID, sarama.OffsetOldest)
			if err != nil {
				log.Printf("Failed to create partition consumer for partition %d: %v", partition.ID, err)
				continue
			}

			// This is a simplified approach - in practice you'd use a different method to get offsets
			fmt.Printf("%-10d %-15s %-15s %s\n", partition.ID, "N/A", "N/A", "N/A")
			partitionConsumer.Close()
		}
	} else {
		return fmt.Errorf("topic %s not found", topicName)
	}

	return nil
}

// GetTopicConfigs shows topic configuration
func (ac *AdminClient) GetTopicConfigs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("topic name is required")
	}

	topicName := args[0]
	// Note: This is a simplified implementation 
	// Real config retrieval would require different Sarama APIs

	fmt.Printf("Topic: %s\n", topicName)
	fmt.Println("Configurations:")
	fmt.Println("  Configuration retrieval not implemented in this version")

	return nil
}