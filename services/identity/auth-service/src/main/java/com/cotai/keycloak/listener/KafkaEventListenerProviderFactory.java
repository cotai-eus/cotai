package com.cotai.keycloak.listener;

import com.cotai.keycloak.config.KafkaProducerManager;
import org.jboss.logging.Logger;
import org.keycloak.Config;
import org.keycloak.events.EventListenerProvider;
import org.keycloak.events.EventListenerProviderFactory;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.KeycloakSessionFactory;

/**
 * Factory for creating KafkaEventListenerProvider instances.
 *
 * This factory initializes the Kafka producer on startup and creates
 * event listener instances for each Keycloak session.
 *
 * Configuration is loaded from environment variables:
 * - KAFKA_BOOTSTRAP_SERVERS (default: localhost:9092)
 * - KAFKA_PRODUCER_CLIENT_ID (default: keycloak-event-publisher)
 *
 * @author CotAI Development Team
 * @version 1.0.0
 */
public class KafkaEventListenerProviderFactory implements EventListenerProviderFactory {

    private static final Logger logger = Logger.getLogger(KafkaEventListenerProviderFactory.class);

    public static final String PROVIDER_ID = "cotai-kafka-event-listener";

    @Override
    public EventListenerProvider create(KeycloakSession session) {
        return new KafkaEventListenerProvider();
    }

    @Override
    public void init(Config.Scope config) {
        logger.info("Initializing CotAI Kafka Event Listener Provider");

        try {
            // Initialize Kafka producer (singleton pattern)
            KafkaProducerManager.getInstance();
            logger.info("Kafka producer initialized successfully");
        } catch (Exception e) {
            logger.error("Failed to initialize Kafka producer. Events will not be published.", e);
            // Don't throw - allow Keycloak to start even if Kafka is unavailable
        }
    }

    @Override
    public void postInit(KeycloakSessionFactory factory) {
        logger.info("CotAI Kafka Event Listener Provider post-initialization complete");
    }

    @Override
    public void close() {
        logger.info("Shutting down CotAI Kafka Event Listener Provider");

        try {
            KafkaProducerManager.getInstance().close();
            logger.info("Kafka producer closed successfully");
        } catch (Exception e) {
            logger.error("Error closing Kafka producer", e);
        }
    }

    @Override
    public String getId() {
        return PROVIDER_ID;
    }
}
