package com.cotai.keycloak.config;

import org.apache.kafka.clients.producer.KafkaProducer;
import org.apache.kafka.clients.producer.Producer;
import org.apache.kafka.clients.producer.ProducerConfig;
import org.apache.kafka.clients.producer.ProducerRecord;
import org.apache.kafka.clients.producer.RecordMetadata;
import org.apache.kafka.common.serialization.StringSerializer;
import org.jboss.logging.Logger;

import java.time.Duration;
import java.util.Properties;
import java.util.concurrent.Future;

/**
 * Singleton manager for Kafka producer used by event listeners.
 *
 * Provides thread-safe access to a Kafka producer instance for publishing
 * authentication and admin events to Kafka topics.
 *
 * Configuration via environment variables:
 * - KAFKA_BOOTSTRAP_SERVERS: Kafka broker addresses (default: kafka:9092)
 * - KAFKA_PRODUCER_CLIENT_ID: Client identifier (default: keycloak-event-publisher)
 * - KAFKA_ACKS: Acknowledgment mode (default: 1)
 * - KAFKA_RETRIES: Number of retries (default: 3)
 * - KAFKA_LINGER_MS: Batch linger time (default: 10)
 * - KAFKA_COMPRESSION_TYPE: Compression algorithm (default: snappy)
 *
 * @author CotAI Development Team
 * @version 1.0.0
 */
public class KafkaProducerManager {

    private static final Logger logger = Logger.getLogger(KafkaProducerManager.class);

    private static volatile KafkaProducerManager instance;
    private static final Object lock = new Object();

    private final Producer<String, String> producer;
    private volatile boolean isClosing = false;

    /**
     * Private constructor for singleton pattern.
     * Initializes Kafka producer with configuration from environment.
     */
    private KafkaProducerManager() {
        logger.info("Initializing Kafka Producer for CotAI event publishing");

        Properties props = new Properties();

        // Bootstrap servers
        String bootstrapServers = getEnvOrDefault("KAFKA_BOOTSTRAP_SERVERS", "kafka:9092");
        props.put(ProducerConfig.BOOTSTRAP_SERVERS_CONFIG, bootstrapServers);

        // Client ID
        String clientId = getEnvOrDefault("KAFKA_PRODUCER_CLIENT_ID", "keycloak-event-publisher");
        props.put(ProducerConfig.CLIENT_ID_CONFIG, clientId);

        // Serializers
        props.put(ProducerConfig.KEY_SERIALIZER_CLASS_CONFIG, StringSerializer.class.getName());
        props.put(ProducerConfig.VALUE_SERIALIZER_CLASS_CONFIG, StringSerializer.class.getName());

        // Acknowledgment configuration
        String acks = getEnvOrDefault("KAFKA_ACKS", "1");
        props.put(ProducerConfig.ACKS_CONFIG, acks);

        // Retry configuration
        int retries = Integer.parseInt(getEnvOrDefault("KAFKA_RETRIES", "3"));
        props.put(ProducerConfig.RETRIES_CONFIG, retries);
        props.put(ProducerConfig.MAX_IN_FLIGHT_REQUESTS_PER_CONNECTION, 1); // Ensure ordering

        // Batching and latency
        int lingerMs = Integer.parseInt(getEnvOrDefault("KAFKA_LINGER_MS", "10"));
        props.put(ProducerConfig.LINGER_MS_CONFIG, lingerMs);

        int batchSize = Integer.parseInt(getEnvOrDefault("KAFKA_BATCH_SIZE", "16384"));
        props.put(ProducerConfig.BATCH_SIZE_CONFIG, batchSize);

        // Compression
        String compressionType = getEnvOrDefault("KAFKA_COMPRESSION_TYPE", "snappy");
        props.put(ProducerConfig.COMPRESSION_TYPE_CONFIG, compressionType);

        // Timeouts
        props.put(ProducerConfig.REQUEST_TIMEOUT_MS_CONFIG, 30000);
        props.put(ProducerConfig.DELIVERY_TIMEOUT_MS_CONFIG, 120000);

        // Idempotence for exactly-once semantics
        props.put(ProducerConfig.ENABLE_IDEMPOTENCE_CONFIG, true);

        logger.infof("Kafka Producer configuration: bootstrap.servers=%s, client.id=%s, acks=%s",
                    bootstrapServers, clientId, acks);

        try {
            this.producer = new KafkaProducer<>(props);
            logger.info("Kafka Producer initialized successfully");
        } catch (Exception e) {
            logger.error("Failed to create Kafka Producer", e);
            throw new RuntimeException("Cannot initialize Kafka Producer", e);
        }
    }

    /**
     * Get singleton instance (thread-safe double-checked locking).
     */
    public static KafkaProducerManager getInstance() {
        if (instance == null) {
            synchronized (lock) {
                if (instance == null) {
                    instance = new KafkaProducerManager();
                }
            }
        }
        return instance;
    }

    /**
     * Send a message to Kafka asynchronously.
     *
     * @param topic The Kafka topic
     * @param key The message key (used for partitioning)
     * @param value The message value (JSON payload)
     * @return Future with record metadata
     */
    public Future<RecordMetadata> send(String topic, String key, String value) {
        if (isClosing) {
            logger.warn("Attempted to send message while producer is closing");
            throw new IllegalStateException("Kafka producer is closing");
        }

        try {
            ProducerRecord<String, String> record = new ProducerRecord<>(topic, key, value);

            return producer.send(record, (metadata, exception) -> {
                if (exception != null) {
                    logger.errorf(exception, "Failed to send message to topic %s with key %s",
                                topic, key);
                } else {
                    logger.debugf("Message sent successfully to %s partition %d offset %d",
                                metadata.topic(), metadata.partition(), metadata.offset());
                }
            });

        } catch (Exception e) {
            logger.errorf(e, "Error sending message to topic %s", topic);
            throw new RuntimeException("Failed to send Kafka message", e);
        }
    }

    /**
     * Flush all buffered messages to Kafka.
     */
    public void flush() {
        if (!isClosing) {
            producer.flush();
        }
    }

    /**
     * Close the Kafka producer gracefully.
     */
    public void close() {
        if (isClosing) {
            return;
        }

        isClosing = true;
        logger.info("Closing Kafka Producer");

        try {
            // Flush remaining messages
            producer.flush();

            // Close with timeout (10 seconds)
            producer.close(Duration.ofSeconds(10));
            logger.info("Kafka Producer closed successfully");

        } catch (Exception e) {
            logger.error("Error closing Kafka Producer", e);
        }
    }

    /**
     * Get environment variable with default fallback.
     */
    private String getEnvOrDefault(String key, String defaultValue) {
        String value = System.getenv(key);
        return (value != null && !value.isBlank()) ? value : defaultValue;
    }
}
