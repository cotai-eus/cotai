-- Disable Row-Level Security

-- Disable RLS on audit_logs
DROP POLICY IF EXISTS tenant_isolation_select_audit_logs ON audit_logs;
DROP POLICY IF EXISTS tenant_isolation_insert_audit_logs ON audit_logs;
ALTER TABLE audit_logs DISABLE ROW LEVEL SECURITY;

-- Disable RLS on cotacoes
DROP POLICY IF EXISTS tenant_isolation_select_cotacoes ON cotacoes;
DROP POLICY IF EXISTS tenant_isolation_insert_cotacoes ON cotacoes;
DROP POLICY IF EXISTS tenant_isolation_update_cotacoes ON cotacoes;
DROP POLICY IF EXISTS tenant_isolation_delete_cotacoes ON cotacoes;
ALTER TABLE cotacoes DISABLE ROW LEVEL SECURITY;

-- Disable RLS on fornecedores
DROP POLICY IF EXISTS tenant_isolation_select_fornecedores ON fornecedores;
DROP POLICY IF EXISTS tenant_isolation_insert_fornecedores ON fornecedores;
DROP POLICY IF EXISTS tenant_isolation_update_fornecedores ON fornecedores;
DROP POLICY IF EXISTS tenant_isolation_delete_fornecedores ON fornecedores;
ALTER TABLE fornecedores DISABLE ROW LEVEL SECURITY;

-- Disable RLS on licitacoes
DROP POLICY IF EXISTS tenant_isolation_select_licitacoes ON licitacoes;
DROP POLICY IF EXISTS tenant_isolation_insert_licitacoes ON licitacoes;
DROP POLICY IF EXISTS tenant_isolation_update_licitacoes ON licitacoes;
DROP POLICY IF EXISTS tenant_isolation_delete_licitacoes ON licitacoes;
ALTER TABLE licitacoes DISABLE ROW LEVEL SECURITY;
