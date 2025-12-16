import { Injectable, OnModuleInit, OnModuleDestroy, Logger } from '@nestjs/common';
import { Kafka, Producer, ProducerRecord, RecordMetadata } from 'kafkajs';
import { v4 as uuidv4 } from 'uuid';
import { DomainEvent } from '../types';

export interface KafkaProducerConfig {
  clientId: string;
  brokers: string[];
  retries?: number;
  compression?: number;
}

@Injectable()
export class KafkaProducerService implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(KafkaProducerService.name);
  private kafka: Kafka;
  private producer: Producer;
  private isConnected = false;

  constructor(private readonly config: KafkaProducerConfig) {
    this.kafka = new Kafka({
      clientId: this.config.clientId,
      brokers: this.config.brokers,
      retry: {
        retries: this.config.retries || 3,
        initialRetryTime: 100,
        multiplier: 2,
      },
    });

    this.producer = this.kafka.producer({
      idempotent: true,
      maxInFlightRequests: 5,
    });
  }

  async onModuleInit() {
    try {
      await this.producer.connect();
      this.isConnected = true;
      this.logger.log(
        `Kafka producer connected to brokers: ${this.config.brokers.join(', ')}`,
      );
    } catch (error) {
      this.logger.error(`Failed to connect Kafka producer: ${error.message}`, error.stack);
      throw error;
    }
  }

  /**
   * Publish a domain event to Kafka
   * @param topic Kafka topic
   * @param event Domain event
   * @returns Record metadata
   */
  async publishEvent<T = any>(
    topic: string,
    event: DomainEvent<T>,
  ): Promise<RecordMetadata[]> {
    if (!this.isConnected) {
      throw new Error('Kafka producer not connected');
    }

    const record: ProducerRecord = {
      topic,
      messages: [
        {
          key: event.aggregateId,
          value: JSON.stringify(event),
          headers: {
            'event-type': event.eventType,
            'event-id': event.eventId,
            'tenant-id': event.tenantId,
            'correlation-id': event.metadata.correlationId,
            'aggregate-type': event.aggregateType,
            timestamp: event.timestamp,
          },
        },
      ],
    };

    try {
      const metadata = await this.producer.send(record);
      this.logger.debug(
        `Event published: ${event.eventType} to topic ${topic}, partition ${metadata[0].partition}`,
      );
      return metadata;
    } catch (error) {
      this.logger.error(
        `Failed to publish event ${event.eventType} to ${topic}: ${error.message}`,
        error.stack,
      );
      throw error;
    }
  }

  /**
   * Send raw message to Kafka
   * @param topic Kafka topic
   * @param key Message key
   * @param value Message value
   * @param headers Optional headers
   */
  async sendMessage(
    topic: string,
    key: string,
    value: any,
    headers?: Record<string, string>,
  ): Promise<RecordMetadata[]> {
    if (!this.isConnected) {
      throw new Error('Kafka producer not connected');
    }

    const record: ProducerRecord = {
      topic,
      messages: [
        {
          key,
          value: typeof value === 'string' ? value : JSON.stringify(value),
          headers,
        },
      ],
    };

    try {
      const metadata = await this.producer.send(record);
      this.logger.debug(`Message sent to topic ${topic}, partition ${metadata[0].partition}`);
      return metadata;
    } catch (error) {
      this.logger.error(`Failed to send message to ${topic}: ${error.message}`, error.stack);
      throw error;
    }
  }

  /**
   * Create a domain event helper
   */
  createDomainEvent<T = any>(
    eventType: string,
    aggregateId: string,
    aggregateType: string,
    tenantId: string,
    payload: T,
    correlationId?: string,
  ): DomainEvent<T> {
    return {
      eventId: uuidv4(),
      eventType,
      aggregateId,
      aggregateType,
      tenantId,
      timestamp: new Date().toISOString(),
      version: 1,
      payload,
      metadata: {
        correlationId: correlationId || uuidv4(),
      },
    };
  }

  async onModuleDestroy() {
    try {
      await this.producer.disconnect();
      this.isConnected = false;
      this.logger.log('Kafka producer disconnected');
    } catch (error) {
      this.logger.error(`Error disconnecting Kafka producer: ${error.message}`);
    }
  }
}
