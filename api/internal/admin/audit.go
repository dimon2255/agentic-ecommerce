package admin

import (
	"context"
	"net/http"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

// AuditEntry represents a single admin action to be logged.
type AuditEntry struct {
	UserID       string `json:"user_id"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id,omitempty"`
	Changes      any    `json:"changes,omitempty"`
	IPAddress    string `json:"ip_address,omitempty"`
}

// AuditService records admin actions to the admin_audit_log table.
type AuditService struct {
	db *supabase.Client
}

// NewAuditService creates a new audit logger backed by Supabase.
func NewAuditService(db *supabase.Client) *AuditService {
	return &AuditService{db: db}
}

// Log records an audit entry. It is fire-and-forget; errors are returned
// but callers may choose to log and continue rather than fail the request.
func (s *AuditService) Log(ctx context.Context, entry AuditEntry) error {
	return s.db.From("admin_audit_log").Insert(entry).Execute(nil)
}

// LogFromRequest is a convenience that fills IPAddress from the request.
// The caller must provide userID (extracted from context by the handler/middleware).
func (s *AuditService) LogFromRequest(r *http.Request, userID, action, resourceType, resourceID string, changes any) error {
	return s.Log(r.Context(), AuditEntry{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Changes:      changes,
		IPAddress:    realIP(r),
	})
}

// realIP extracts the client IP, preferring X-Forwarded-For if present.
func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}
