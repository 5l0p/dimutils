package kafka

import (
	"fmt"

	"github.com/og-dim9/dimutils/pkg/consume"
	"github.com/og-dim9/dimutils/pkg/kafkaadmin"
	"github.com/og-dim9/dimutils/pkg/produce"
)

// Run is the main entry point for kafka functionality
func Run(args []string) error {
	if len(args) == 0 {
		return printHelp()
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "consume", "c":
		return consume.Run(subArgs)
	case "produce", "p":
		return produce.Run(subArgs)
	case "admin", "a":
		return kafkaadmin.Run(subArgs)
	case "help", "-h", "--help":
		return printHelp()
	default:
		return fmt.Errorf("unknown kafka subcommand: %s. Use 'kafka help' to see available commands", subcommand)
	}
}

func printHelp() error {
	help := `Usage: kafka <subcommand> [options]

Kafka utilities for consuming, producing, and administering Kafka clusters.

Subcommands:
  consume, c        Consume messages from Kafka topics
  produce, p        Produce messages to Kafka topics  
  admin, a          Administer Kafka topics and consumer groups
  help              Show this help message

Global Options:
  --brokers, -b BROKERS     Comma-separated list of brokers (default: localhost:9092)
  --verbose, -v             Verbose output

Examples:
  kafka consume my-topic
  kafka produce my-topic --key mykey < data.txt
  kafka admin list-topics
  kafka admin create-topic my-topic --partitions 3

Use 'kafka <subcommand> --help' for detailed help on each subcommand.`

	fmt.Println(help)
	return nil
}