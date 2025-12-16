# @cotai/shared-node

Shared utilities for CotAI Node.js microservices. This package provides common functionality for building multi-tenant, event-driven microservices using NestJS.

## Features

- **Authentication & Authorization**: JWT validation with Keycloak, tenant guards, permission guards
- **Multi-Tenancy**: Schema-per-tenant database connections with Row-Level Security (RLS)
- **Observability**: Structured logging (Winston), metrics (Prometheus), distributed tracing (Jaeger)
- **Messaging**: Kafka producer/consumer with domain events
- **Error Handling**: Domain exceptions and HTTP exception filters
- **Decorators**: Convenient parameter decorators for tenant ID, user ID, etc.

## Installation

```bash
npm install @cotai/shared-node
```

## Usage

### JWT Authentication

```typescript
import { JwtStrategy, JwtConfig } from '@cotai/shared-node';

const jwtConfig: JwtConfig = {
  jwksUri: 'http://keycloak:8080/realms/cotai/protocol/openid-connect/certs',
  issuer: 'http://keycloak:8080/realms/cotai',
  audience: 'cotai-web',
};

// In your auth module
providers: [
  {
    provide: JwtStrategy,
    useFactory: () => new JwtStrategy(jwtConfig),
  },
];
```

### Multi-Tenant Database

```typescript
import { TenantConnectionService } from '@cotai/shared-node';

const connectionService = new TenantConnectionService({
  type: 'postgres',
  host: 'localhost',
  port: 5432,
  database: 'cotai_collab',
  username: 'cotai_dev',
  password: 'dev_password',
});

// Execute in tenant context
await connectionService.executeInTenantContext('tenant-123', async (dataSource) => {
  const users = await dataSource.query('SELECT * FROM users');
  return users;
});
```

### Controllers with Decorators

```typescript
import { Controller, Get, UseGuards } from '@nestjs/common';
import { TenantGuard, TenantId, UserId, RequireRoles } from '@cotai/shared-node';

@Controller('eventos')
@UseGuards(TenantGuard)
export class EventoController {
  @Get()
  @RequireRoles('user', 'admin')
  async findAll(@TenantId() tenantId: string, @UserId() userId: string) {
    // tenantId and userId are automatically extracted
    return this.eventoService.findByTenant(tenantId);
  }
}
```

### Kafka Messaging

```typescript
import { KafkaProducerService, DomainEvent } from '@cotai/shared-node';

// Publish event
const producer = new KafkaProducerService({
  clientId: 'agenda-service',
  brokers: ['kafka:9092'],
});

const event = producer.createDomainEvent(
  'agenda.reminder.due',
  'evento-123',
  'Evento',
  'tenant-123',
  { message: 'Reminder due!' },
);

await producer.publishEvent('agenda.events', event);
```

## Development

```bash
# Install dependencies
npm install

# Build
npm run build

# Run tests
npm test

# Lint
npm run lint
```

## License

UNLICENSED - CotAI Platform
