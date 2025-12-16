package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/cotai/tenant-manager/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// KafkaProducer handles publishing events to Kafka
type KafkaProducer struct {
	producer sarama.AsyncProducer
	topic    string
	logger   *zap.Logger
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(brokers []string, topic string, logger *zap.Logger) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all replicas
	config.Producer.Retry.Max = 3
	config.Producer.Compression = sarama.CompressionSnappy
	config.Version = sarama.V3_0_0_0

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	kp := &KafkaProducer{
		producer: producer,
		topic:    topic,
		logger:   logger,
	}

	// Start goroutines to handle successes and errors
	go kp.handleSuccesses()
	go kp.handleErrors()

	return kp, nil
}

// PublishTenantCreated publishes a tenant.created event
func (p *KafkaProducer) PublishTenantCreated(ctx context.Context, tenant *domain.Tenant) error {
	return p.publishEvent(ctx, EventTenantCreated, tenant)
}

// PublishTenantActivated publishes a tenant.activated event
func (p *KafkaProducer) PublishTenantActivated(ctx context.Context, tenant *domain.Tenant) error {
	return p.publishEvent(ctx, EventTenantActivated, tenant)
}

// PublishTenantSuspended publishes a tenant.suspended event
func (p *KafkaProducer) PublishTenantSuspended(ctx context.Context, tenant *domain.Tenant) error {
	return p.publishEvent(ctx, EventTenantSuspended, tenant)
}

// PublishTenantDeleted publishes a tenant.deleted event
func (p *KafkaProducer) PublishTenantDeleted(ctx context.Context, tenant *domain.Tenant) error {
	return p.publishEvent(ctx, EventTenantDeleted, tenant)
}

// PublishTenantUpdated publishes a tenant.updated event
func (p *KafkaProducer) PublishTenantUpdated(ctx context.Context, tenant *domain.Tenant) error {
	return p.publishEvent(ctx, EventTenantUpdated, tenant)
}

// publishEvent is a generic method to publish any tenant lifecycle event
func (p *KafkaProducer) publishEvent(ctx context.Context, eventType EventType, tenant *domain.Tenant) error {
	// Get correlation ID from context, or generate new one
	correlationID := GetCorrelationID(ctx)
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	// Create event
	event := TenantLifecycleEvent{
		EventID:       uuid.New().String(),
		EventType:     eventType,
		TenantID:      tenant.TenantID.String(),
		Timestamp:     time.Now().UTC(),
		CorrelationID: correlationID,
		Payload:       TenantToEventPayload(tenant),
	}

	// Marshal to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("Failed to marshal event",
			zap.String("eventType", string(eventType)),
			zap.String("tenantId", tenant.TenantID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(tenant.TenantID.String()),
		Value: sarama.ByteEncoder(eventJSON),
		Headers: []sarama.RecordHeader{
			{Key: []byte("event-type"), Value: []byte(eventType)},
			{Key: []byte("tenant-id"), Value: []byte(tenant.TenantID.String())},
			{Key: []byte("correlation-id"), Value: []byte(correlationID)},
		},
	}

	// Send message (non-blocking)
	p.producer.Input() <- msg

	p.logger.Info("Event published to Kafka",
		zap.String("eventType", string(eventType)),
		zap.String("tenantId", tenant.TenantID.String()),
		zap.String("correlationId", correlationID),
	)

	return nil
}

// handleSuccesses processes successful message deliveries
func (p *KafkaProducer) handleSuccesses() {
	for success := range p.producer.Successes() {
		p.logger.Debug("Message sent successfully",
			zap.String("topic", success.Topic),
			zap.Int32("partition", success.Partition),
			zap.Int64("offset", success.Offset),
		)
	}
}

// handleErrors processes message delivery errors
func (p *KafkaProducer) handleErrors() {
	for err := range p.producer.Errors() {
		p.logger.Error("Failed to send message to Kafka",
			zap.String("topic", err.Msg.Topic),
			zap.Error(err.Err),
		)
	}
}

// Close closes the Kafka producer
func (p *KafkaProducer) Close() error {
	p.logger.Info("Closing Kafka producer...")
	return p.producer.Close()
}

// GetCorrelationID extracts correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value("correlationId").(string); ok {
		return correlationID
	}
	if correlationID, ok := ctx.Value("correlation_id").(string); ok {
		return correlationID
	}
	return ""
}
