package handler

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/quanttide/qtcloud-connect/provider/internal/domain"
	"github.com/quanttide/qtcloud-connect/provider/internal/store"
)

const (
	maxRelationTypeLength     = 100
	maxGraphNameLength        = 200
	maxGraphDescriptionLength = 4000
	maxGraphNodeCoordinate    = 100000
	maxGraphNodePositionBytes = 1024
)

// ConsensusGraphHandler 是共识图和共识关系 API 处理器。
type ConsensusGraphHandler struct {
	storage *store.Storage
}

// NewConsensusGraphHandler 创建新的共识图处理器。
func NewConsensusGraphHandler(s *store.Storage) *ConsensusGraphHandler {
	return &ConsensusGraphHandler{storage: s}
}

type createConsensusRelationRequest struct {
	From         string `json:"from"`
	To           string `json:"to"`
	RelationType string `json:"relation_type"`
}

type createConsensusGraphRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type graphNodeRequest struct {
	ConsensusID string `json:"consensus_id"`
}

type updateGraphNodePositionRequest struct {
	X *float64 `json:"x"`
	Y *float64 `json:"y"`
}

type graphEdgeRequest struct {
	RelationID string `json:"relation_id"`
}

type consensusRelationListResponse struct {
	Items []*domain.ConsensusRelation `json:"items"`
	Total int                         `json:"total"`
}

// CreateConsensusRelation 创建共识之间的关系。
func (h *ConsensusGraphHandler) CreateConsensusRelation(w http.ResponseWriter, r *http.Request) {
	var req createConsensusRelationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.From = strings.TrimSpace(req.From)
	req.To = strings.TrimSpace(req.To)
	req.RelationType = strings.TrimSpace(req.RelationType)
	if req.From == "" || req.To == "" || req.RelationType == "" {
		writeError(w, "from, to and relation_type are required", http.StatusBadRequest)
		return
	}
	if len(req.RelationType) > maxRelationTypeLength {
		writeError(w, "relation_type is too long", http.StatusBadRequest)
		return
	}

	relation := &domain.ConsensusRelation{
		ID:           newID(),
		From:         req.From,
		To:           req.To,
		RelationType: req.RelationType,
	}
	if err := h.storage.AddConsensusRelation(relation); err != nil {
		if errors.Is(err, store.ErrConsensusGraphEdge) {
			writeError(w, "relation endpoints must reference existing distinct consensuses", http.StatusBadRequest)
			return
		}
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, relation, http.StatusCreated)
}

// GetConsensusRelation 获取共识关系。
func (h *ConsensusGraphHandler) GetConsensusRelation(w http.ResponseWriter, r *http.Request) {
	relation, err := h.storage.GetConsensusRelation(r.PathValue("id"))
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if relation == nil {
		writeError(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, relation, http.StatusOK)
}

// DeleteConsensusRelation 删除共识关系。
func (h *ConsensusGraphHandler) DeleteConsensusRelation(w http.ResponseWriter, r *http.Request) {
	relation, err := h.storage.GetConsensusRelation(r.PathValue("id"))
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if relation == nil {
		writeError(w, "not found", http.StatusNotFound)
		return
	}
	if err := h.storage.RemoveConsensusRelation(relation.ID); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListConsensusRelations 列出某个共识相关的关系。
func (h *ConsensusGraphHandler) ListConsensusRelations(w http.ResponseWriter, r *http.Request) {
	direction := r.URL.Query().Get("direction")
	if direction != "" && direction != "all" && direction != "incoming" && direction != "outgoing" {
		writeError(w, "direction must be all, incoming or outgoing", http.StatusBadRequest)
		return
	}
	relations, err := h.storage.ListConsensusRelations(
		r.PathValue("id"),
		direction,
		strings.TrimSpace(r.URL.Query().Get("relation_type")),
	)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, consensusRelationListResponse{Items: relations, Total: len(relations)}, http.StatusOK)
}

// CreateConsensusGraph 创建共识图。
func (h *ConsensusGraphHandler) CreateConsensusGraph(w http.ResponseWriter, r *http.Request) {
	var req createConsensusGraphRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	if req.Name == "" {
		writeError(w, "name is required", http.StatusBadRequest)
		return
	}
	if len(req.Name) > maxGraphNameLength || len(req.Description) > maxGraphDescriptionLength {
		writeError(w, "name or description is too long", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	graph := &domain.ConsensusGraph{
		ID:            newID(),
		Name:          req.Name,
		Description:   req.Description,
		CreatedAt:     now,
		UpdatedAt:     now,
		Nodes:         []*domain.Consensus{},
		Edges:         []*domain.ConsensusRelation{},
		NodePositions: map[string]domain.ConsensusGraphNodePosition{},
	}
	if err := h.storage.AddConsensusGraph(graph); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, graph, http.StatusCreated)
}

// ListConsensusGraphs 列出共识图。
func (h *ConsensusGraphHandler) ListConsensusGraphs(w http.ResponseWriter, r *http.Request) {
	graphs, err := h.storage.ListConsensusGraphs()
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"items": graphs, "total": len(graphs)}, http.StatusOK)
}

// GetConsensusGraph 获取共识图详情。
func (h *ConsensusGraphHandler) GetConsensusGraph(w http.ResponseWriter, r *http.Request) {
	graph, err := h.storage.GetConsensusGraph(r.PathValue("id"))
	if err != nil {
		if errors.Is(err, store.ErrConsensusGraphCycle) {
			writeError(w, err.Error(), http.StatusConflict)
			return
		}
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if graph == nil {
		writeError(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, graph, http.StatusOK)
}

// UpdateConsensusGraph 更新共识图元数据。
func (h *ConsensusGraphHandler) UpdateConsensusGraph(w http.ResponseWriter, r *http.Request) {
	var req createConsensusGraphRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	if req.Name == "" {
		writeError(w, "name is required", http.StatusBadRequest)
		return
	}
	if len(req.Name) > maxGraphNameLength || len(req.Description) > maxGraphDescriptionLength {
		writeError(w, "name or description is too long", http.StatusBadRequest)
		return
	}

	graph, err := h.storage.UpdateConsensusGraph(r.PathValue("id"), req.Name, req.Description)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if graph == nil {
		writeError(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, graph, http.StatusOK)
}

// AddConsensusGraphNode 将共识加入图。
func (h *ConsensusGraphHandler) AddConsensusGraphNode(w http.ResponseWriter, r *http.Request) {
	var req graphNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.ConsensusID = strings.TrimSpace(req.ConsensusID)
	if req.ConsensusID == "" {
		writeError(w, "consensus_id is required", http.StatusBadRequest)
		return
	}

	graph, err := h.storage.AddConsensusGraphNode(r.PathValue("id"), req.ConsensusID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if graph == nil {
		if existing, getErr := h.storage.GetConsensus(req.ConsensusID); getErr != nil {
			writeError(w, getErr.Error(), http.StatusInternalServerError)
			return
		} else if existing == nil {
			writeError(w, "consensus not found", http.StatusNotFound)
			return
		}
		writeError(w, "graph not found", http.StatusNotFound)
		return
	}
	writeJSON(w, graph, http.StatusOK)
}

// RemoveConsensusGraphNode 从图中移除共识。
func (h *ConsensusGraphHandler) RemoveConsensusGraphNode(w http.ResponseWriter, r *http.Request) {
	graph, err := h.storage.RemoveConsensusGraphNode(r.PathValue("id"), r.PathValue("consensus_id"))
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if graph == nil {
		writeError(w, "graph not found", http.StatusNotFound)
		return
	}
	writeJSON(w, graph, http.StatusOK)
}

// UpdateConsensusGraphNodePosition 保存图中节点的画布位置。
func (h *ConsensusGraphHandler) UpdateConsensusGraphNodePosition(w http.ResponseWriter, r *http.Request) {
	var req updateGraphNodePositionRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxGraphNodePositionBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.X == nil ||
		req.Y == nil ||
		!isValidGraphNodeCoordinate(*req.X) ||
		!isValidGraphNodeCoordinate(*req.Y) {
		writeError(w, "x and y must be finite coordinates within the canvas limit", http.StatusBadRequest)
		return
	}

	graph, err := h.storage.UpdateConsensusGraphNodePosition(
		r.PathValue("id"),
		r.PathValue("consensus_id"),
		domain.ConsensusGraphNodePosition{X: *req.X, Y: *req.Y},
	)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if graph == nil {
		writeError(w, "graph node not found", http.StatusNotFound)
		return
	}
	writeJSON(w, graph, http.StatusOK)
}

// AddConsensusGraphEdge 将共识关系加入图。
func (h *ConsensusGraphHandler) AddConsensusGraphEdge(w http.ResponseWriter, r *http.Request) {
	var req graphEdgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.RelationID = strings.TrimSpace(req.RelationID)
	if req.RelationID == "" {
		writeError(w, "relation_id is required", http.StatusBadRequest)
		return
	}

	graph, err := h.storage.AddConsensusGraphEdge(r.PathValue("id"), req.RelationID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrConsensusGraphCycle):
			writeError(w, "edge would create a cycle", http.StatusConflict)
		case errors.Is(err, store.ErrConsensusGraphEdge):
			writeError(w, "relation endpoints must both be graph nodes", http.StatusBadRequest)
		default:
			writeError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if graph == nil {
		relation, relationErr := h.storage.GetConsensusRelation(req.RelationID)
		if relationErr != nil {
			writeError(w, relationErr.Error(), http.StatusInternalServerError)
			return
		}
		if relation == nil {
			writeError(w, "relation not found", http.StatusNotFound)
			return
		}
		writeError(w, "graph not found", http.StatusNotFound)
		return
	}
	writeJSON(w, graph, http.StatusOK)
}

// CreateConsensusGraphRelation 原子创建关系并加入图。
func (h *ConsensusGraphHandler) CreateConsensusGraphRelation(w http.ResponseWriter, r *http.Request) {
	var req createConsensusRelationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.From = strings.TrimSpace(req.From)
	req.To = strings.TrimSpace(req.To)
	req.RelationType = strings.TrimSpace(req.RelationType)
	if req.From == "" || req.To == "" || req.RelationType == "" {
		writeError(w, "from, to and relation_type are required", http.StatusBadRequest)
		return
	}
	if len(req.RelationType) > maxRelationTypeLength {
		writeError(w, "relation_type is too long", http.StatusBadRequest)
		return
	}

	relation := &domain.ConsensusRelation{
		ID:           newID(),
		From:         req.From,
		To:           req.To,
		RelationType: req.RelationType,
	}
	graph, err := h.storage.AddConsensusGraphRelation(r.PathValue("id"), relation)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrConsensusGraphCycle):
			writeError(w, "edge would create a cycle", http.StatusConflict)
		case errors.Is(err, store.ErrConsensusGraphEdge):
			writeError(w, "relation endpoints must both be graph nodes", http.StatusBadRequest)
		default:
			writeError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if graph == nil {
		writeError(w, "graph not found", http.StatusNotFound)
		return
	}
	writeJSON(w, graph, http.StatusCreated)
}

// RemoveConsensusGraphEdge 从图中移除关系边。
func (h *ConsensusGraphHandler) RemoveConsensusGraphEdge(w http.ResponseWriter, r *http.Request) {
	graph, err := h.storage.RemoveConsensusGraphEdge(r.PathValue("id"), r.PathValue("relation_id"))
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if graph == nil {
		writeError(w, "graph not found", http.StatusNotFound)
		return
	}
	writeJSON(w, graph, http.StatusOK)
}

func isValidGraphNodeCoordinate(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		math.Abs(value) <= maxGraphNodeCoordinate
}
