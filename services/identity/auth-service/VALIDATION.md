# Phase 1A: Keycloak Auth Service - Validation & Testing Guide

## Overview

This document provides **comprehensive validation and testing procedures** for the Keycloak Authentication Service with custom CotAI extensions. All tests must pass before proceeding to Phase 1B.

## Prerequisites

Ensure the following infrastructure is running:

```bash
# Check all required containers are running
docker ps | grep -E "cotai-(keycloak|kafka|postgres-identity|zookeeper)"

# Expected output (all healthy):
# cotai-keycloak             Up XX minutes (healthy)
# cotai-kafka                Up XX minutes (healthy)
# cotai-postgres-identity    Up XX minutes (healthy)
# cotai-zookeeper            Up XX minutes
```

## Success Criteria

Phase 1A is considered **successfully validated** when:

1. ✅ Custom Keycloak image builds without errors
2. ✅ Keycloak starts and loads custom SPI extensions
3. ✅ Database schema includes user-tenant mapping tables
4. ✅ CotAI realm can be imported successfully
5. ✅ TenantIdMapper injects `tenant_id` into JWT tokens
6. ✅ Kafka Event Listener publishes auth events to Kafka
7. ✅ Authentication flow works end-to-end
8. ✅ No errors in Keycloak logs related to extensions

---

## Validation Tests

### Test 1: Verify Custom Keycloak Extensions Are Loaded

**Objective**: Confirm Keycloak loaded our custom SPI extensions.

**Procedure**:

```bash
# 1. Check extension JAR is present
docker exec cotai-keycloak ls -lh /opt/keycloak/providers/

# Expected: cotai-keycloak-extensions-1.0.0-SNAPSHOT.jar (~15-16MB)

# 2. Check Keycloak startup logs for SPI registration
docker logs cotai-keycloak 2>&1 | grep -i "cotai"

# Expected output (look for these lines):
# "Initializing CotAI Kafka Event Listener Provider"
# "Kafka producer initialized successfully"
```

**Success Criteria**:
- ✅ Extension JAR exists in `/opt/keycloak/providers/`
- ✅ Logs show SPI initialization messages
- ✅ No errors mentioning "cotai" packages

**Troubleshooting**:
- If JAR missing: Rebuild image with `./build.sh`
- If SPI not loading: Check META-INF/services registration files
- If Kafka errors: Verify Kafka is running and accessible

---

### Test 2: Verify Database Migration

**Objective**: Confirm database tables for multi-tenancy exist.

**Procedure**:

```bash
# 1. List tables in public schema
docker exec cotai-postgres-identity psql -U cotai_dev -d cotai_identity \
  -c "\dt public.user_tenant_mapping"

# Expected: Table "public.user_tenant_mapping" displayed with columns

# 2. Verify tenant registry table
docker exec cotai-postgres-identity psql -U cotai_dev -d cotai_identity \
  -c "SELECT tenant_id, tenant_name, status FROM public.tenant_registry;"

# Expected: Default tenant (00000000-0000-0000-0000-000000000000) with status 'active'

# 3. Check indexes
docker exec cotai-postgres-identity psql -U cotai_dev -d cotai_identity \
  -c "\d public.user_tenant_mapping"

# Expected: Multiple indexes including idx_user_tenant_keycloak_user
```

**Success Criteria**:
- ✅ `user_tenant_mapping` table exists with correct schema
- ✅ `tenant_registry` table exists with default tenant
- ✅ All indexes created (at least 5 for user_tenant_mapping)
- ✅ Triggers exist (trigger_user_tenant_mapping_updated_at)

**SQL Schema Verification**:

```sql
-- Run this to verify table structure
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'user_tenant_mapping'
AND table_schema = 'public'
ORDER BY ordinal_position;

-- Expected columns:
-- id, keycloak_user_id, tenant_id, tenant_role, is_primary,
-- is_active, created_at, updated_at, created_by, updated_by, metadata
```

---

### Test 3: Import CotAI Realm

**Objective**: Import the CotAI realm configuration with custom clients and mappers.

**Procedure**:

```bash
# Option A: Import via Admin UI
# 1. Navigate to http://localhost:8080
# 2. Login with admin/admin
# 3. Hover over "master" realm dropdown → "Create Realm"
# 4. Click "Browse" and select: infra/keycloak/realm-cotai.json
# 5. Click "Create"

# Option B: Import via CLI (if available)
docker exec cotai-keycloak /opt/keycloak/bin/kc.sh import \
  --file /opt/keycloak/data/import/realm-cotai.json

# Verify realm was created
curl -s http://localhost:8080/realms/cotai/.well-known/openid-configuration | jq .
```

**Success Criteria**:
- ✅ Realm "cotai" appears in realm selector
- ✅ Realm has 4 clients: cotai-web-app, cotai-mobile-app, cotai-backend-services, cotai-cli
- ✅ Event listener "cotai-kafka-event-listener" is enabled
- ✅ User `admin@cotai.local` exists with password `Admin@123`
- ✅ Roles exist: cotai_admin, cotai_tenant_admin, cotai_user, cotai_viewer

**Verification Commands**:

```bash
# Get admin token
export ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/realms/master/protocol/openid-connect/token \
  -d "client_id=admin-cli" \
  -d "username=admin" \
  -d "password=admin" \
  -d "grant_type=password" | jq -r .access_token)

# List clients in cotai realm
curl -s http://localhost:8080/admin/realms/cotai/clients \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq '.[].clientId'

# Expected output:
# "realm-management"
# "cotai-web-app"
# "cotai-mobile-app"
# "cotai-backend-services"
# "cotai-cli"
# ...
```

---

### Test 4: Verify TenantIdMapper Works

**Objective**: Confirm `tenant_id` is injected into JWT tokens.

**Procedure**:

```bash
# 1. Enable direct access grants on cotai-cli client (if needed)
# Via Admin UI: Clients → cotai-cli → Settings → "Direct access grants enabled" ON

# 2. Get access token for admin@cotai.local
export TOKEN_RESPONSE=$(curl -s -X POST \
  http://localhost:8080/realms/cotai/protocol/openid-connect/token \
  -d "client_id=cotai-cli" \
  -d "username=admin@cotai.local" \
  -d "password=Admin@123" \
  -d "grant_type=password")

echo $TOKEN_RESPONSE | jq .

# 3. Extract and decode access token
export ACCESS_TOKEN=$(echo $TOKEN_RESPONSE | jq -r .access_token)

# Decode JWT (you can use https://jwt.io or this command):
echo $ACCESS_TOKEN | cut -d. -f2 | base64 -d 2>/dev/null | jq .

# 4. Verify tenant_id claim exists
echo $ACCESS_TOKEN | cut -d. -f2 | base64 -d 2>/dev/null | jq '.tenant_id'
```

**Success Criteria**:
- ✅ Token response includes `access_token`, `id_token`, `refresh_token`
- ✅ Decoded access token contains `"tenant_id": "00000000-0000-0000-0000-000000000000"`
- ✅ Token contains `"email": "admin@cotai.local"`
- ✅ Token contains `"roles"` array with `"cotai_admin"` and `"cotai_user"`

**Expected JWT Payload** (partial):

```json
{
  "exp": 1702345678,
  "iat": 1702344778,
  "jti": "...",
  "iss": "http://localhost:8080/realms/cotai",
  "sub": "user-uuid-here",
  "typ": "Bearer",
  "azp": "cotai-cli",
  "email": "admin@cotai.local",
  "email_verified": true,
  "tenant_id": "00000000-0000-0000-0000-000000000000",
  "roles": ["cotai_admin", "cotai_user"],
  "preferred_username": "admin@cotai.local"
}
```

**Troubleshooting**:
- If `tenant_id` missing: Check mapper is configured on client
- If token generation fails: Verify user exists and password is correct
- If direct grants not working: Enable in client settings

---

### Test 5: Verify Kafka Event Publishing

**Objective**: Confirm authentication events are published to Kafka.

**Procedure**:

```bash
# 1. Start Kafka console consumer in background
docker exec -it cotai-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic auth.events \
  --from-beginning &

# 2. Perform authentication (generates LOGIN event)
curl -X POST http://localhost:8080/realms/cotai/protocol/openid-connect/token \
  -d "client_id=cotai-cli" \
  -d "username=admin@cotai.local" \
  -d "password=Admin@123" \
  -d "grant_type=password"

# 3. Wait 5-10 seconds, then check consumer output
# You should see a JSON message like:

{
  "eventId": "uuid",
  "eventType": "LOGIN",
  "timestamp": "2024-12-16T12:00:00Z",
  "realmId": "cotai",
  "userId": "user-uuid",
  "sessionId": "session-uuid",
  "ipAddress": "172.x.x.x",
  "clientId": "cotai-cli",
  "details": {
    "username": "admin@cotai.local",
    "auth_method": "openid-connect",
    ...
  }
}

# 4. Stop consumer (Ctrl+C)

# 5. Check Keycloak logs for Kafka publishing
docker logs cotai-keycloak 2>&1 | grep -i "kafka"

# Expected: "Message sent successfully to auth.events partition X offset Y"
```

**Success Criteria**:
- ✅ Consumer receives JSON event message
- ✅ Event contains `eventType: "LOGIN"`
- ✅ Event includes `userId`, `sessionId`, `ipAddress`
- ✅ Keycloak logs show successful Kafka publishing
- ✅ No Kafka connection errors in logs

**Additional Event Types to Test**:

```bash
# Test LOGOUT event
curl -X POST http://localhost:8080/realms/cotai/protocol/openid-connect/logout \
  -d "client_id=cotai-cli" \
  -d "refresh_token=$REFRESH_TOKEN"

# Test UPDATE_PASSWORD event (via Admin API)
# Requires changing user password via Admin Console or API

# Test REGISTER event (if registration enabled)
```

---

### Test 6: End-to-End Authentication Flow

**Objective**: Validate complete OAuth 2.0 PKCE flow.

**Procedure**:

```bash
# 1. Generate PKCE code verifier and challenge
CODE_VERIFIER=$(openssl rand -base64 64 | tr -d '=' | tr '+/' '-_')
CODE_CHALLENGE=$(echo -n $CODE_VERIFIER | openssl dgst -sha256 -binary | base64 | tr -d '=' | tr '+/' '-_')

# 2. Get authorization code (simulated - in real flow this is via browser)
AUTH_URL="http://localhost:8080/realms/cotai/protocol/openid-connect/auth"
AUTH_URL+="?client_id=cotai-web-app"
AUTH_URL+="&response_type=code"
AUTH_URL+="&redirect_uri=http://localhost:3000/callback"
AUTH_URL+="&code_challenge=$CODE_CHALLENGE"
AUTH_URL+="&code_challenge_method=S256"
AUTH_URL+="&scope=openid email profile"

echo "Visit this URL in browser (login with admin@cotai.local / Admin@123):"
echo $AUTH_URL

# After login, you'll be redirected to: http://localhost:3000/callback?code=XXXX
# Extract the code parameter

# 3. Exchange authorization code for tokens
export AUTH_CODE="<code-from-redirect>"

curl -X POST http://localhost:8080/realms/cotai/protocol/openid-connect/token \
  -d "client_id=cotai-web-app" \
  -d "grant_type=authorization_code" \
  -d "code=$AUTH_CODE" \
  -d "redirect_uri=http://localhost:3000/callback" \
  -d "code_verifier=$CODE_VERIFIER" | jq .

# 4. Verify tokens received
# 5. Test token refresh

export REFRESH_TOKEN="<refresh-token-from-step-3>"

curl -X POST http://localhost:8080/realms/cotai/protocol/openid-connect/token \
  -d "client_id=cotai-web-app" \
  -d "grant_type=refresh_token" \
  -d "refresh_token=$REFRESH_TOKEN" | jq .
```

**Success Criteria**:
- ✅ Authorization URL opens login page
- ✅ User can login with correct credentials
- ✅ Redirect contains authorization code
- ✅ Code exchange returns access_token, id_token, refresh_token
- ✅ All tokens contain `tenant_id` claim
- ✅ Token refresh works without re-authentication
- ✅ All events published to Kafka (LOGIN, REFRESH_TOKEN)

---

### Test 7: Multi-User Tenant Mapping

**Objective**: Verify user-tenant associations work correctly.

**Procedure**:

```bash
# 1. Create a new test user via Keycloak Admin API or UI
# Via UI: Users → Add User
#   - Username: test@cotai.local
#   - Email: test@cotai.local
#   - Email Verified: ON
#   - Save → Credentials → Set Password: Test@456

# 2. Get the user's UUID
export USER_UUID=$(curl -s http://localhost:8080/admin/realms/cotai/users?username=test@cotai.local \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.[0].id')

echo "User UUID: $USER_UUID"

# 3. Add tenant_id attribute to user
curl -X PUT http://localhost:8080/admin/realms/cotai/users/$USER_UUID \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "attributes": {
      "tenant_id": ["00000000-0000-0000-0000-000000000000"]
    }
  }'

# 4. Insert user-tenant mapping in database
docker exec cotai-postgres-identity psql -U cotai_dev -d cotai_identity <<EOF
INSERT INTO public.user_tenant_mapping (
    keycloak_user_id,
    tenant_id,
    tenant_role,
    is_primary,
    is_active
) VALUES (
    '$USER_UUID'::uuid,
    '00000000-0000-0000-0000-000000000000'::uuid,
    'tenant_user',
    true,
    true
);
EOF

# 5. Verify mapping
docker exec cotai-postgres-identity psql -U cotai_dev -d cotai_identity \
  -c "SELECT * FROM public.user_tenant_mapping WHERE keycloak_user_id='$USER_UUID'::uuid;"

# 6. Test authentication with new user
curl -X POST http://localhost:8080/realms/cotai/protocol/openid-connect/token \
  -d "client_id=cotai-cli" \
  -d "username=test@cotai.local" \
  -d "password=Test@456" \
  -d "grant_type=password" | jq .

# 7. Verify tenant_id in token
# (decode access_token as in Test 4)
```

**Success Criteria**:
- ✅ New user created successfully
- ✅ User-tenant mapping inserted without errors
- ✅ User can authenticate
- ✅ Token contains correct `tenant_id`
- ✅ Database query shows mapping with `is_primary=true`

---

### Test 8: Error Handling and Edge Cases

**Objective**: Verify system handles errors gracefully.

**Test Cases**:

#### 8.1: User Without tenant_id Attribute

```bash
# Create user without tenant_id
# Expected: Login succeeds but token has no tenant_id claim
# Expected: Keycloak logs warn about missing tenant_id
```

#### 8.2: Invalid Credentials

```bash
curl -X POST http://localhost:8080/realms/cotai/protocol/openid-connect/token \
  -d "client_id=cotai-cli" \
  -d "username=admin@cotai.local" \
  -d "password=WrongPassword" \
  -d "grant_type=password"

# Expected: HTTP 401 Unauthorized
# Expected: LOGIN_ERROR event published to Kafka
```

#### 8.3: Kafka Unavailable

```bash
# Stop Kafka
docker stop cotai-kafka

# Perform login
# Expected: Login still succeeds (non-blocking)
# Expected: Keycloak logs error about Kafka connection
# Expected: No crash or authentication failure

# Restart Kafka
docker start cotai-kafka
```

**Success Criteria**:
- ✅ Missing tenant_id doesn't break authentication
- ✅ Invalid credentials return proper error
- ✅ Kafka unavailability doesn't block auth flow
- ✅ All errors logged appropriately

---

## Performance Validation

### Test 9: Load Testing (Optional but Recommended)

**Objective**: Verify performance under load.

**Procedure** (requires `k6` or `ab`):

```bash
# Install k6 (if not installed)
# https://k6.io/docs/getting-started/installation/

# Run load test script
cat <<'EOF' > load-test.js
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 users
    { duration: '1m', target: 10 },   // Stay at 10 users
    { duration: '30s', target: 0 },   // Ramp down
  ],
};

export default function () {
  const payload = {
    client_id: 'cotai-cli',
    username: 'admin@cotai.local',
    password: 'Admin@123',
    grant_type: 'password',
  };

  const res = http.post(
    'http://localhost:8080/realms/cotai/protocol/openid-connect/token',
    payload
  );

  check(res, {
    'status is 200': (r) => r.status === 200,
    'has access_token': (r) => r.json('access_token') !== undefined,
    'has tenant_id in token': (r) => {
      const token = r.json('access_token');
      const payload = JSON.parse(atob(token.split('.')[1]));
      return payload.tenant_id !== undefined;
    },
  });
}
EOF

k6 run load-test.js
```

**Success Criteria**:
- ✅ P95 response time < 500ms
- ✅ 100% success rate for token issuance
- ✅ No memory leaks (check docker stats)
- ✅ All tokens contain tenant_id

---

## Final Checklist

Before marking Phase 1A as complete:

- [ ] All 8 validation tests passed
- [ ] Keycloak logs show no errors related to custom extensions
- [ ] Kafka topics `auth.events` and `auth.admin.events` exist
- [ ] Database tables `user_tenant_mapping` and `tenant_registry` populated
- [ ] JWT tokens consistently contain `tenant_id` claim
- [ ] Documentation updated (README.md, VALIDATION.md)
- [ ] Docker image `cotai-keycloak-custom:latest` built and tagged
- [ ] docker-compose.dev.yml updated to use custom image
- [ ] At least one test user with tenant mapping created

---

## Troubleshooting Guide

### Issue: Extension JAR not loading

**Symptoms**: No SPI initialization messages in logs

**Solution**:
1. Verify JAR exists: `docker exec cotai-keycloak ls /opt/keycloak/providers/`
2. Check permissions: `docker exec cotai-keycloak ls -la /opt/keycloak/providers/`
3. Rebuild image: `cd services/identity/auth-service && ./build.sh`
4. Restart container: `docker compose restart keycloak`

### Issue: tenant_id not in JWT

**Symptoms**: Token decoding shows no tenant_id claim

**Solution**:
1. Verify user has tenant_id attribute in Keycloak
2. Check client has TenantIdMapper configured
3. Review Keycloak logs for mapper errors
4. Test with direct grants (password flow) first

### Issue: Kafka events not publishing

**Symptoms**: No messages in Kafka consumer

**Solution**:
1. Verify Kafka is running: `docker ps | grep kafka`
2. Check Keycloak can reach Kafka: `docker exec cotai-keycloak ping kafka`
3. Verify event listener enabled: Check realm settings → Events
4. Review Keycloak logs for Kafka errors
5. Test Kafka independently: `docker exec cotai-kafka kafka-topics --list --bootstrap-server localhost:9092`

### Issue: Database migration failed

**Symptoms**: Tables don't exist

**Solution**:
1. Re-run migration: `docker exec -i cotai-postgres-identity psql -U cotai_dev -d cotai_identity < infra/sql/init-identity.sql`
2. Check for errors in output
3. Verify PostgreSQL is healthy: `docker logs cotai-postgres-identity`

---

## Success Confirmation

**Phase 1A is VALIDATED** when you can execute this complete flow:

```bash
# 1. Get token with tenant_id
TOKEN=$(curl -s -X POST http://localhost:8080/realms/cotai/protocol/openid-connect/token \
  -d "client_id=cotai-cli" \
  -d "username=admin@cotai.local" \
  -d "password=Admin@123" \
  -d "grant_type=password" | jq -r .access_token)

# 2. Decode and verify tenant_id
echo $TOKEN | cut -d. -f2 | base64 -d 2>/dev/null | jq .tenant_id

# Output: "00000000-0000-0000-0000-000000000000"

# 3. Verify event in Kafka (within 10 seconds)
docker exec cotai-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic auth.events \
  --max-messages 1 \
  --from-beginning \
  --timeout-ms 10000 | jq .eventType

# Output: "LOGIN"

# If both outputs are correct: Phase 1A VALIDATED ✅
```

---

## Next Steps

After validation:
1. Document any deviations or issues encountered
2. Commit all changes to version control
3. Tag release: `git tag v1.0.0-phase1a`
4. Proceed to **Phase 1B: Tenant Manager Service**
5. Share validation results with team

## Support

For issues during validation:
- Check logs: `docker logs cotai-keycloak`
- Review README: `services/identity/auth-service/README.md`
- Consult architecture docs: `docs/architecture/multi-tenancy.md`
