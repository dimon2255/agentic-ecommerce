-- Phase 3, PR 3C: Full-text search on products

-- Generated tsvector column for full-text search
ALTER TABLE products ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
        to_tsvector('english', coalesce(name, '') || ' ' || coalesce(description, ''))
    ) STORED;

-- GIN index for fast full-text search
CREATE INDEX idx_products_search ON products USING GIN(search_vector);
