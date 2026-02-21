package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKBHandler_ManagementEndpointsRemoved(t *testing.T) {
	h := NewKBHandler(nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/kb", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}
