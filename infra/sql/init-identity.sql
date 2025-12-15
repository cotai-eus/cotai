-- CotAI Identity Database Initialization Script
-- PostgreSQL 15.4

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

-- Placeholder tables (will be created by migrations)
COMMENT ON DATABASE cotai_identity IS 'CotAI Identity & Auth Database - Multi-tenant isolation via schemas';
