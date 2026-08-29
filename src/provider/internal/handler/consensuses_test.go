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
	return doJSONWithHeaders(t, router, method, path, body, nil)
}

func doJSONWithHeaders(t *testing.T, router http.Handler, method string, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
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

func TestAuthTokenProtectsAPIWhenConfigured(t *testing.T) {
	t.Setenv("CONNECT_AUTH_TOKEN", "test-secret")
	router := newTestRouter(t)

	unauthorized := doJSON(t, router, http.MethodPost, "/api/consensuses", map[string]string{
		"title": "未授权写入",
	})
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized POST status = %d, body = %s", unauthorized.Code, unauthorized.Body.String())
	}
	if unauthorized.Header().Get("WWW-Authenticate") == "" {
		t.Fatalf("WWW-Authenticate header is empty")
	}

	authorized := doJSONWithHeaders(
		t,
		router,
		http.MethodPost,
		"/api/consensuses",
		map[string]string{"title": "授权写入"},
		map[string]string{"Authorization": "Bearer test-secret"},
	)
	if authorized.Code != http.StatusCreated {
		t.Fatalf("authorized POST status = %d, body = %s", authorized.Code, authorized.Body.String())
	}

	health := httptest.NewRecorder()
	router.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health status = %d, body = %s", health.Code, health.Body.String())
	}

	optionsReq := httptest.NewRequest(http.MethodOptions, "/api/consensuses", nil)
	optionsReq.Header.Set("Origin", "https://studio.connect.cloud.quanttide.com")
	optionsReq.Header.Set("Access-Control-Request-Method", http.MethodPost)
	options := httptest.NewRecorder()
	router.ServeHTTP(options, optionsReq)
	if options.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS status = %d, body = %s", options.Code, options.Body.String())
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

func TestConsensusGraphCanBeEditedWithArbitraryDAGLinks(t *testing.T) {
	router := newTestRouter(t)

	consensusIDs := make([]string, 0, 3)
	for _, title := range []string{"发现问题", "评估方案", "回归通过"} {
		rec := doJSON(t, router, http.MethodPost, "/api/consensuses", map[string]string{
			"title": title,
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("create consensus status = %d, body = %s", rec.Code, rec.Body.String())
		}

		var created struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
			t.Fatalf("decode consensus: %v", err)
		}
		consensusIDs = append(consensusIDs, created.ID)
	}

	graphResponse := doJSON(t, router, http.MethodPost, "/api/consensus-graphs", map[string]string{
		"name":        "支付超时处理",
		"description": "记录问题、决策和验证之间的演进链路。",
	})
	if graphResponse.Code != http.StatusCreated {
		t.Fatalf("create graph status = %d, body = %s", graphResponse.Code, graphResponse.Body.String())
	}

	var graphCreated struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(graphResponse.Body.Bytes(), &graphCreated); err != nil {
		t.Fatalf("decode graph: %v", err)
	}

	for _, consensusID := range consensusIDs {
		rec := doJSON(
			t,
			router,
			http.MethodPost,
			"/api/consensus-graphs/"+graphCreated.ID+"/nodes",
			map[string]string{"consensus_id": consensusID},
		)
		if rec.Code != http.StatusOK {
			t.Fatalf("add node status = %d, body = %s", rec.Code, rec.Body.String())
		}
	}

	relationIDs := make([]string, 0, 2)
	for index, relationType := range []string{"前置条件", "验证"} {
		rec := doJSON(t, router, http.MethodPost, "/api/consensus-relations", map[string]string{
			"from":          consensusIDs[index],
			"to":            consensusIDs[index+1],
			"relation_type": relationType,
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("create relation status = %d, body = %s", rec.Code, rec.Body.String())
		}

		var relation struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &relation); err != nil {
			t.Fatalf("decode relation: %v", err)
		}
		relationIDs = append(relationIDs, relation.ID)

		rec = doJSON(
			t,
			router,
			http.MethodPost,
			"/api/consensus-graphs/"+graphCreated.ID+"/edges",
			map[string]string{"relation_id": relation.ID},
		)
		if rec.Code != http.StatusOK {
			t.Fatalf("add edge status = %d, body = %s", rec.Code, rec.Body.String())
		}
	}

	graph := doJSON(t, router, http.MethodGet, "/api/consensus-graphs/"+graphCreated.ID, nil)
	if graph.Code != http.StatusOK {
		t.Fatalf("get graph status = %d, body = %s", graph.Code, graph.Body.String())
	}
	var graphValue struct {
		Nodes []map[string]any `json:"nodes"`
		Edges []map[string]any `json:"edges"`
	}
	if err := json.Unmarshal(graph.Body.Bytes(), &graphValue); err != nil {
		t.Fatalf("decode graph detail: %v", err)
	}
	if len(graphValue.Nodes) != 3 || len(graphValue.Edges) != 2 {
		t.Fatalf("graph detail nodes/edges = %d/%d", len(graphValue.Nodes), len(graphValue.Edges))
	}
	if graphValue.Nodes[0]["id"] != consensusIDs[0] || graphValue.Nodes[2]["id"] != consensusIDs[2] {
		t.Fatalf("graph nodes are not topologically ordered: %+v", graphValue.Nodes)
	}

	relations := doJSON(
		t,
		router,
		http.MethodGet,
		"/api/consensuses/"+consensusIDs[1]+"/relations",
		nil,
	)
	if relations.Code != http.StatusOK {
		t.Fatalf("list consensus relations status = %d, body = %s", relations.Code, relations.Body.String())
	}
	var relationList struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(relations.Body.Bytes(), &relationList); err != nil {
		t.Fatalf("decode relation list: %v", err)
	}
	if relationList.Total != 2 {
		t.Fatalf("relation list total = %d", relationList.Total)
	}

	cycleEdge := doJSON(t, router, http.MethodPost, "/api/consensus-graphs/"+graphCreated.ID+"/relations", map[string]string{
		"from":          consensusIDs[2],
		"to":            consensusIDs[0],
		"relation_type": "反向验证",
	})
	if cycleEdge.Code != http.StatusConflict {
		t.Fatalf("cycle edge status = %d, body = %s", cycleEdge.Code, cycleEdge.Body.String())
	}
	outgoingAfterCycle := doJSON(
		t,
		router,
		http.MethodGet,
		"/api/consensuses/"+consensusIDs[2]+"/relations?direction=outgoing",
		nil,
	)
	var outgoingAfterCycleValue struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(outgoingAfterCycle.Body.Bytes(), &outgoingAfterCycleValue); err != nil {
		t.Fatalf("decode outgoing relations after cycle rejection: %v", err)
	}
	if outgoingAfterCycleValue.Total != 0 {
		t.Fatalf("cycle rejection left outgoing relations = %d", outgoingAfterCycleValue.Total)
	}

	updated := doJSON(
		t,
		router,
		http.MethodPut,
		"/api/consensuses/"+consensusIDs[1],
		map[string]string{
			"title":       "采用方案",
			"description": "记录可编辑的共识节点。",
		},
	)
	if updated.Code != http.StatusOK {
		t.Fatalf("update consensus status = %d, body = %s", updated.Code, updated.Body.String())
	}

	confirmed := doJSON(
		t,
		router,
		http.MethodPost,
		"/api/consensuses/confirm",
		map[string]string{"consensus_id": consensusIDs[0]},
	)
	if confirmed.Code != http.StatusOK {
		t.Fatalf("confirm consensus status = %d, body = %s", confirmed.Code, confirmed.Body.String())
	}
	updateConfirmed := doJSON(
		t,
		router,
		http.MethodPut,
		"/api/consensuses/"+consensusIDs[0],
		map[string]string{
			"title":       "确认后不应编辑",
			"description": "确认状态必须保持不可变。",
		},
	)
	if updateConfirmed.Code != http.StatusConflict {
		t.Fatalf("update confirmed consensus status = %d, body = %s", updateConfirmed.Code, updateConfirmed.Body.String())
	}

	removedEdge := doJSON(
		t,
		router,
		http.MethodDelete,
		"/api/consensus-graphs/"+graphCreated.ID+"/edges/"+relationIDs[0],
		nil,
	)
	if removedEdge.Code != http.StatusOK {
		t.Fatalf("remove graph edge status = %d, body = %s", removedEdge.Code, removedEdge.Body.String())
	}

	removedNode := doJSON(
		t,
		router,
		http.MethodDelete,
		"/api/consensus-graphs/"+graphCreated.ID+"/nodes/"+consensusIDs[1],
		nil,
	)
	if removedNode.Code != http.StatusOK {
		t.Fatalf("remove graph node status = %d, body = %s", removedNode.Code, removedNode.Body.String())
	}

	graphAfterRemoval := doJSON(t, router, http.MethodGet, "/api/consensus-graphs/"+graphCreated.ID, nil)
	var graphAfterRemovalValue struct {
		Nodes []map[string]any `json:"nodes"`
		Edges []map[string]any `json:"edges"`
	}
	if err := json.Unmarshal(graphAfterRemoval.Body.Bytes(), &graphAfterRemovalValue); err != nil {
		t.Fatalf("decode graph after removal: %v", err)
	}
	if len(graphAfterRemovalValue.Nodes) != 2 || len(graphAfterRemovalValue.Edges) != 0 {
		t.Fatalf(
			"graph after removal nodes/edges = %d/%d",
			len(graphAfterRemovalValue.Nodes),
			len(graphAfterRemovalValue.Edges),
		)
	}
}

func TestConsensusGraphNodePositionPersists(t *testing.T) {
	router := newTestRouter(t)

	created := doJSON(t, router, http.MethodPost, "/api/consensuses", map[string]string{
		"title": "拖动后的共识节点",
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create consensus status = %d, body = %s", created.Code, created.Body.String())
	}
	var consensus struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &consensus); err != nil {
		t.Fatalf("decode consensus: %v", err)
	}

	graphCreated := doJSON(t, router, http.MethodPost, "/api/consensus-graphs", map[string]string{
		"name": "拖拽持久化",
	})
	if graphCreated.Code != http.StatusCreated {
		t.Fatalf("create graph status = %d, body = %s", graphCreated.Code, graphCreated.Body.String())
	}
	var graph struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(graphCreated.Body.Bytes(), &graph); err != nil {
		t.Fatalf("decode graph: %v", err)
	}

	added := doJSON(
		t,
		router,
		http.MethodPost,
		"/api/consensus-graphs/"+graph.ID+"/nodes",
		map[string]string{"consensus_id": consensus.ID},
	)
	if added.Code != http.StatusOK {
		t.Fatalf("add graph node status = %d, body = %s", added.Code, added.Body.String())
	}

	position := doJSON(
		t,
		router,
		http.MethodPut,
		"/api/consensus-graphs/"+graph.ID+"/nodes/"+consensus.ID+"/position",
		map[string]float64{"x": 318.5, "y": 204.25},
	)
	if position.Code != http.StatusOK {
		t.Fatalf("update graph node position status = %d, body = %s", position.Code, position.Body.String())
	}

	detail := doJSON(t, router, http.MethodGet, "/api/consensus-graphs/"+graph.ID, nil)
	if detail.Code != http.StatusOK {
		t.Fatalf("get graph status = %d, body = %s", detail.Code, detail.Body.String())
	}
	var graphDetail struct {
		NodePositions map[string]struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
		} `json:"node_positions"`
	}
	if err := json.Unmarshal(detail.Body.Bytes(), &graphDetail); err != nil {
		t.Fatalf("decode graph detail: %v", err)
	}
	got, ok := graphDetail.NodePositions[consensus.ID]
	if !ok || got.X != 318.5 || got.Y != 204.25 {
		t.Fatalf("persisted node position = %+v", graphDetail.NodePositions)
	}

	invalid := doJSON(
		t,
		router,
		http.MethodPut,
		"/api/consensus-graphs/"+graph.ID+"/nodes/"+consensus.ID+"/position",
		map[string]float64{"x": 100001, "y": 0},
	)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid graph node position status = %d, body = %s", invalid.Code, invalid.Body.String())
	}

	missing := doJSON(
		t,
		router,
		http.MethodPut,
		"/api/consensus-graphs/"+graph.ID+"/nodes/"+consensus.ID+"/position",
		map[string]float64{"x": 12},
	)
	if missing.Code != http.StatusBadRequest {
		t.Fatalf("missing graph node coordinate status = %d, body = %s", missing.Code, missing.Body.String())
	}

	oversizedRequest := httptest.NewRequest(
		http.MethodPut,
		"/api/consensus-graphs/"+graph.ID+"/nodes/"+consensus.ID+"/position",
		strings.NewReader(
			`{"x":1,"y":2,"padding":"`+strings.Repeat("x", 2048)+`"}`,
		),
	)
	oversizedRequest.Header.Set("Content-Type", "application/json")
	oversized := httptest.NewRecorder()
	router.ServeHTTP(oversized, oversizedRequest)
	if oversized.Code != http.StatusBadRequest {
		t.Fatalf("oversized graph node position status = %d, body = %s", oversized.Code, oversized.Body.String())
	}
}
