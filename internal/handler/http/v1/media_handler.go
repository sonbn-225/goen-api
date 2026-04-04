package v1

import (
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
	"github.com/sonbn-225/goen-api/internal/pkg/storage"
)

type MediaHandler struct {
	s3 *storage.S3Client
}

func NewMediaHandler(s3 *storage.S3Client) *MediaHandler {
	return &MediaHandler{s3: s3}
}

func (h *MediaHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Get("/media/{bucket}/*", h.GetMedia)
}

// GetMedia godoc
// @Summary Proxy Media Images
// @Description Retrieve images via minio reverse proxy
// @Tags Public
// @Produce image/*
// @Param bucket path string true "Bucket Name"
// @Param key path string true "Object Key"
// @Success 200 {file} file
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 503 {object} response.ErrorEnvelope
// @Router /media/{bucket}/{key} [get]
func (h *MediaHandler) GetMedia(w http.ResponseWriter, r *http.Request) {
	if h.s3 == nil {
		response.WriteError(w, http.StatusServiceUnavailable, "unavailable", "storage not configured", nil)
		return
	}

	bucket := chi.URLParam(r, "bucket")
	key := strings.TrimPrefix(chi.URLParam(r, "*"), "/")
	if bucket == "" || key == "" {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "missing bucket or key", nil)
		return
	}

	obj, info, err := h.s3.GetObject(r.Context(), bucket, key)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "media not found", nil)
		return
	}
	defer obj.Close()

	w.Header().Set("Content-Type", info.ContentType)
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, obj)
}
