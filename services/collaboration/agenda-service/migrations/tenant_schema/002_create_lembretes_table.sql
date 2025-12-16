-- Create lembretes table for Agenda Service
CREATE TABLE IF NOT EXISTS lembretes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  evento_id UUID NOT NULL REFERENCES eventos(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL,
  offset_minutes INT NOT NULL,
  channel VARCHAR(20) DEFAULT 'push',
  status VARCHAR(20) DEFAULT 'pendente',
  due_at TIMESTAMPTZ NOT NULL,
  sent_at TIMESTAMPTZ,
  error TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),

  -- Constraints
  CONSTRAINT check_tenant_id_not_null CHECK (tenant_id IS NOT NULL),
  CONSTRAINT check_offset_positive CHECK (offset_minutes > 0),
  CONSTRAINT check_status_valid CHECK (status IN ('pendente', 'enviado', 'falhou'))
);

-- Indexes
CREATE INDEX idx_lembretes_tenant_id ON lembretes(tenant_id);
CREATE INDEX idx_lembretes_evento_id ON lembretes(evento_id);
CREATE INDEX idx_lembretes_due_status ON lembretes(status, due_at) WHERE status = 'pendente';

-- Comments
COMMENT ON TABLE lembretes IS 'Reminder notifications for eventos';
COMMENT ON COLUMN lembretes.offset_minutes IS 'Minutes before evento.inicio to send reminder';
COMMENT ON COLUMN lembretes.channel IS 'Notification channel (push, email, sms)';
COMMENT ON COLUMN lembretes.status IS 'Reminder status (pendente, enviado, falhou)';
COMMENT ON COLUMN lembretes.due_at IS 'Computed timestamp when reminder should be sent';
