# ADR-0001: Go API Bypasses RLS via Service Role Key

## Status

Accepted

## Context

Supabase PostgreSQL uses Row Level Security (RLS) to control data access. The frontend uses an anon key with RLS policies for public read access. The Go API needs to perform mutations (cart writes, order creation, admin operations) that span multiple tables and require elevated permissions.

We needed a pattern that allows the API to perform trusted server-side operations without being blocked by per-row policies designed for direct client access.

## Options Considered

### Option A — Anon key with permissive RLS policies
- Pro: Single key, simpler config
- Con: Requires complex RLS policies for every API operation; policies become a maintenance burden and security risk

### Option B — Service role key in Go API, anon key in frontend
- Pro: Clean separation — frontend gets read-only public data, API has full trusted access
- Pro: RLS policies stay simple (public read for anon)
- Con: Service role key bypasses all RLS — a bug in Go API logic could expose/modify any data

### Option C — Custom PostgreSQL roles per operation
- Pro: Fine-grained database-level permissions
- Con: Significant complexity; hard to manage with Supabase hosted PostgreSQL

## Decision

Use **Option B**. The Go API authenticates to Supabase with the service role key for all mutations. The frontend anon key provides public read-only access through RLS. The API validates JWTs itself via middleware before performing any operations.

## Consequences

- **Positive:** RLS policies are simple — just public read for anon key
- **Positive:** Go API has full control over authorization logic in application code
- **Negative:** Service role key is a high-value secret — must never leak to frontend
- **Risks:** Authorization bugs in Go handlers could bypass intended access controls since RLS is not a safety net for API operations
