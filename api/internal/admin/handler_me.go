package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

// MeHandler returns the authenticated admin user's profile and permissions.
type MeHandler struct {
	rbac *RBACService
}

func NewMeHandler(rbac *RBACService) *MeHandler {
	return &MeHandler{rbac: rbac}
}

func (h *MeHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetMe)
	return r
}

func (h *MeHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	perms := middleware.GetPermissions(r.Context())

	roles, err := h.rbac.GetUserRoles(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to load roles")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"user_id":     userID,
		"permissions": perms,
		"roles":       roles,
	})
}
