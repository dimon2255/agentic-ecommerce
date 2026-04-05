package admin

import (
	"context"
	"sync"
	"time"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

// RBACService resolves user permissions with an in-memory TTL cache.
type RBACService struct {
	db    *supabase.Client
	mu    sync.RWMutex
	cache map[string]cachedEntry
	ttl   time.Duration
}

type cachedEntry struct {
	permissions []string
	expiresAt   time.Time
}

// NewRBACService creates a service that caches permission lookups for the given TTL.
func NewRBACService(db *supabase.Client, ttl time.Duration) *RBACService {
	return &RBACService{
		db:    db,
		cache: make(map[string]cachedEntry),
		ttl:   ttl,
	}
}

// GetPermissions returns the permission keys for the given user ID.
// Results are cached in-memory for the configured TTL.
func (s *RBACService) GetPermissions(ctx context.Context, userID string) ([]string, error) {
	// Check cache
	s.mu.RLock()
	entry, ok := s.cache[userID]
	s.mu.RUnlock()
	if ok && time.Now().Before(entry.expiresAt) {
		return entry.permissions, nil
	}

	// Call the get_user_permissions RPC
	var perms []string
	err := s.db.RPC("get_user_permissions", map[string]string{"p_user_id": userID}, &perms)
	if err != nil {
		return nil, err
	}
	if perms == nil {
		perms = []string{}
	}

	// Update cache
	s.mu.Lock()
	s.cache[userID] = cachedEntry{
		permissions: perms,
		expiresAt:   time.Now().Add(s.ttl),
	}
	s.mu.Unlock()

	return perms, nil
}

// GetUserRoles returns the role names for the given user ID.
func (s *RBACService) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	// Step 1: get role IDs
	type userRole struct {
		RoleID string `json:"role_id"`
	}
	var urs []userRole
	err := s.db.From("user_roles").
		Select("role_id").
		Eq("user_id", userID).
		Execute(&urs)
	if err != nil {
		return nil, err
	}
	if len(urs) == 0 {
		return []string{}, nil
	}

	// Step 2: get role names
	roleIDs := make([]string, len(urs))
	for i, ur := range urs {
		roleIDs[i] = ur.RoleID
	}
	type role struct {
		Name string `json:"name"`
	}
	var roles []role
	err = s.db.From("roles").
		Select("name").
		In("id", roleIDs).
		Execute(&roles)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(roles))
	for i, r := range roles {
		names[i] = r.Name
	}
	return names, nil
}

// InvalidateCache removes a user's cached permissions (e.g., after role change).
func (s *RBACService) InvalidateCache(userID string) {
	s.mu.Lock()
	delete(s.cache, userID)
	s.mu.Unlock()
}
