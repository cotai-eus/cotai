-- Initial schema for tenant
-- This creates the base tables that every tenant needs
-- All tables MUST include tenant_id column for RLS

-- Example: Licitacoes table (procurement notices)
CREATE TABLE IF NOT EXISTS licitacoes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    numero VARCHAR(50) NOT NULL,
    objeto TEXT,
    status VARCHAR(20) DEFAULT 'RECEBIDO',
    modalidade VARCHAR(50),
    valor_estimado DECIMAL(15,2),
    data_abertura TIMESTAMPTZ,
    data_entrega TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT check_tenant_id CHECK (tenant_id IS NOT NULL)
);

CREATE INDEX idx_licitacoes_tenant ON licitacoes(tenant_id);
CREATE INDEX idx_licitacoes_status ON licitacoes(status);
CREATE INDEX idx_licitacoes_numero ON licitacoes(numero);

-- Example: Fornecedores table (suppliers)
CREATE TABLE IF NOT EXISTS fornecedores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    razao_social VARCHAR(255) NOT NULL,
    cnpj VARCHAR(18) UNIQUE,
    email VARCHAR(255),
    telefone VARCHAR(20),
    endereco TEXT,
    ativo BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT check_tenant_id CHECK (tenant_id IS NOT NULL)
);

CREATE INDEX idx_fornecedores_tenant ON fornecedores(tenant_id);
CREATE INDEX idx_fornecedores_cnpj ON fornecedores(cnpj);

-- Example: Cotacoes table (quotes)
CREATE TABLE IF NOT EXISTS cotacoes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    licitacao_id UUID REFERENCES licitacoes(id) ON DELETE CASCADE,
    fornecedor_id UUID REFERENCES fornecedores(id) ON DELETE CASCADE,
    valor_total DECIMAL(15,2),
    prazo_entrega INTEGER, -- days
    observacoes TEXT,
    status VARCHAR(20) DEFAULT 'PENDENTE',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT check_tenant_id CHECK (tenant_id IS NOT NULL)
);

CREATE INDEX idx_cotacoes_tenant ON cotacoes(tenant_id);
CREATE INDEX idx_cotacoes_licitacao ON cotacoes(licitacao_id);
CREATE INDEX idx_cotacoes_fornecedor ON cotacoes(fornecedor_id);

-- Example: Audit log table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID,
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    changes JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT check_tenant_id CHECK (tenant_id IS NOT NULL)
);

CREATE INDEX idx_audit_logs_tenant ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
