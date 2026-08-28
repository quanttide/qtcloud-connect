package handler

import (
	"net/http"

	"github.com/quanttide/qtcloud-connect/provider/internal/store"
)

// NewRouter 创建新的路由器。
func NewRouter(s *store.Storage) *http.ServeMux {
	mux := http.NewServeMux()

	// 创建处理器
	messageHandler := NewMessageHandler(s)
	consensusHandler := NewConsensusHandler(s)

	// 消息 API
	mux.HandleFunc("GET /api/messages", messageHandler.ListMessages)
	mux.HandleFunc("GET /api/messages/{id}", messageHandler.GetMessage)

	// 共识 API
	mux.HandleFunc("GET /api/consensuses", consensusHandler.ListConsensuses)
	mux.HandleFunc("POST /api/consensuses/confirm", consensusHandler.ConfirmConsensus)
	mux.HandleFunc("POST /api/consensuses/deprecate", consensusHandler.DeprecateConsensus)

	// 健康检查
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	return mux
}
