-- Create eventos table for Agenda Service
CREATE TABLE IF NOT EXISTS eventos (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  titulo TEXT NOT NULL,
  descricao TEXT,
  inicio TIMESTAMPTZ NOT NULL,
  fim TIMESTAMPTZ,
  licitacao_id UUID,
  tipo VARCHAR(30),
  created_by UUID,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),

  -- Constraints
  CONSTRAINT check_tenant_id_not_null CHECK (tenant_id IS NOT NULL),
  CONSTRAINT check_inicio_before_fim CHECK (fim IS NULL OR inicio < fim)
);

-- Indexes
CREATE INDEX idx_eventos_tenant_id ON eventos(tenant_id);
CREATE INDEX idx_eventos_tenant_inicio ON eventos(tenant_id, inicio);
CREATE INDEX idx_eventos_licitacao_id ON eventos(licitacao_id) WHERE licitacao_id IS NOT NULL;
CREATE INDEX idx_eventos_tipo ON eventos(tipo) WHERE tipo IS NOT NULL;

-- Comments
COMMENT ON TABLE eventos IS 'Calendar events for tenants';
COMMENT ON COLUMN eventos.tenant_id IS 'Tenant identifier for multi-tenancy';
COMMENT ON COLUMN eventos.licitacao_id IS 'Optional reference to related licitacao';
COMMENT ON COLUMN eventos.tipo IS 'Event type (reuniao, prazo_ocr, prazo_cotacao, entrega, etc.)';
