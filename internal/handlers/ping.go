package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type PingResponse struct {
	Service   string    `json:"service"`
	Env       string    `json:"env"`
	Time      time.Time `json:"time"`
	RequestID string    `json:"request_id"`
}

// Ping godoc
// @Summary Ping
// @Description Lightweight endpoint for browser/app connectivity test.
// @Tags meta
// @Produce json
// @Success 200 {object} PingResponse
// @Router /ping [get]
func Ping(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, PingResponse{
			Service:   "goen-api",
			Env:       d.Cfg.Env,
			Time:      time.Now().UTC(),
			RequestID: middleware.GetReqID(r.Context()),
		})
	}
}
 
