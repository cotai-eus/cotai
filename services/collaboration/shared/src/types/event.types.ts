export interface DomainEvent<T = any> {
  /**
   * Unique event ID
   */
  eventId: string;

  /**
   * Event type (e.g., 'tenant.created', 'agenda.reminder.due')
   */
  eventType: string;

  /**
   * Aggregate ID (primary entity ID)
   */
  aggregateId: string;

  /**
   * Aggregate type (e.g., 'Tenant', 'Evento')
   */
  aggregateType: string;

  /**
   * Tenant ID for multi-tenancy
   */
  tenantId: string;

  /**
   * Event timestamp (ISO 8601)
   */
  timestamp: string;

  /**
   * Event version for schema evolution
   */
  version: number;

  /**
   * Event payload
   */
  payload: T;

  /**
   * Event metadata
   */
  metadata: EventMetadata;
}

export interface EventMetadata {
  /**
   * Correlation ID for tracing across services
   */
  correlationId: string;

  /**
   * Causation ID (ID of event that caused this event)
   */
  causationId?: string;

  /**
   * User ID who triggered the event
   */
  userId?: string;

  /**
   * Additional metadata
   */
  [key: string]: any;
}
