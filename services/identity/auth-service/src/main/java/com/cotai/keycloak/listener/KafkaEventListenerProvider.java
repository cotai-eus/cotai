package com.cotai.keycloak.listener;

import com.cotai.keycloak.config.KafkaProducerManager;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.jboss.logging.Logger;
import org.keycloak.events.Event;
import org.keycloak.events.EventListenerProvider;
import org.keycloak.events.EventType;
import org.keycloak.events.admin.AdminEvent;
import org.keycloak.events.admin.OperationType;

import java.time.Instant;
import java.util.Map;
import java.util.UUID;

/**
 * Keycloak Event Listener that publishes authentication and admin events to Kafka.
 *
 * Publishes to topics:
 * - auth.events (user authentication events: login, logout, register, etc.)
 * - auth.admin.events (admin operations: user creation, role assignment, etc.)
 *
 * Event Schema includes:
 * - eventId: Unique event identifier
 * - eventType: LOGIN, LOGOUT, REGISTER, etc.
 * - tenantId: Extracted from user attributes
 * - userId: Keycloak user ID
 * - username: User's username/email
 * - timestamp: ISO-8601 timestamp
 * - ipAddress: Client IP address
 * - details: Additional event-specific data
 *
 * @author CotAI Development Team
 * @version 1.0.0
 */
public class KafkaEventListenerProvider implements EventListenerProvider {

    private static final Logger logger = Logger.getLogger(KafkaEventListenerProvider.class);

    private static final String AUTH_EVENTS_TOPIC = "auth.events";
    private static final String ADMIN_EVENTS_TOPIC = "auth.admin.events";

    private final ObjectMapper objectMapper;

    public KafkaEventListenerProvider() {
        this.objectMapper = new ObjectMapper();
    }

    /**
     * Called when a user event occurs (login, logout, register, etc.)
     */
    @Override
    public void onEvent(Event event) {
        // Filter events we want to publish
        if (!shouldPublishEvent(event.getType())) {
            return;
        }

        try {
            // Build event JSON payload
            ObjectNode payload = buildEventPayload(event);

            // Publish to Kafka
            String key = event.getUserId() != null ? event.getUserId() : event.getSessionId();
            KafkaProducerManager.getInstance().send(AUTH_EVENTS_TOPIC, key, payload.toString());

            logger.debugf("Published event %s for user %s to Kafka topic %s",
                         event.getType(), event.getUserId(), AUTH_EVENTS_TOPIC);

        } catch (Exception e) {
            logger.errorf(e, "Failed to publish event %s to Kafka", event.getType());
            // Don't throw - we don't want to break auth flow if Kafka is down
        }
    }

    /**
     * Called when an admin operation occurs (user management, configuration changes, etc.)
     */
    @Override
    public void onEvent(AdminEvent adminEvent, boolean includeRepresentation) {
        // Filter admin events we want to publish
        if (!shouldPublishAdminEvent(adminEvent.getOperationType())) {
            return;
        }

        try {
            // Build admin event JSON payload
            ObjectNode payload = buildAdminEventPayload(adminEvent, includeRepresentation);

            // Publish to Kafka
            String key = adminEvent.getAuthDetails() != null ?
                        adminEvent.getAuthDetails().getUserId() : UUID.randomUUID().toString();
            KafkaProducerManager.getInstance().send(ADMIN_EVENTS_TOPIC, key, payload.toString());

            logger.debugf("Published admin event %s on resource %s to Kafka topic %s",
                         adminEvent.getOperationType(),
                         adminEvent.getResourcePath(),
                         ADMIN_EVENTS_TOPIC);

        } catch (Exception e) {
            logger.errorf(e, "Failed to publish admin event %s to Kafka",
                         adminEvent.getOperationType());
            // Don't throw - we don't want to break admin operations if Kafka is down
        }
    }

    @Override
    public void close() {
        // Cleanup is handled by KafkaProducerManager singleton
    }

    /**
     * Determine if this event type should be published to Kafka.
     * We publish security-relevant events for audit and analytics.
     */
    private boolean shouldPublishEvent(EventType eventType) {
        return switch (eventType) {
            case LOGIN,
                 LOGIN_ERROR,
                 LOGOUT,
                 LOGOUT_ERROR,
                 REGISTER,
                 REGISTER_ERROR,
                 UPDATE_PASSWORD,
                 UPDATE_PASSWORD_ERROR,
                 UPDATE_EMAIL,
                 UPDATE_PROFILE,
                 VERIFY_EMAIL,
                 REFRESH_TOKEN,
                 REFRESH_TOKEN_ERROR,
                 CLIENT_LOGIN,
                 CLIENT_LOGIN_ERROR -> true;
            default -> false;
        };
    }

    /**
     * Determine if this admin event should be published.
     */
    private boolean shouldPublishAdminEvent(OperationType operationType) {
        return switch (operationType) {
            case CREATE, UPDATE, DELETE, ACTION -> true;
        };
    }

    /**
     * Build JSON payload for user events.
     */
    private ObjectNode buildEventPayload(Event event) {
        ObjectNode payload = objectMapper.createObjectNode();

        payload.put("eventId", UUID.randomUUID().toString());
        payload.put("eventType", event.getType().name());
        payload.put("timestamp", Instant.ofEpochMilli(event.getTime()).toString());
        payload.put("realmId", event.getRealmId());

        if (event.getUserId() != null) {
            payload.put("userId", event.getUserId());
        }

        if (event.getSessionId() != null) {
            payload.put("sessionId", event.getSessionId());
        }

        if (event.getIpAddress() != null) {
            payload.put("ipAddress", event.getIpAddress());
        }

        if (event.getClientId() != null) {
            payload.put("clientId", event.getClientId());
        }

        // Add event-specific details
        if (event.getDetails() != null && !event.getDetails().isEmpty()) {
            ObjectNode details = payload.putObject("details");
            for (Map.Entry<String, String> entry : event.getDetails().entrySet()) {
                details.put(entry.getKey(), entry.getValue());
            }

            // Extract tenant_id if present in details
            if (event.getDetails().containsKey("tenant_id")) {
                payload.put("tenantId", event.getDetails().get("tenant_id"));
            }
        }

        if (event.getError() != null) {
            payload.put("error", event.getError());
        }

        return payload;
    }

    /**
     * Build JSON payload for admin events.
     */
    private ObjectNode buildAdminEventPayload(AdminEvent adminEvent, boolean includeRepresentation) {
        ObjectNode payload = objectMapper.createObjectNode();

        payload.put("eventId", UUID.randomUUID().toString());
        payload.put("operationType", adminEvent.getOperationType().name());
        payload.put("timestamp", Instant.ofEpochMilli(adminEvent.getTime()).toString());
        payload.put("realmId", adminEvent.getRealmId());
        payload.put("resourceType", adminEvent.getResourceType().name());
        payload.put("resourcePath", adminEvent.getResourcePath());

        if (adminEvent.getAuthDetails() != null) {
            ObjectNode authDetails = payload.putObject("authDetails");
            authDetails.put("userId", adminEvent.getAuthDetails().getUserId());
            authDetails.put("realmId", adminEvent.getAuthDetails().getRealmId());
            authDetails.put("ipAddress", adminEvent.getAuthDetails().getIpAddress());
        }

        if (includeRepresentation && adminEvent.getRepresentation() != null) {
            payload.put("representation", adminEvent.getRepresentation());
        }

        if (adminEvent.getError() != null) {
            payload.put("error", adminEvent.getError());
        }

        return payload;
    }
}
