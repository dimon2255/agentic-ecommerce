-- Phase 1: AI Shopping Assistant — pgvector + chat tables

-- Enable pgvector extension for semantic search
CREATE EXTENSION IF NOT EXISTS vector;

-- ============================================================
-- Product embeddings for semantic search
-- ============================================================
CREATE TABLE product_embeddings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  content TEXT NOT NULL,
  embedding VECTOR(1024) NOT NULL,
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_product_embeddings_product ON product_embeddings(product_id);
CREATE INDEX idx_product_embeddings_vector ON product_embeddings
  USING hnsw (embedding vector_cosine_ops);

-- ============================================================
-- Chat sessions
-- ============================================================
CREATE TABLE chat_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  title TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_chat_sessions_user ON chat_sessions(user_id);

-- ============================================================
-- Chat messages
-- ============================================================
CREATE TABLE chat_messages (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id UUID NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
  role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
  content TEXT NOT NULL,
  product_ids UUID[] DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_chat_messages_session ON chat_messages(session_id, created_at);

-- ============================================================
-- Row Level Security
-- ============================================================
ALTER TABLE product_embeddings ENABLE ROW LEVEL SECURITY;
ALTER TABLE chat_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE chat_messages ENABLE ROW LEVEL SECURITY;

-- Product embeddings: public read for search, service role full access
CREATE POLICY "product_embeddings_public_read" ON product_embeddings
  FOR SELECT USING (true);
CREATE POLICY "product_embeddings_service_all" ON product_embeddings
  FOR ALL USING (auth.role() = 'service_role');

-- Chat sessions: users see only their own, service role full access
CREATE POLICY "chat_sessions_user_read" ON chat_sessions
  FOR SELECT USING (auth.uid() = user_id);
CREATE POLICY "chat_sessions_service_all" ON chat_sessions
  FOR ALL USING (auth.role() = 'service_role');

-- Chat messages: users see messages in their sessions, service role full access
CREATE POLICY "chat_messages_user_read" ON chat_messages
  FOR SELECT USING (
    EXISTS (
      SELECT 1 FROM chat_sessions
      WHERE chat_sessions.id = chat_messages.session_id
      AND chat_sessions.user_id = auth.uid()
    )
  );
CREATE POLICY "chat_messages_service_all" ON chat_messages
  FOR ALL USING (auth.role() = 'service_role');

-- ============================================================
-- Triggers (reuse update_updated_at from 00001)
-- ============================================================
CREATE TRIGGER product_embeddings_updated_at
  BEFORE UPDATE ON product_embeddings
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER chat_sessions_updated_at
  BEFORE UPDATE ON chat_sessions
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- Vector similarity search RPC
-- ============================================================
CREATE OR REPLACE FUNCTION match_products(
  query_embedding VECTOR(1024),
  match_threshold FLOAT DEFAULT 0.3,
  match_count INT DEFAULT 5
)
RETURNS TABLE (
  id UUID,
  product_id UUID,
  content TEXT,
  metadata JSONB,
  similarity FLOAT
)
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
  RETURN QUERY
  SELECT
    pe.id,
    pe.product_id,
    pe.content,
    pe.metadata,
    1 - (pe.embedding <=> query_embedding) AS similarity
  FROM product_embeddings pe
  WHERE 1 - (pe.embedding <=> query_embedding) > match_threshold
  ORDER BY pe.embedding <=> query_embedding
  LIMIT match_count;
END;
$$;
