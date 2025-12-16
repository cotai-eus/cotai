-- CotAI Identity Database Initialization Script
-- PostgreSQL 15.4
-- Database: cotai_identity
-- Purpose: Keycloak authentication + User-Tenant mapping

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create schema for default tenant (for development)
CREATE SCHEMA IF NOT EXISTS tenant_default;

-- Set default search path
ALTER DATABASE cotai_identity SET search_path TO public, tenant_default;

-- Grant permissions
GRANT ALL PRIVILEGES ON SCHEMA public TO cotai_dev;
GRANT ALL PRIVILEGES ON SCHEMA tenant_default TO cotai_dev;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO cotai_dev;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA tenant_default TO cotai_dev;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO cotai_dev;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA tenant_default TO cotai_dev;

-- ============================================================================
-- User-Tenant Mapping Table
-- ============================================================================
-- Maps Keycloak user UUIDs to CotAI tenant UUIDs
-- Supports multi-tenancy where users can belong to multiple tenants
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.user_tenant_mapping (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Keycloak user ID (from Keycloak's database)
    keycloak_user_id UUID NOT NULL,

    -- CotAI tenant ID (matches tenant schema naming: tenant_{uuid})
    tenant_id UUID NOT NULL,

    -- User's role within this specific tenant
    -- Maps to Keycloak realm roles but scoped per-tenant
    tenant_role VARCHAR(50) NOT NULL CHECK (tenant_role IN (
        'tenant_admin',     -- Full control within tenant
        'tenant_manager',   -- Can manage resources and users
        'tenant_user',      -- Standard user permissions
        'tenant_viewer'     -- Read-only access
    )),

    -- Is this the user's primary tenant? (used for default selection)
    is_primary BOOLEAN NOT NULL DEFAULT false,

    -- Status flags
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Audit timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Who created/modified this mapping (for audit trail)
    created_by UUID, -- Can be NULL for system-created
    updated_by UUID,

    -- Additional metadata (JSON for flexibility)
    metadata JSONB DEFAULT '{}'::jsonb,

    -- Constraints
    CONSTRAINT unique_user_tenant UNIQUE (keycloak_user_id, tenant_id),
    CONSTRAINT check_single_primary_per_user EXCLUDE USING btree (
        keycloak_user_id WITH =,
        is_primary WITH =
    ) WHERE (is_primary = true)
);

-- Indexes for performance
CREATE INDEX idx_user_tenant_keycloak_user ON public.user_tenant_mapping(keycloak_user_id)
    WHERE is_active = true;

CREATE INDEX idx_user_tenant_tenant_id ON public.user_tenant_mapping(tenant_id)
    WHERE is_active = true;

CREATE INDEX idx_user_tenant_primary ON public.user_tenant_mapping(keycloak_user_id, is_primary)
    WHERE is_primary = true AND is_active = true;

CREATE INDEX idx_user_tenant_role ON public.user_tenant_mapping(tenant_id, tenant_role)
    WHERE is_active = true;

CREATE INDEX idx_user_tenant_metadata ON public.user_tenant_mapping USING GIN (metadata);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_user_tenant_mapping_updated_at
    BEFORE UPDATE ON public.user_tenant_mapping
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Tenant Registry Table
-- ============================================================================
-- Central registry of all tenants in the system
-- Manages tenant lifecycle and schema provisioning
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.tenant_registry (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Tenant identification
    tenant_id UUID NOT NULL UNIQUE,
    tenant_name VARCHAR(255) NOT NULL UNIQUE,
    tenant_slug VARCHAR(100) NOT NULL UNIQUE, -- URL-safe identifier

    -- Subscription/Plan information
    plan_tier VARCHAR(50) NOT NULL CHECK (plan_tier IN (
        'free',
        'basic',
        'professional',
        'enterprise'
    )) DEFAULT 'free',

    -- Status
    status VARCHAR(50) NOT NULL CHECK (status IN (
        'provisioning',   -- Schema being created
        'active',         -- Fully operational
        'suspended',      -- Temporarily disabled (billing issue, etc.)
        'archived',       -- Soft-deleted, data retained
        'deleted'         -- Marked for deletion, data will be purged
    )) DEFAULT 'provisioning',

    -- Schema information
    database_schema VARCHAR(100) NOT NULL UNIQUE, -- e.g., tenant_550e8400e29b41d4a716446655440000
    schema_version VARCHAR(20) NOT NULL DEFAULT '1.0.0',

    -- Limits and quotas (enforced by API gateway)
    max_users INTEGER NOT NULL DEFAULT 5,
    max_storage_gb INTEGER NOT NULL DEFAULT 10,

    -- Contact information
    primary_contact_email VARCHAR(255),
    primary_contact_name VARCHAR(255),

    -- Billing
    billing_email VARCHAR(255),

    -- Audit timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    activated_at TIMESTAMP WITH TIME ZONE,
    suspended_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Audit actors
    created_by UUID,
    updated_by UUID,

    -- Additional configuration
    settings JSONB DEFAULT '{}'::jsonb,

    -- Feature flags per tenant
    features JSONB DEFAULT '{}'::jsonb
);

-- Indexes
CREATE INDEX idx_tenant_registry_status ON public.tenant_registry(status)
    WHERE status IN ('active', 'provisioning');

CREATE INDEX idx_tenant_registry_slug ON public.tenant_registry(tenant_slug);

CREATE INDEX idx_tenant_registry_plan ON public.tenant_registry(plan_tier, status);

CREATE INDEX idx_tenant_registry_settings ON public.tenant_registry USING GIN (settings);

CREATE INDEX idx_tenant_registry_features ON public.tenant_registry USING GIN (features);

-- Trigger for updated_at
CREATE TRIGGER trigger_tenant_registry_updated_at
    BEFORE UPDATE ON public.tenant_registry
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Seed Data for Development
-- ============================================================================

-- Insert default tenant for development
INSERT INTO public.tenant_registry (
    tenant_id,
    tenant_name,
    tenant_slug,
    database_schema,
    status,
    plan_tier,
    max_users,
    max_storage_gb,
    primary_contact_email,
    primary_contact_name,
    settings
) VALUES (
    '00000000-0000-0000-0000-000000000000'::uuid,
    'Default Development Tenant',
    'default-dev',
    'tenant_default',
    'active',
    'enterprise',
    999,
    1000,
    'admin@cotai.local',
    'Admin CotAI',
    '{"environment": "development", "debug": true}'::jsonb
) ON CONFLICT (tenant_id) DO NOTHING;

-- ============================================================================
-- Comments for Documentation
-- ============================================================================

COMMENT ON DATABASE cotai_identity IS 'CotAI Identity & Auth Database - Multi-tenant isolation via schemas';

COMMENT ON TABLE public.user_tenant_mapping IS
'Maps Keycloak users to CotAI tenants. Supports multi-tenant users with role-based access per tenant.';

COMMENT ON TABLE public.tenant_registry IS
'Central registry of all tenants. Tracks lifecycle, schema provisioning, and subscription details.';

COMMENT ON COLUMN public.user_tenant_mapping.keycloak_user_id IS
'References user ID from Keycloak database (external FK, not enforced)';

COMMENT ON COLUMN public.user_tenant_mapping.tenant_id IS
'References tenant_id in tenant_registry. Tenant data lives in schema tenant_{uuid}';

COMMENT ON COLUMN public.user_tenant_mapping.is_primary IS
'Only ONE primary tenant per user allowed via exclusion constraint';

COMMENT ON COLUMN public.tenant_registry.database_schema IS
'PostgreSQL schema name where tenant data resides. Format: tenant_{uuid without hyphens}';
