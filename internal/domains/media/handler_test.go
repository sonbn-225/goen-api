package media

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

type fakeStorage struct {
	content     []byte
	contentType string
	err         error
}

func (s fakeStorage) GetObject(_ context.Context, _, _ string) (io.ReadCloser, ObjectInfo, error) {
	if s.err != nil {
		return nil, ObjectInfo{}, s.err
	}
	return io.NopCloser(bytes.NewReader(s.content)), ObjectInfo{ContentType: s.contentType}, nil
}

type errorEnvelope struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

func TestHandlerGetMediaSuccess(t *testing.T) {
	mod := NewModule(ModuleDeps{Storage: fakeStorage{content: []byte("fake-image"), contentType: "image/png"}})

	r := chi.NewRouter()
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/media/goen/avatars/u1/file.png", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get("Content-Type"); got != "image/png" {
		t.Fatalf("expected Content-Type image/png, got %s", got)
	}
	if rr.Body.String() != "fake-image" {
		t.Fatalf("expected streamed content, got %q", rr.Body.String())
	}
}

func TestHandlerGetMediaNotFound(t *testing.T) {
	mod := NewModule(ModuleDeps{Storage: fakeStorage{err: errors.New("not found")}})

	r := chi.NewRouter()
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/media/goen/avatars/u1/missing.png", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}

	var errResp errorEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Error.Code != "not_found" {
		t.Fatalf("expected code not_found, got %s", errResp.Error.Code)
	}
}

func TestHandlerGetMediaInvalidPath(t *testing.T) {
	mod := NewModule(ModuleDeps{Storage: fakeStorage{content: []byte("ignored"), contentType: "image/png"}})

	r := chi.NewRouter()
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/media/goen/", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	var errResp errorEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Error.Code != "validation_error" {
		t.Fatalf("expected code validation_error, got %s", errResp.Error.Code)
	}
}

func TestHandlerGetMediaStorageNotConfigured(t *testing.T) {
	mod := &Module{Handler: NewHandler(nil)}

	r := chi.NewRouter()
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/media/goen/avatars/u1/file.png", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusInternalServerError, rr.Code, rr.Body.String())
	}
}
