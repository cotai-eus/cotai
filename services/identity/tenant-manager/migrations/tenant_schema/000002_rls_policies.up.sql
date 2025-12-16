-- Enable Row-Level Security on all tables
-- This ensures tenant isolation at the database level

-- Enable RLS on licitacoes
ALTER TABLE licitacoes ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_select_licitacoes ON licitacoes
    FOR SELECT
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY tenant_isolation_insert_licitacoes ON licitacoes
    FOR INSERT
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY tenant_isolation_update_licitacoes ON licitacoes
    FOR UPDATE
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY tenant_isolation_delete_licitacoes ON licitacoes
    FOR DELETE
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

-- Enable RLS on fornecedores
ALTER TABLE fornecedores ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_select_fornecedores ON fornecedores
    FOR SELECT
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY tenant_isolation_insert_fornecedores ON fornecedores
    FOR INSERT
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY tenant_isolation_update_fornecedores ON fornecedores
    FOR UPDATE
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY tenant_isolation_delete_fornecedores ON fornecedores
    FOR DELETE
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

-- Enable RLS on cotacoes
ALTER TABLE cotacoes ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_select_cotacoes ON cotacoes
    FOR SELECT
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY tenant_isolation_insert_cotacoes ON cotacoes
    FOR INSERT
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY tenant_isolation_update_cotacoes ON cotacoes
    FOR UPDATE
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY tenant_isolation_delete_cotacoes ON cotacoes
    FOR DELETE
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

-- Enable RLS on audit_logs
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_select_audit_logs ON audit_logs
    FOR SELECT
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY tenant_isolation_insert_audit_logs ON audit_logs
    FOR INSERT
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true)::uuid);

-- Note: No UPDATE or DELETE policies for audit_logs (immutable)
