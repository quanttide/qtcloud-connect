package handler

import (
	"net/http"
	"os"
	"strings"

	"github.com/quanttide/qtcloud-connect/provider/internal/store"
)

// NewRouter 创建新的路由器。
func NewRouter(s *store.Storage) http.Handler {
	mux := http.NewServeMux()

	// 创建处理器
	messageHandler := NewMessageHandler(s)
	consensusHandler := NewConsensusHandler(s)

	// 消息 API
	mux.HandleFunc("GET /api/messages", messageHandler.ListMessages)
	mux.HandleFunc("GET /api/messages/{id}", messageHandler.GetMessage)

	// 共识 API
	mux.HandleFunc("POST /api/consensuses", consensusHandler.CreateConsensus)
	mux.HandleFunc("GET /api/consensuses", consensusHandler.ListConsensuses)
	mux.HandleFunc("GET /api/consensuses/{id}", consensusHandler.GetConsensus)
	mux.HandleFunc("PUT /api/consensuses/{id}", consensusHandler.UpdateConsensus)
	mux.HandleFunc("POST /api/consensuses/confirm", consensusHandler.ConfirmConsensus)
	mux.HandleFunc("POST /api/consensuses/deprecate", consensusHandler.DeprecateConsensus)

	// 健康检查
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigin := isAllowedOrigin(origin)
		if allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		}

		if r.Method == http.MethodOptions && origin != "" {
			if !allowedOrigin {
				writeError(w, "origin is not allowed", http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isAllowedOrigin(origin string) bool {
	if origin == "" {
		return false
	}

	configured := os.Getenv("CONNECT_ALLOWED_ORIGINS")
	if configured != "" {
		for _, allowed := range strings.Split(configured, ",") {
			if strings.TrimSpace(allowed) == origin {
				return true
			}
		}
		return false
	}

	return origin == "https://studio.connect.cloud.quanttide.com" ||
		strings.HasPrefix(origin, "http://localhost:") ||
		strings.HasPrefix(origin, "http://127.0.0.1:")
}
