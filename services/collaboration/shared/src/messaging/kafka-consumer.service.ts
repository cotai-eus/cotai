import { Injectable, OnModuleInit, OnModuleDestroy, Logger } from '@nestjs/common';
import { Kafka, Consumer, EachMessagePayload, KafkaMessage } from 'kafkajs';
import { DomainEvent } from '../types';

export interface KafkaConsumerConfig {
  clientId: string;
  brokers: string[];
  groupId: string;
  topics: string[];
}

export type MessageHandler = (event: DomainEvent) => Promise<void>;

@Injectable()
export class KafkaConsumerService implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(KafkaConsumerService.name);
  private kafka: Kafka;
  private consumer: Consumer;
  private handlers: Map<string, MessageHandler> = new Map();
  private isConnected = false;

  constructor(private readonly config: KafkaConsumerConfig) {
    this.kafka = new Kafka({
      clientId: this.config.clientId,
      brokers: this.config.brokers,
      retry: {
        retries: 5,
        initialRetryTime: 300,
        multiplier: 2,
      },
    });

    this.consumer = this.kafka.consumer({
      groupId: this.config.groupId,
      retry: {
        retries: 5,
        initialRetryTime: 300,
        multiplier: 2,
      },
    });
  }

  /**
   * Register a handler for a specific event type
   * @param eventType Event type to handle
   * @param handler Message handler function
   */
  registerHandler(eventType: string, handler: MessageHandler): void {
    this.handlers.set(eventType, handler);
    this.logger.log(`Registered handler for event type: ${eventType}`);
  }

  /**
   * Register multiple handlers
   * @param handlers Map of event types to handlers
   */
  registerHandlers(handlers: Record<string, MessageHandler>): void {
    Object.entries(handlers).forEach(([eventType, handler]) => {
      this.registerHandler(eventType, handler);
    });
  }

  async onModuleInit() {
    try {
      await this.consumer.connect();
      this.isConnected = true;
      this.logger.log('Kafka consumer connected');

      // Subscribe to topics
      for (const topic of this.config.topics) {
        await this.consumer.subscribe({ topic, fromBeginning: false });
        this.logger.log(`Subscribed to topic: ${topic}`);
      }

      // Start consuming
      await this.consumer.run({
        eachMessage: async (payload: EachMessagePayload) => {
          await this.handleMessage(payload);
        },
      });

      this.logger.log('Kafka consumer started');
    } catch (error) {
      this.logger.error(`Failed to start Kafka consumer: ${error.message}`, error.stack);
      throw error;
    }
  }

  /**
   * Handle incoming Kafka message
   */
  private async handleMessage(payload: EachMessagePayload): Promise<void> {
    const { topic, partition, message } = payload;
    const eventType = message.headers?.['event-type']?.toString();
    const tenantId = message.headers?.['tenant-id']?.toString();
    const correlationId = message.headers?.['correlation-id']?.toString();

    this.logger.debug(
      `Received message: topic=${topic}, partition=${partition}, eventType=${eventType}, tenantId=${tenantId}`,
    );

    if (!eventType) {
      this.logger.warn('Message missing event-type header, skipping', {
        topic,
        partition,
        offset: message.offset,
      });
      return;
    }

    const handler = this.handlers.get(eventType);
    if (!handler) {
      this.logger.warn(`No handler registered for event type: ${eventType}`);
      return;
    }

    try {
      const event = this.parseMessage(message);
      await handler(event);

      this.logger.debug(
        `Successfully processed event: ${eventType}, correlationId=${correlationId}`,
      );
    } catch (error) {
      this.logger.error(
        `Error processing message: ${error.message}`,
        {
          eventType,
          tenantId,
          correlationId,
          topic,
          partition,
          offset: message.offset,
          stack: error.stack,
        },
      );

      // TODO: Implement DLQ (Dead Letter Queue) logic here
      // For now, we log and continue
      // In production, you might want to:
      // 1. Send to DLQ topic
      // 2. Pause consumption and alert
      // 3. Retry with exponential backoff
    }
  }

  /**
   * Parse Kafka message to DomainEvent
   */
  private parseMessage(message: KafkaMessage): DomainEvent {
    try {
      const value = message.value?.toString();
      if (!value) {
        throw new Error('Message value is empty');
      }

      const event = JSON.parse(value) as DomainEvent;
      return event;
    } catch (error) {
      this.logger.error(`Failed to parse message: ${error.message}`);
      throw error;
    }
  }

  async onModuleDestroy() {
    try {
      await this.consumer.disconnect();
      this.isConnected = false;
      this.logger.log('Kafka consumer disconnected');
    } catch (error) {
      this.logger.error(`Error disconnecting Kafka consumer: ${error.message}`);
    }
  }

  /**
   * Get consumer status
   */
  getStatus() {
    return {
      isConnected: this.isConnected,
      groupId: this.config.groupId,
      topics: this.config.topics,
      handlers: Array.from(this.handlers.keys()),
    };
  }
}
