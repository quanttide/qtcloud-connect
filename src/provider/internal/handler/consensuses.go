package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/quanttide/qtcloud-connect/provider/internal/domain"
	"github.com/quanttide/qtcloud-connect/provider/internal/store"
)

const maxConsensusPageSize = 100

// ConsensusHandler 是共识 API 处理器。
type ConsensusHandler struct {
	storage *store.Storage
}

// NewConsensusHandler 创建新的共识处理器。
func NewConsensusHandler(s *store.Storage) *ConsensusHandler {
	return &ConsensusHandler{storage: s}
}

// CreateConsensusRequest 是创建共识请求。
type CreateConsensusRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UpdateConsensusRequest 是更新共识请求。
type UpdateConsensusRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// ListConsensusesResponse 是共识列表响应。
type ListConsensusesResponse struct {
	Items    []*domain.Consensus `json:"items"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// CreateConsensus 创建共识。
func (h *ConsensusHandler) CreateConsensus(w http.ResponseWriter, r *http.Request) {
	var req CreateConsensusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	if req.Title == "" {
		writeError(w, "title is required", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	c := &domain.Consensus{
		ID:          newID(),
		Title:       req.Title,
		Description: req.Description,
		Status:      "proposed",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := h.storage.AddConsensus(c); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, c, http.StatusCreated)
}

// ListConsensuses 列出所有共识。
func (h *ConsensusHandler) ListConsensuses(w http.ResponseWriter, r *http.Request) {
	consensuses, err := h.storage.ListConsensuses(nil)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	pageSize := parsePositiveInt(r.URL.Query().Get("page_size"), 20)
	if pageSize > maxConsensusPageSize {
		pageSize = maxConsensusPageSize
	}
	total := len(consensuses)
	totalPages := (total + pageSize - 1) / pageSize
	start := total
	if page <= totalPages {
		start = (page - 1) * pageSize
	}
	if total == 0 || start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	writeJSON(w, ListConsensusesResponse{
		Items:    consensuses[start:end],
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, http.StatusOK)
}

// GetConsensus 获取共识详情。
func (h *ConsensusHandler) GetConsensus(w http.ResponseWriter, r *http.Request) {
	c, err := h.storage.GetConsensus(r.PathValue("id"))
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if c == nil {
		writeError(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, c, http.StatusOK)
}

// UpdateConsensus 更新共识标题和描述。
func (h *ConsensusHandler) UpdateConsensus(w http.ResponseWriter, r *http.Request) {
	var req UpdateConsensusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	if req.Title == "" {
		writeError(w, "title is required", http.StatusBadRequest)
		return
	}

	c, err := h.storage.UpdateConsensus(r.PathValue("id"), req.Title, req.Description)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if c == nil {
		writeError(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, c, http.StatusOK)
}

// ConfirmRequest 确认共识请求。
type ConfirmRequest struct {
	ConsensusID string `json:"consensus_id"`
}

// ConfirmConsensus 确认共识。
func (h *ConsensusHandler) ConfirmConsensus(w http.ResponseWriter, r *http.Request) {
	var req ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	c, err := h.storage.UpdateConsensusStatus(req.ConsensusID, "confirmed")
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if c == nil {
		writeError(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":     c.ID,
		"status": c.Status,
	})
}

// DeprecateRequest 废弃共识请求。
type DeprecateRequest struct {
	ConsensusID string `json:"consensus_id"`
}

// DeprecateConsensus 废弃共识。
func (h *ConsensusHandler) DeprecateConsensus(w http.ResponseWriter, r *http.Request) {
	var req DeprecateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	c, err := h.storage.UpdateConsensusStatus(req.ConsensusID, "deprecated")
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if c == nil {
		writeError(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":     c.ID,
		"status": c.Status,
	})
}

func writeJSON(w http.ResponseWriter, value any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, message string, status int) {
	writeJSON(w, map[string]string{"error": message}, status)
}

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func newID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		hex.EncodeToString(buf[0:4]),
		hex.EncodeToString(buf[4:6]),
		hex.EncodeToString(buf[6:8]),
		hex.EncodeToString(buf[8:10]),
		hex.EncodeToString(buf[10:16]),
	)
}
