package middleware

import (
	"context"
	"net/http"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

// PermissionResolver loads permission keys for a given user ID.
// Implemented by admin.RBACService — defined here as an interface to avoid import cycles.
type PermissionResolver interface {
	GetPermissions(ctx context.Context, userID string) ([]string, error)
}

// permissionsCtxKey is the context key for resolved user permissions.
const PermissionsKey contextKey = "user_permissions"

// GetPermissions extracts the resolved permission slice from request context.
func GetPermissions(ctx context.Context) []string {
	perms, _ := ctx.Value(PermissionsKey).([]string)
	return perms
}

// RequirePermission returns middleware that checks the authenticated user has ALL of the
// specified permission keys. It stores the resolved permissions in context so downstream
// handlers can read them via middleware.GetPermissions.
// Must be chained after RequireAuth (relies on UserIDKey in context).
func RequirePermission(resolver PermissionResolver, required ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, perms, ok := resolvePermissions(w, r, resolver)
			if !ok {
				return
			}
			r = r.WithContext(ctx)

			for _, key := range required {
				if !containsStr(perms, key) {
					response.Error(w, http.StatusForbidden, "insufficient permissions")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission returns middleware that checks the authenticated user has at least
// ONE of the specified permission keys. Used as a gate on the /admin route group to ensure
// the user has some admin access before proceeding to sub-routes.
func RequireAnyPermission(resolver PermissionResolver, required ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, perms, ok := resolvePermissions(w, r, resolver)
			if !ok {
				return
			}
			r = r.WithContext(ctx)

			for _, key := range required {
				if containsStr(perms, key) {
					next.ServeHTTP(w, r)
					return
				}
			}
			response.Error(w, http.StatusForbidden, "insufficient permissions")
		})
	}
}

// resolvePermissions loads and caches permissions in context. If permissions are already
// present in context (from an outer middleware), they are reused.
func resolvePermissions(w http.ResponseWriter, r *http.Request, resolver PermissionResolver) (context.Context, []string, bool) {
	// Reuse if already resolved by outer middleware
	if perms := GetPermissions(r.Context()); perms != nil {
		return r.Context(), perms, true
	}

	userID, ok := GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "authentication required")
		return nil, nil, false
	}

	perms, err := resolver.GetPermissions(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to check permissions")
		return nil, nil, false
	}

	ctx := context.WithValue(r.Context(), PermissionsKey, perms)
	return ctx, perms, true
}

func containsStr(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
