# CotAI Keycloak Extensions

Custom Keycloak SPIs for CotAI multi-tenant authentication and event publishing.

## Overview

This project provides two custom Keycloak Service Provider Interface (SPI) implementations:

1. **TenantIdMapper**: Protocol mapper that injects `tenant_id` claim into JWT tokens
2. **KafkaEventListenerProvider**: Event listener that publishes auth events to Kafka

## Project Structure

```
auth-service/
├── src/
│   ├── main/
│   │   ├── java/com/cotai/keycloak/
│   │   │   ├── mapper/
│   │   │   │   └── TenantIdMapper.java
│   │   │   ├── listener/
│   │   │   │   ├── KafkaEventListenerProvider.java
│   │   │   │   └── KafkaEventListenerProviderFactory.java
│   │   │   └── config/
│   │   │       └── KafkaProducerManager.java
│   │   └── resources/
│   │       └── META-INF/services/
│   │           ├── org.keycloak.protocol.ProtocolMapper
│   │           └── org.keycloak.events.EventListenerProviderFactory
│   └── test/
├── pom.xml
├── Dockerfile
├── build.sh
└── README.md
```

## Features

### TenantIdMapper

- Reads `tenant_id` from user attributes
- Injects into Access Token, ID Token, and UserInfo responses
- Supports multi-tenant users (uses first tenant_id)
- Configurable claim name (default: `tenant_id`)

### Kafka Event Listener

- Publishes authentication events to `auth.events` topic
- Publishes admin events to `auth.admin.events` topic
- Includes: event type, timestamp, tenant_id, user details, IP address
- Non-blocking: failures don't break auth flow
- Configurable via environment variables

## Building

### Using Docker (Recommended)

```bash
# Build the custom Keycloak image with extensions
./build.sh

# Or manually:
docker build -t cotai-keycloak-custom:latest .
```

### Using Maven (Requires JDK 17+ and Maven 3.9+)

```bash
# Build JAR only
mvn clean package

# Run tests
mvn test

# Build with tests
mvn clean package -DskipTests=false
```

The built JAR will be in `target/cotai-keycloak-extensions-1.0.0-SNAPSHOT.jar`

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `KAFKA_BOOTSTRAP_SERVERS` | `kafka:9092` | Kafka broker addresses |
| `KAFKA_PRODUCER_CLIENT_ID` | `keycloak-event-publisher` | Kafka client ID |
| `KAFKA_ACKS` | `1` | Acknowledgment mode (0, 1, all) |
| `KAFKA_RETRIES` | `3` | Number of send retries |
| `KAFKA_LINGER_MS` | `10` | Batch linger time (ms) |
| `KAFKA_COMPRESSION_TYPE` | `snappy` | Compression (none, gzip, snappy, lz4) |

### Keycloak Realm Configuration

To enable the custom extensions in your realm:

1. **Enable Kafka Event Listener**:
   - Navigate to: Realm Settings → Events → Event Listeners
   - Add: `cotai-kafka-event-listener`
   - Save

2. **Configure Tenant ID Mapper**:
   - Navigate to: Clients → [Your Client] → Client Scopes → Dedicated Scope
   - Add Mapper → From Configuration → `CotAI Tenant ID Mapper`
   - Configure claim name (default: `tenant_id`)
   - Save

## Integration with CotAI Platform

### User Setup

Each user must have a `tenant_id` attribute set:

```bash
# Via Keycloak Admin API or UI
curl -X PUT http://localhost:8080/admin/realms/cotai/users/{userId} \
  -H "Authorization: Bearer {admin-token}" \
  -H "Content-Type: application/json" \
  -d '{
    "attributes": {
      "tenant_id": ["550e8400-e29b-41d4-a716-446655440000"]
    }
  }'
```

### JWT Token Example

After authentication, tokens will include:

```json
{
  "exp": 1702345678,
  "iat": 1702344778,
  "sub": "user-uuid",
  "email": "user@example.com",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "roles": ["cotai_user"]
}
```

### Kafka Event Schema

**Authentication Event** (topic: `auth.events`):

```json
{
  "eventId": "uuid",
  "eventType": "LOGIN",
  "timestamp": "2024-12-16T12:00:00Z",
  "realmId": "cotai",
  "userId": "user-uuid",
  "tenantId": "tenant-uuid",
  "sessionId": "session-uuid",
  "ipAddress": "192.168.1.100",
  "clientId": "cotai-web-app",
  "details": {
    "username": "user@example.com",
    "auth_method": "oauth",
    "redirect_uri": "http://localhost:3000/callback"
  }
}
```

**Admin Event** (topic: `auth.admin.events`):

```json
{
  "eventId": "uuid",
  "operationType": "CREATE",
  "timestamp": "2024-12-16T12:00:00Z",
  "realmId": "cotai",
  "resourceType": "USER",
  "resourcePath": "users/user-uuid",
  "authDetails": {
    "userId": "admin-uuid",
    "realmId": "cotai",
    "ipAddress": "192.168.1.100"
  }
}
```

## Testing

### Unit Tests

```bash
mvn test
```

### Integration Testing

1. Start infrastructure:
   ```bash
   docker compose -f ../../docker-compose.dev.yml up -d postgres-identity kafka
   ```

2. Build and start custom Keycloak:
   ```bash
   ./build.sh
   docker compose up -d
   ```

3. Import realm configuration:
   ```bash
   # Via Keycloak Admin UI
   # Navigate to: Create Realm → Import → Select realm-cotai.json
   ```

4. Test authentication flow:
   ```bash
   # Get access token
   curl -X POST http://localhost:8080/realms/cotai/protocol/openid-connect/token \
     -d "client_id=cotai-web-app" \
     -d "username=admin@cotai.local" \
     -d "password=Admin@123" \
     -d "grant_type=password" | jq .

   # Verify tenant_id in token
   # Check Kafka topic for events
   docker exec cotai-kafka kafka-console-consumer \
     --bootstrap-server localhost:9092 \
     --topic auth.events \
     --from-beginning
   ```

## Troubleshooting

### Extension Not Loading

1. Verify JAR is in `/opt/keycloak/providers/`:
   ```bash
   docker exec cotai-keycloak ls -la /opt/keycloak/providers/
   ```

2. Check Keycloak logs:
   ```bash
   docker logs cotai-keycloak
   ```

3. Look for SPI registration messages:
   ```
   INFO  [org.keycloak.services] (main) KC-SERVICES0050: Initializing master realm
   INFO  [com.cotai.keycloak.listener.KafkaEventListenerProviderFactory] Initializing CotAI Kafka Event Listener Provider
   ```

### Kafka Connection Issues

1. Verify Kafka is running:
   ```bash
   docker exec cotai-kafka kafka-broker-api-versions --bootstrap-server localhost:9092
   ```

2. Check environment variables:
   ```bash
   docker exec cotai-keycloak env | grep KAFKA
   ```

3. Review Kafka producer logs in Keycloak container

### tenant_id Not in Token

1. Verify user has `tenant_id` attribute
2. Ensure mapper is configured on client
3. Check token endpoint response includes the claim
4. Review Keycloak server logs for mapper errors

## Development

### Adding New Mappers

1. Create class extending `AbstractOIDCProtocolMapper`
2. Implement required interfaces (`OIDCAccessTokenMapper`, etc.)
3. Override `setClaim()` method
4. Register in `META-INF/services/org.keycloak.protocol.ProtocolMapper`

### Adding New Event Listeners

1. Create `EventListenerProvider` implementation
2. Create corresponding `EventListenerProviderFactory`
3. Register in `META-INF/services/org.keycloak.events.EventListenerProviderFactory`

## Dependencies

- Keycloak 23.0.3
- Kafka Clients 3.6.1
- Jackson 2.16.1
- JDK 17+
- Maven 3.9+

## License

Copyright © 2024 CotAI. All rights reserved.

## Support

For issues and questions:
- Internal: Slack #cotai-identity-team
- Documentation: [CotAI Architecture Docs](../../../docs/architecture/)
