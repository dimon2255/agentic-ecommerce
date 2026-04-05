-- ============================================================
-- Migration 00010: RBAC, Audit Log, and Report Views
-- ============================================================

-- ============================================================
-- RBAC Tables
-- ============================================================

CREATE TABLE roles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT UNIQUE NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE permissions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  key TEXT UNIQUE NOT NULL,
  description TEXT,
  grp TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE role_permissions (
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE user_roles (
  user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  granted_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
  PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);
CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);

-- ============================================================
-- RLS: service_role only on all RBAC tables
-- ============================================================

ALTER TABLE roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE permissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE role_permissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;

CREATE POLICY "roles_service_all" ON roles
  FOR ALL USING (auth.role() = 'service_role');

CREATE POLICY "permissions_service_all" ON permissions
  FOR ALL USING (auth.role() = 'service_role');

CREATE POLICY "role_permissions_service_all" ON role_permissions
  FOR ALL USING (auth.role() = 'service_role');

CREATE POLICY "user_roles_service_all" ON user_roles
  FOR ALL USING (auth.role() = 'service_role');

-- ============================================================
-- Seed permissions
-- ============================================================

INSERT INTO permissions (key, description, grp) VALUES
  ('catalog:read',   'View catalog data in admin',            'catalog'),
  ('catalog:write',  'Create, update, delete catalog items',  'catalog'),
  ('orders:read',    'View orders in admin',                  'orders'),
  ('orders:write',   'Update order status',                   'orders'),
  ('users:read',     'View user profiles',                    'users'),
  ('users:write',    'Manage user roles',                     'users'),
  ('reports:read',   'View reports and dashboards',           'reports'),
  ('settings:read',  'View site settings',                    'settings'),
  ('settings:write', 'Modify site settings',                  'settings'),
  ('audit:read',     'View audit log',                        'audit');

-- ============================================================
-- Seed roles
-- ============================================================

INSERT INTO roles (name, description) VALUES
  ('admin',    'Full access to all admin features'),
  ('customer', 'Regular customer with no admin access');

-- Grant all permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin';

-- ============================================================
-- RPC: get user permissions (called by Go middleware)
-- ============================================================

CREATE OR REPLACE FUNCTION get_user_permissions(p_user_id UUID)
RETURNS TEXT[]
LANGUAGE sql STABLE SECURITY DEFINER
AS $$
  SELECT COALESCE(
    array_agg(DISTINCT p.key),
    ARRAY[]::TEXT[]
  )
  FROM user_roles ur
  JOIN role_permissions rp ON rp.role_id = ur.role_id
  JOIN permissions p ON p.id = rp.permission_id
  WHERE ur.user_id = p_user_id;
$$;

-- ============================================================
-- Audit log
-- ============================================================

CREATE TABLE admin_audit_log (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  action TEXT NOT NULL,
  resource_type TEXT NOT NULL,
  resource_id TEXT,
  changes JSONB,
  ip_address TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_log_user ON admin_audit_log(user_id);
CREATE INDEX idx_audit_log_action ON admin_audit_log(action);
CREATE INDEX idx_audit_log_resource ON admin_audit_log(resource_type, resource_id);
CREATE INDEX idx_audit_log_created ON admin_audit_log(created_at DESC);

ALTER TABLE admin_audit_log ENABLE ROW LEVEL SECURITY;

CREATE POLICY "audit_log_service_all" ON admin_audit_log
  FOR ALL USING (auth.role() = 'service_role');

-- ============================================================
-- Report views
-- ============================================================

CREATE OR REPLACE VIEW admin_dashboard_kpis AS
SELECT
  (SELECT COUNT(*) FROM orders WHERE status NOT IN ('draft', 'cancelled') AND deleted_at IS NULL)::INT AS total_orders,
  (SELECT COALESCE(SUM(total), 0) FROM orders WHERE status IN ('paid', 'shipped', 'completed') AND deleted_at IS NULL)::NUMERIC(12,2) AS total_revenue,
  (SELECT COUNT(*) FROM products WHERE status = 'active')::INT AS active_products,
  (SELECT COUNT(DISTINCT user_id) FROM orders WHERE user_id IS NOT NULL AND deleted_at IS NULL)::INT AS total_customers;

CREATE OR REPLACE VIEW admin_sales_by_day AS
SELECT
  date_trunc('day', created_at)::DATE AS day,
  COUNT(*)::INT AS order_count,
  SUM(total)::NUMERIC(12,2) AS revenue
FROM orders
WHERE status IN ('paid', 'shipped', 'completed') AND deleted_at IS NULL
GROUP BY 1
ORDER BY 1 DESC;

CREATE OR REPLACE VIEW admin_token_usage_by_day AS
SELECT
  date_trunc('day', created_at)::DATE AS day,
  SUM(input_tokens)::BIGINT AS input_tokens,
  SUM(output_tokens)::BIGINT AS output_tokens,
  COUNT(*)::INT AS request_count
FROM chat_token_usage
GROUP BY 1
ORDER BY 1 DESC;
