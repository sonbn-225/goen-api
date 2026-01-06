//go:build ignore

package handlers

import (
	"net/http"

	"github.com/sonbn-225/goen-api/internal/apierror"
)

type ErrorEnvelope = apierror.Envelope

type APIError = apierror.Error

func writeError(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	apierror.Write(w, status, code, message, details)
}
