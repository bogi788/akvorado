package kafka

import "github.com/Shopify/sarama"

// Configuration defines how we connect to a Kafka cluster.
type Configuration struct {
	// Topic defines the topic to write flows to.
	Topic string
	// Brokers is the list of brokers to connect to.
	Brokers []string
	// Version is the version of Kafka we assume to work
	Version Version
}

// DefaultConfiguration represents the default configuration for connecting to Kafka.
var DefaultConfiguration = Configuration{
	Topic:   "flows",
	Brokers: []string{"127.0.0.1:9092"},
	Version: Version(sarama.V2_8_1_0),
}

// Version represents a supported version of Kafka
type Version sarama.KafkaVersion

// UnmarshalText parses a version of Kafka
func (v *Version) UnmarshalText(text []byte) error {
	version, err := sarama.ParseKafkaVersion(string(text))
	if err != nil {
		return err
	}
	*v = Version(version)
	return nil
}

// String turns a Kafka version into a string
func (v Version) String() string {
	return sarama.KafkaVersion(v).String()
}

// MarshalText turns a Kafka version intro a string
func (v Version) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
