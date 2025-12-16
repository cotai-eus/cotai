import { Injectable } from '@nestjs/common';
import * as client from 'prom-client';

export interface MetricsConfig {
  serviceName: string;
}

@Injectable()
export class MetricsService {
  private readonly registry: client.Registry;

  // HTTP metrics
  public readonly httpRequestDuration: client.Histogram;
  public readonly httpRequestTotal: client.Counter;

  // Database metrics
  public readonly dbQueryDuration: client.Histogram;
  public readonly dbConnectionsActive: client.Gauge;

  // Kafka metrics
  public readonly kafkaMessagesProduced: client.Counter;
  public readonly kafkaMessagesConsumed: client.Counter;
  public readonly kafkaProducerErrors: client.Counter;
  public readonly kafkaConsumerLag: client.Gauge;

  constructor(config: MetricsConfig) {
    this.registry = new client.Registry();

    // Set default labels
    this.registry.setDefaultLabels({
      service: config.serviceName,
    });

    // Collect default metrics (CPU, memory, etc.)
    client.collectDefaultMetrics({ register: this.registry });

    // HTTP request duration histogram
    this.httpRequestDuration = new client.Histogram({
      name: 'http_request_duration_seconds',
      help: 'Duration of HTTP requests in seconds',
      labelNames: ['method', 'route', 'status_code', 'tenant_id'],
      buckets: [0.01, 0.05, 0.1, 0.5, 1, 2, 5],
      registers: [this.registry],
    });

    // HTTP request counter
    this.httpRequestTotal = new client.Counter({
      name: 'http_requests_total',
      help: 'Total number of HTTP requests',
      labelNames: ['method', 'route', 'status_code', 'tenant_id'],
      registers: [this.registry],
    });

    // Database query duration
    this.dbQueryDuration = new client.Histogram({
      name: 'db_query_duration_seconds',
      help: 'Duration of database queries in seconds',
      labelNames: ['operation', 'table', 'tenant_id'],
      buckets: [0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1],
      registers: [this.registry],
    });

    // Database active connections
    this.dbConnectionsActive = new client.Gauge({
      name: 'db_connections_active',
      help: 'Number of active database connections',
      labelNames: ['tenant_id'],
      registers: [this.registry],
    });

    // Kafka messages produced
    this.kafkaMessagesProduced = new client.Counter({
      name: 'kafka_messages_produced_total',
      help: 'Total number of Kafka messages produced',
      labelNames: ['topic', 'tenant_id'],
      registers: [this.registry],
    });

    // Kafka messages consumed
    this.kafkaMessagesConsumed = new client.Counter({
      name: 'kafka_messages_consumed_total',
      help: 'Total number of Kafka messages consumed',
      labelNames: ['topic', 'group_id'],
      registers: [this.registry],
    });

    // Kafka producer errors
    this.kafkaProducerErrors = new client.Counter({
      name: 'kafka_producer_errors_total',
      help: 'Total number of Kafka producer errors',
      labelNames: ['topic', 'error_type'],
      registers: [this.registry],
    });

    // Kafka consumer lag
    this.kafkaConsumerLag = new client.Gauge({
      name: 'kafka_consumer_lag',
      help: 'Kafka consumer lag (messages behind)',
      labelNames: ['topic', 'partition', 'group_id'],
      registers: [this.registry],
    });
  }

  /**
   * Get metrics in Prometheus format
   */
  async getMetrics(): Promise<string> {
    return this.registry.metrics();
  }

  /**
   * Get registry for custom metrics
   */
  getRegistry(): client.Registry {
    return this.registry;
  }

  /**
   * Create a custom counter
   */
  createCounter(name: string, help: string, labelNames?: string[]): client.Counter {
    return new client.Counter({
      name,
      help,
      labelNames,
      registers: [this.registry],
    });
  }

  /**
   * Create a custom gauge
   */
  createGauge(name: string, help: string, labelNames?: string[]): client.Gauge {
    return new client.Gauge({
      name,
      help,
      labelNames,
      registers: [this.registry],
    });
  }

  /**
   * Create a custom histogram
   */
  createHistogram(
    name: string,
    help: string,
    labelNames?: string[],
    buckets?: number[],
  ): client.Histogram {
    return new client.Histogram({
      name,
      help,
      labelNames,
      buckets,
      registers: [this.registry],
    });
  }
}
