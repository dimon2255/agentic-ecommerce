# ADR-0005: Guest Carts Use X-Session-ID, Auth Carts Use JWT User ID

## Status

Accepted

## Context

E-commerce sites must support both guest browsing (no account) and authenticated shopping. Cart state needs to persist across page loads for guests and be tied to a user account for logged-in customers. When a guest logs in, their guest cart should merge with any existing authenticated cart.

## Options Considered

### Option A — Require login before adding to cart
- Pro: Simple — one identity model
- Con: Massive friction; most e-commerce users expect to browse and add-to-cart before creating an account

### Option B — Cookie/localStorage-only guest cart (frontend state)
- Pro: No server-side guest cart management
- Con: Cart lost on device switch; no server-side validation of prices/inventory

### Option C — Dual identity: X-Session-ID header for guests, JWT user ID for authenticated
- Pro: Guests get server-persisted carts with price validation
- Pro: Authenticated users get persistent cross-device carts
- Pro: Merge endpoint consolidates guest cart into user cart at login
- Con: Two code paths for cart identity; merge logic adds complexity

## Decision

Use **Option C**. Guest carts are keyed by a client-generated `X-Session-ID` header (UUID stored in localStorage). Authenticated carts are keyed by the JWT user ID. On login, `POST /cart/merge` combines the guest cart into the user's cart, preferring the higher quantity for duplicate SKUs.

## Consequences

- **Positive:** Zero-friction guest shopping experience
- **Positive:** Cart survives across sessions for authenticated users
- **Positive:** Merge at login prevents lost cart items
- **Negative:** API middleware must handle both identity types (optional auth middleware extracts user ID or falls back to session ID)
- **Risks:** Session ID collisions are theoretically possible but negligible with UUIDs; stale guest carts need periodic cleanup
