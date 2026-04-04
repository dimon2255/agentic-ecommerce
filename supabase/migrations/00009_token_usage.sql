-- Token usage tracking for AI assistant cost management
CREATE TABLE chat_token_usage (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id UUID REFERENCES chat_sessions(id) ON DELETE SET NULL,
  input_tokens INT NOT NULL DEFAULT 0,
  output_tokens INT NOT NULL DEFAULT 0,
  model TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_token_usage_created ON chat_token_usage(created_at);
CREATE INDEX idx_token_usage_session ON chat_token_usage(session_id);

ALTER TABLE chat_token_usage ENABLE ROW LEVEL SECURITY;

CREATE POLICY "token_usage_service_all" ON chat_token_usage
  FOR ALL USING (auth.role() = 'service_role');

-- RPC: get daily token usage totals (UTC day boundary)
CREATE OR REPLACE FUNCTION get_daily_token_usage()
RETURNS TABLE(total_input_tokens BIGINT, total_output_tokens BIGINT)
LANGUAGE sql STABLE
AS $$
  SELECT
    COALESCE(SUM(input_tokens), 0)::BIGINT AS total_input_tokens,
    COALESCE(SUM(output_tokens), 0)::BIGINT AS total_output_tokens
  FROM chat_token_usage
  WHERE created_at >= date_trunc('day', now() AT TIME ZONE 'UTC');
$$;
