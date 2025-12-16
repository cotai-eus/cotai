-- Enable Row-Level Security for Agenda Service tables

-- Enable RLS on eventos table
ALTER TABLE eventos ENABLE ROW LEVEL SECURITY;

-- Create RLS policies for eventos
CREATE POLICY tenant_isolation_select ON eventos
  FOR SELECT
  USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY tenant_isolation_insert ON eventos
  FOR INSERT
  WITH CHECK (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY tenant_isolation_update ON eventos
  FOR UPDATE
  USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY tenant_isolation_delete ON eventos
  FOR DELETE
  USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

-- Enable RLS on lembretes table
ALTER TABLE lembretes ENABLE ROW LEVEL SECURITY;

-- Create RLS policies for lembretes
CREATE POLICY tenant_isolation_select ON lembretes
  FOR SELECT
  USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY tenant_isolation_insert ON lembretes
  FOR INSERT
  WITH CHECK (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY tenant_isolation_update ON lembretes
  FOR UPDATE
  USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY tenant_isolation_delete ON lembretes
  FOR DELETE
  USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

-- Comments
COMMENT ON POLICY tenant_isolation_select ON eventos IS 'Ensure users can only select eventos from their tenant';
COMMENT ON POLICY tenant_isolation_select ON lembretes IS 'Ensure users can only select lembretes from their tenant';
