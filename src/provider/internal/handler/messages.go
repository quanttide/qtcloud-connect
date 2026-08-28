package handler

import (
	"encoding/json"
	"net/http"

	"github.com/quanttide/qtcloud-connect/provider/internal/store"
)

// MessageHandler 是消息 API 处理器。
type MessageHandler struct {
	storage *store.Storage
}

// NewMessageHandler 创建新的消息处理器。
func NewMessageHandler(s *store.Storage) *MessageHandler {
	return &MessageHandler{storage: s}
}

// ListMessages 列出所有消息。
func (h *MessageHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	messages, err := h.storage.ListMessages()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// GetMessage 获取消息详情。
func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	msg, err := h.storage.GetMessage(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if msg == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}
