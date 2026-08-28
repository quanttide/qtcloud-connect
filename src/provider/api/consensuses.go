package api

import (
	"encoding/json"
	"net/http"

	"github.com/quanttide/qtcloud-connect/provider/storage"
)

// ConsensusHandler 是共识 API 处理器。
type ConsensusHandler struct {
	storage *storage.Storage
}

// NewConsensusHandler 创建新的共识处理器。
func NewConsensusHandler(s *storage.Storage) *ConsensusHandler {
	return &ConsensusHandler{storage: s}
}

// ListConsensuses 列出所有共识。
func (h *ConsensusHandler) ListConsensuses(w http.ResponseWriter, r *http.Request) {
	consensuses, err := h.storage.ListConsensuses(nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 构建响应，包含关联的消息ID
	type ConsensusResponse struct {
		ID               string   `json:"id"`
		Content          string   `json:"content"`
		Status           string   `json:"status"`
		CreatedAt        string   `json:"created_at"`
		RelatedMessageIDs []string `json:"related_message_ids"`
	}

	var result []ConsensusResponse
	for _, c := range consensuses {
		rels, err := h.storage.GetRelationsForConsensus(c.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var messageIDs []string
		for _, r := range rels {
			messageIDs = append(messageIDs, r.MessageID)
		}

		result = append(result, ConsensusResponse{
			ID:               c.ID,
			Content:          c.Content,
			Status:           c.Status,
			CreatedAt:        c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			RelatedMessageIDs: messageIDs,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ConfirmRequest 确认共识请求。
type ConfirmRequest struct {
	ConsensusID string `json:"consensus_id"`
}

// ConfirmConsensus 确认共识。
func (h *ConsensusHandler) ConfirmConsensus(w http.ResponseWriter, r *http.Request) {
	var req ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := h.storage.UpdateConsensusStatus(req.ConsensusID, "confirmed")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if c == nil {
		http.Error(w, "not found", http.StatusNotFound)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := h.storage.UpdateConsensusStatus(req.ConsensusID, "deprecated")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if c == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":     c.ID,
		"status": c.Status,
	})
}
