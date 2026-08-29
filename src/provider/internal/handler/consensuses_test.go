package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/quanttide/qtcloud-connect/provider/internal/store"
)

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	s, err := store.New(filepath.Join(t.TempDir(), "qtcloud-connect.db"))
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	return NewRouter(s)
}

func doJSON(t *testing.T, router http.Handler, method string, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestConsensusCreateListAndGetFollowSpec(t *testing.T) {
	router := newTestRouter(t)

	create := doJSON(t, router, http.MethodPost, "/api/consensuses", map[string]string{
		"title":       "Studio 优先展示共识追溯页面",
		"description": "v0.1 验收先打通 CLI 写入、Provider 持久化、Studio 展示。",
	})
	if create.Code != http.StatusCreated {
		t.Fatalf("POST /api/consensuses status = %d, body = %s", create.Code, create.Body.String())
	}

	var created struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.ID == "" || created.CreatedAt == "" || created.UpdatedAt == "" {
		t.Fatalf("created consensus missing identity/timestamps: %+v", created)
	}
	if len(created.ID) != 36 || created.ID[8] != '-' || created.ID[13] != '-' ||
		created.ID[18] != '-' || created.ID[23] != '-' {
		t.Fatalf("created consensus id is not UUID-shaped: %q", created.ID)
	}
	if created.Title != "Studio 优先展示共识追溯页面" {
		t.Fatalf("created title = %q", created.Title)
	}

	list := doJSON(t, router, http.MethodGet, "/api/consensuses", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("GET /api/consensuses status = %d, body = %s", list.Code, list.Body.String())
	}
	var listed struct {
		Items    []map[string]any `json:"items"`
		Total    int              `json:"total"`
		Page     int              `json:"page"`
		PageSize int              `json:"page_size"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if listed.Total != 1 || listed.Page != 1 || listed.PageSize != 20 || len(listed.Items) != 1 {
		t.Fatalf("unexpected list wrapper: %+v", listed)
	}
	if listed.Items[0]["title"] != created.Title {
		t.Fatalf("listed title = %v", listed.Items[0]["title"])
	}

	get := doJSON(t, router, http.MethodGet, "/api/consensuses/"+created.ID, nil)
	if get.Code != http.StatusOK {
		t.Fatalf("GET /api/consensuses/{id} status = %d, body = %s", get.Code, get.Body.String())
	}
	var found struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &found); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if found.ID != created.ID || found.Title != created.Title {
		t.Fatalf("unexpected detail response: %+v", found)
	}
}

func TestCreateConsensusRejectsEmptyTitle(t *testing.T) {
	router := newTestRouter(t)

	rec := doJSON(t, router, http.MethodPost, "/api/consensuses", map[string]string{
		"description": "缺少标题不能形成可读共识。",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/consensuses status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestConsensusListHandlesOversizedPage(t *testing.T) {
	router := newTestRouter(t)

	rec := doJSON(
		t,
		router,
		http.MethodGet,
		"/api/consensuses?page=9223372036854775807&page_size=2",
		nil,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/consensuses status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestConsensusListCapsOversizedPageSize(t *testing.T) {
	router := newTestRouter(t)

	create := doJSON(t, router, http.MethodPost, "/api/consensuses", map[string]string{
		"title":       "分页边界共识",
		"description": "用于覆盖极大 page_size 的列表请求。",
	})
	if create.Code != http.StatusCreated {
		t.Fatalf("POST /api/consensuses status = %d, body = %s", create.Code, create.Body.String())
	}

	rec := doJSON(
		t,
		router,
		http.MethodGet,
		"/api/consensuses?page=2&page_size=9223372036854775807",
		nil,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/consensuses status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var listed struct {
		Items    []map[string]any `json:"items"`
		PageSize int              `json:"page_size"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listed.Items) != 0 || listed.PageSize != maxConsensusPageSize {
		t.Fatalf("unexpected list response: %+v", listed)
	}
}

func TestConsensusPreflightAllowsStudioOrigin(t *testing.T) {
	t.Setenv("CONNECT_ALLOWED_ORIGINS", "")
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodOptions, "/api/consensuses", nil)
	req.Header.Set("Origin", "https://studio.connect.cloud.quanttide.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS /api/consensuses status = %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://studio.connect.cloud.quanttide.com" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, http.MethodGet) {
		t.Fatalf("Access-Control-Allow-Methods = %q", got)
	}
}

func TestConsensusPreflightRejectsDisallowedOrigin(t *testing.T) {
	t.Setenv("CONNECT_ALLOWED_ORIGINS", "https://studio.connect.cloud.quanttide.com")
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodOptions, "/api/consensuses", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("OPTIONS /api/consensuses status = %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
}
