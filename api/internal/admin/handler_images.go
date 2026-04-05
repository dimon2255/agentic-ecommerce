package admin

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

// ImageHandler generates presigned upload URLs for Supabase Storage.
type ImageHandler struct {
	storage *supabase.StorageClient
}

func NewImageHandler(storage *supabase.StorageClient) *ImageHandler {
	return &ImageHandler{storage: storage}
}

func (h *ImageHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/upload-url", h.GenerateUploadURL)
	return r
}

func (h *ImageHandler) GenerateUploadURL(w http.ResponseWriter, r *http.Request) {
	var req UploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	uploadURL, publicURL, err := h.storage.CreateSignedUploadURL("product-images", req.Filename, req.ContentType)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate upload URL")
		return
	}

	response.JSON(w, http.StatusOK, UploadURLResponse{
		UploadURL: uploadURL,
		PublicURL: publicURL,
	})
}
