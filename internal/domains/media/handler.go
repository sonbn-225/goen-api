package media

import (
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// getMedia godoc
// @Summary Get media file
// @Description Stream a media object from SeaweedFS through backend proxy.
// @Tags media
// @Produce application/octet-stream
// @Security BearerAuth
// @Param bucket path string true "Bucket name"
// @Param key path string true "Object key"
// @Success 200 {file} binary
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /media/{bucket}/{key} [get]
func (h *Handler) getMedia(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		response.WriteError(w, apperrors.New(apperrors.KindInternal, "service not configured"))
		return
	}

	bucket := chi.URLParam(r, "bucket")
	key := strings.TrimPrefix(chi.URLParam(r, "*"), "/")

	obj, info, err := h.service.GetObject(r.Context(), bucket, key)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	defer obj.Close()

	contentType := strings.TrimSpace(info.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, obj)
}
