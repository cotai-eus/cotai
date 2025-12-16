package com.cotai.keycloak.mapper;

import org.jboss.logging.Logger;
import org.keycloak.models.ClientSessionContext;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.ProtocolMapperModel;
import org.keycloak.models.UserModel;
import org.keycloak.models.UserSessionModel;
import org.keycloak.protocol.oidc.mappers.AbstractOIDCProtocolMapper;
import org.keycloak.protocol.oidc.mappers.OIDCAccessTokenMapper;
import org.keycloak.protocol.oidc.mappers.OIDCAttributeMapperHelper;
import org.keycloak.protocol.oidc.mappers.OIDCIDTokenMapper;
import org.keycloak.protocol.oidc.mappers.UserInfoTokenMapper;
import org.keycloak.provider.ProviderConfigProperty;
import org.keycloak.representations.IDToken;

import java.util.ArrayList;
import java.util.List;

/**
 * Custom Keycloak Protocol Mapper for injecting tenant_id into JWT claims.
 *
 * This mapper reads the tenant_id from the user's attributes and adds it to:
 * - Access Token
 * - ID Token
 * - UserInfo endpoint response
 *
 * Required for CotAI's schema-per-tenant multi-tenancy architecture.
 *
 * @author CotAI Development Team
 * @version 1.0.0
 */
public class TenantIdMapper extends AbstractOIDCProtocolMapper
        implements OIDCAccessTokenMapper, OIDCIDTokenMapper, UserInfoTokenMapper {

    private static final Logger logger = Logger.getLogger(TenantIdMapper.class);

    public static final String PROVIDER_ID = "cotai-tenant-id-mapper";
    public static final String TENANT_ID_CLAIM = "tenant_id";
    public static final String USER_ATTRIBUTE_TENANT_ID = "tenant_id";

    private static final List<ProviderConfigProperty> configProperties = new ArrayList<>();

    static {
        // Define the claim name property
        OIDCAttributeMapperHelper.addTokenClaimNameConfig(configProperties);

        // Define which tokens to include this claim in
        OIDCAttributeMapperHelper.addIncludeInTokensConfig(configProperties, TenantIdMapper.class);
    }

    @Override
    public String getDisplayCategory() {
        return TOKEN_MAPPER_CATEGORY;
    }

    @Override
    public String getDisplayType() {
        return "CotAI Tenant ID Mapper";
    }

    @Override
    public String getHelpText() {
        return "Maps the user's tenant_id attribute to a JWT claim. Required for multi-tenant isolation.";
    }

    @Override
    public List<ProviderConfigProperty> getConfigProperties() {
        return configProperties;
    }

    @Override
    public String getId() {
        return PROVIDER_ID;
    }

    /**
     * Sets the tenant_id claim in the token.
     *
     * @param token The token being generated
     * @param mappingModel The mapper configuration
     * @param userSession The user's session
     * @param keycloakSession The Keycloak session
     * @param clientSessionCtx The client session context
     */
    @Override
    protected void setClaim(IDToken token, ProtocolMapperModel mappingModel,
                          UserSessionModel userSession, KeycloakSession keycloakSession,
                          ClientSessionContext clientSessionCtx) {

        UserModel user = userSession.getUser();

        if (user == null) {
            logger.warnf("User is null in session %s, cannot set tenant_id claim", userSession.getId());
            return;
        }

        // Get tenant_id from user attributes
        List<String> tenantIds = user.getAttributeStream(USER_ATTRIBUTE_TENANT_ID).toList();

        if (tenantIds.isEmpty()) {
            logger.warnf("User %s (id: %s) has no tenant_id attribute. " +
                        "This user cannot access tenant-specific resources.",
                        user.getUsername(), user.getId());
            return;
        }

        // For multi-tenant users, use the first tenant_id
        // Future enhancement: Support tenant switching via additional claim
        String tenantId = tenantIds.get(0);

        if (tenantIds.size() > 1) {
            logger.debugf("User %s has %d tenant associations. Using tenant_id: %s",
                         user.getUsername(), tenantIds.size(), tenantId);
        }

        // Get the claim name from mapper configuration (default: "tenant_id")
        String claimName = mappingModel.getConfig()
                .getOrDefault(OIDCAttributeMapperHelper.TOKEN_CLAIM_NAME, TENANT_ID_CLAIM);

        // Set the claim in the token
        token.getOtherClaims().put(claimName, tenantId);

        logger.debugf("Successfully set %s=%s for user %s in token",
                     claimName, tenantId, user.getUsername());
    }

    /**
     * Priority for this mapper. Lower values = higher priority.
     * Set to 50 to run after standard mappers but before custom business logic mappers.
     */
    @Override
    public int getPriority() {
        return 50;
    }
}
