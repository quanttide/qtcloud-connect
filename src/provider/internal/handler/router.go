package handler

import (
	"crypto/subtle"
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
	graphHandler := NewConsensusGraphHandler(s)

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
	mux.HandleFunc("GET /api/consensuses/{id}/relations", graphHandler.ListConsensusRelations)

	// 共识关系 API
	mux.HandleFunc("POST /api/consensus-relations", graphHandler.CreateConsensusRelation)
	mux.HandleFunc("GET /api/consensus-relations/{id}", graphHandler.GetConsensusRelation)
	mux.HandleFunc("DELETE /api/consensus-relations/{id}", graphHandler.DeleteConsensusRelation)

	// 共识图 API
	mux.HandleFunc("POST /api/consensus-graphs", graphHandler.CreateConsensusGraph)
	mux.HandleFunc("GET /api/consensus-graphs", graphHandler.ListConsensusGraphs)
	mux.HandleFunc("GET /api/consensus-graphs/{id}", graphHandler.GetConsensusGraph)
	mux.HandleFunc("PUT /api/consensus-graphs/{id}", graphHandler.UpdateConsensusGraph)
	mux.HandleFunc("POST /api/consensus-graphs/{id}/nodes", graphHandler.AddConsensusGraphNode)
	mux.HandleFunc("DELETE /api/consensus-graphs/{id}/nodes/{consensus_id}", graphHandler.RemoveConsensusGraphNode)
	mux.HandleFunc("POST /api/consensus-graphs/{id}/relations", graphHandler.CreateConsensusGraphRelation)
	mux.HandleFunc("POST /api/consensus-graphs/{id}/edges", graphHandler.AddConsensusGraphEdge)
	mux.HandleFunc("DELETE /api/consensus-graphs/{id}/edges/{relation_id}", graphHandler.RemoveConsensusGraphEdge)

	// 健康检查
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	return withCORS(withAuth(mux))
}

func withAuth(next http.Handler) http.Handler {
	token := strings.TrimSpace(os.Getenv("CONNECT_AUTH_TOKEN"))
	if token == "" {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions || r.URL.Path == "/healthz" ||
			!strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		if subtle.ConstantTimeCompare([]byte(bearerToken(r.Header.Get("Authorization"))), []byte(token)) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="qtcloud-connect"`)
			writeError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func bearerToken(header string) string {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigin := isAllowedOrigin(origin)
		if allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
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
