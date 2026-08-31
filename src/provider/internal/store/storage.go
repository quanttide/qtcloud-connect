package store

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/quanttide/qtcloud-connect/provider/internal/domain"
)

var (
	ErrConsensusGraphCycle = errors.New("consensus graph edge would create a cycle")
	ErrConsensusGraphEdge  = errors.New("consensus graph edge is invalid")
	ErrConsensusImmutable  = errors.New("marked consensus is immutable")
	ErrMessageImmutable    = errors.New("message content is immutable")
)

// Storage 是 SQLite 存储层。
type Storage struct {
	db *sql.DB
}

// New 创建新的存储实例。
func New(dbPath string) (*Storage, error) {
	if err := ensureParentDir(dbPath); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	s := &Storage{db: db}
	if err := s.initTables(); err != nil {
		return nil, err
	}

	return s, nil
}

func ensureParentDir(dbPath string) error {
	if dbPath == "" || dbPath == ":memory:" {
		return nil
	}

	dir := filepath.Dir(dbPath)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

// initTables 初始化数据库表。
func (s *Storage) initTables() error {
	query := `
		CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY, content TEXT NOT NULL, type TEXT NOT NULL,
			created_at TEXT NOT NULL, updated_at TEXT
		);
		CREATE TABLE IF NOT EXISTS consensuses (
			id TEXT PRIMARY KEY, title TEXT NOT NULL, description TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'proposed',
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS relations (
			id TEXT PRIMARY KEY, message_id TEXT NOT NULL, consensus_id TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS consensus_relations (
			id TEXT PRIMARY KEY, from_id TEXT NOT NULL, to_id TEXT NOT NULL,
			relation_type TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS consensus_graphs (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, description TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS consensus_graph_nodes (
			graph_id TEXT NOT NULL, consensus_id TEXT NOT NULL,
			position_x REAL, position_y REAL,
			PRIMARY KEY (graph_id, consensus_id)
		);
		CREATE TABLE IF NOT EXISTS consensus_graph_edges (
			graph_id TEXT NOT NULL, relation_id TEXT NOT NULL,
			PRIMARY KEY (graph_id, relation_id)
		);
		CREATE INDEX IF NOT EXISTS idx_consensus_relations_from ON consensus_relations (from_id);
		CREATE INDEX IF NOT EXISTS idx_consensus_relations_to ON consensus_relations (to_id);
		CREATE INDEX IF NOT EXISTS idx_consensus_graph_edges_graph ON consensus_graph_edges (graph_id);
	`
	if _, err := s.db.Exec(query); err != nil {
		return err
	}
	if err := s.ensureColumn("consensuses", "title", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := s.ensureColumn("consensuses", "description", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := s.ensureColumn("consensuses", "updated_at", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := s.ensureColumn("consensus_graph_nodes", "position_x", "REAL"); err != nil {
		return err
	}
	if err := s.ensureColumn("consensus_graph_nodes", "position_y", "REAL"); err != nil {
		return err
	}
	return s.migrateConsensusRows()
}

func (s *Storage) ensureColumn(table, column, definition string) error {
	exists, err := s.hasColumn(table, column)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = s.db.Exec("ALTER TABLE " + table + " ADD COLUMN " + column + " " + definition)
	return err
}

func (s *Storage) hasColumn(table, column string) (bool, error) {
	rows, err := s.db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var typ string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	return false, nil
}

func (s *Storage) migrateConsensusRows() error {
	hasContent, err := s.hasColumn("consensuses", "content")
	if err != nil {
		return err
	}
	if hasContent {
		if _, err := s.db.Exec("UPDATE consensuses SET title = content WHERE title = '' AND content IS NOT NULL"); err != nil {
			return err
		}
	}
	_, err = s.db.Exec("UPDATE consensuses SET updated_at = created_at WHERE updated_at IS NULL OR updated_at = ''")
	return err
}

// Close 关闭数据库连接。
func (s *Storage) Close() error {
	return s.db.Close()
}

// AddMessage 添加消息。
func (s *Storage) AddMessage(msg *domain.Message) error {
	_, err := s.db.Exec(
		"INSERT INTO messages (id, content, type, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		msg.ID, msg.Content, msg.Type, msg.CreatedAt.Format(time.RFC3339), formatTime(msg.UpdatedAt),
	)
	return err
}

// GetMessage 获取消息。
func (s *Storage) GetMessage(id string) (*domain.Message, error) {
	row := s.db.QueryRow("SELECT * FROM messages WHERE id = ?", id)
	return scanMessage(row)
}

// ListMessages 列出所有消息。
func (s *Storage) ListMessages() ([]*domain.Message, error) {
	rows, err := s.db.Query("SELECT * FROM messages ORDER BY created_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		msg, err := scanMessageRows(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// UpdateMessage 拒绝更新消息内容。
func (s *Storage) UpdateMessage(id, content string) (*domain.Message, error) {
	return nil, ErrMessageImmutable
}

// AddConsensus 添加共识。
func (s *Storage) AddConsensus(c *domain.Consensus) error {
	_, err := s.db.Exec(
		"INSERT INTO consensuses (id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		c.ID, c.Title, c.Description, c.Status, c.CreatedAt.Format(time.RFC3339), c.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetConsensus 获取共识。
func (s *Storage) GetConsensus(id string) (*domain.Consensus, error) {
	row := s.db.QueryRow("SELECT id, title, description, status, created_at, updated_at FROM consensuses WHERE id = ?", id)
	return scanConsensus(row)
}

// ListConsensuses 列出所有共识。
func (s *Storage) ListConsensuses(status *string) ([]*domain.Consensus, error) {
	var rows *sql.Rows
	var err error

	if status != nil {
		rows, err = s.db.Query("SELECT id, title, description, status, created_at, updated_at FROM consensuses WHERE status = ? ORDER BY created_at", *status)
	} else {
		rows, err = s.db.Query("SELECT id, title, description, status, created_at, updated_at FROM consensuses ORDER BY created_at")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var consensuses []*domain.Consensus
	for rows.Next() {
		c, err := scanConsensusRows(rows)
		if err != nil {
			return nil, err
		}
		consensuses = append(consensuses, c)
	}
	return consensuses, nil
}

// UpdateConsensusStatus 更新共识状态。
func (s *Storage) UpdateConsensusStatus(id, status string) (*domain.Consensus, error) {
	now := time.Now()
	result, err := s.db.Exec(
		"UPDATE consensuses SET status = ?, updated_at = ? WHERE id = ?",
		status, now.Format(time.RFC3339), id,
	)
	if err != nil {
		return nil, err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if changed == 0 {
		return nil, nil
	}
	return s.GetConsensus(id)
}

// UpdateConsensus 更新共识标题和描述。
func (s *Storage) UpdateConsensus(id, title, description string) (*domain.Consensus, error) {
	current, err := s.GetConsensus(id)
	if err != nil || current == nil {
		return current, err
	}
	if current.Status != "proposed" {
		return nil, ErrConsensusImmutable
	}

	now := time.Now()
	result, err := s.db.Exec(
		"UPDATE consensuses SET title = ?, description = ?, updated_at = ? WHERE id = ? AND status = 'proposed'",
		title, description, now.Format(time.RFC3339), id,
	)
	if err != nil {
		return nil, err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if changed == 0 {
		latest, err := s.GetConsensus(id)
		if err != nil {
			return nil, err
		}
		if latest != nil && latest.Status != "proposed" {
			return nil, ErrConsensusImmutable
		}
		return latest, nil
	}
	return s.GetConsensus(id)
}

// AddRelation 添加关系。
func (s *Storage) AddRelation(r *domain.Relation) error {
	_, err := s.db.Exec(
		"INSERT INTO relations (id, message_id, consensus_id) VALUES (?, ?, ?)",
		r.ID, r.MessageID, r.ConsensusID,
	)
	return err
}

// RemoveRelation 删除关系。
func (s *Storage) RemoveRelation(id string) error {
	_, err := s.db.Exec("DELETE FROM relations WHERE id = ?", id)
	return err
}

// GetRelationsForConsensus 获取共识的关系。
func (s *Storage) GetRelationsForConsensus(consensusID string) ([]*domain.Relation, error) {
	rows, err := s.db.Query("SELECT * FROM relations WHERE consensus_id = ?", consensusID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relations []*domain.Relation
	for rows.Next() {
		r, err := scanRelationRows(rows)
		if err != nil {
			return nil, err
		}
		relations = append(relations, r)
	}
	return relations, nil
}

// GetRelationsForMessage 获取消息的关系。
func (s *Storage) GetRelationsForMessage(messageID string) ([]*domain.Relation, error) {
	rows, err := s.db.Query("SELECT * FROM relations WHERE message_id = ?", messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relations []*domain.Relation
	for rows.Next() {
		r, err := scanRelationRows(rows)
		if err != nil {
			return nil, err
		}
		relations = append(relations, r)
	}
	return relations, nil
}

// AddConsensusRelation 添加共识之间的关系。
func (s *Storage) AddConsensusRelation(r *domain.ConsensusRelation) error {
	if r.From == r.To {
		return ErrConsensusGraphEdge
	}
	from, err := s.GetConsensus(r.From)
	if err != nil {
		return err
	}
	to, err := s.GetConsensus(r.To)
	if err != nil {
		return err
	}
	if from == nil || to == nil || r.RelationType == "" {
		return ErrConsensusGraphEdge
	}

	_, err = s.db.Exec(
		"INSERT INTO consensus_relations (id, from_id, to_id, relation_type) VALUES (?, ?, ?, ?)",
		r.ID, r.From, r.To, r.RelationType,
	)
	return err
}

// GetConsensusRelation 获取共识关系。
func (s *Storage) GetConsensusRelation(id string) (*domain.ConsensusRelation, error) {
	row := s.db.QueryRow(
		"SELECT id, from_id, to_id, relation_type FROM consensus_relations WHERE id = ?",
		id,
	)
	var relation domain.ConsensusRelation
	if err := row.Scan(&relation.ID, &relation.From, &relation.To, &relation.RelationType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &relation, nil
}

// ListConsensusRelations 列出与共识有关的关系。
func (s *Storage) ListConsensusRelations(consensusID, direction, relationType string) ([]*domain.ConsensusRelation, error) {
	query := "SELECT id, from_id, to_id, relation_type FROM consensus_relations"
	args := make([]any, 0, 3)
	conditions := make([]string, 0, 2)

	if consensusID != "" {
		switch direction {
		case "outgoing":
			conditions = append(conditions, "from_id = ?")
			args = append(args, consensusID)
		case "incoming":
			conditions = append(conditions, "to_id = ?")
			args = append(args, consensusID)
		default:
			conditions = append(conditions, "(from_id = ? OR to_id = ?)")
			args = append(args, consensusID, consensusID)
		}
	}
	if relationType != "" {
		conditions = append(conditions, "relation_type = ?")
		args = append(args, relationType)
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY id"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	relations := make([]*domain.ConsensusRelation, 0)
	for rows.Next() {
		var relation domain.ConsensusRelation
		if err := rows.Scan(&relation.ID, &relation.From, &relation.To, &relation.RelationType); err != nil {
			return nil, err
		}
		relations = append(relations, &relation)
	}
	return relations, rows.Err()
}

// RemoveConsensusRelation 删除共识关系及其图内引用。
func (s *Storage) RemoveConsensusRelation(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM consensus_graph_edges WHERE relation_id = ?", id); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.Exec("DELETE FROM consensus_relations WHERE id = ?", id); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// AddConsensusGraph 添加共识图。
func (s *Storage) AddConsensusGraph(graph *domain.ConsensusGraph) error {
	_, err := s.db.Exec(
		"INSERT INTO consensus_graphs (id, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		graph.ID,
		graph.Name,
		graph.Description,
		graph.CreatedAt.Format(time.RFC3339),
		graph.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// UpdateConsensusGraph 更新共识图元数据。
func (s *Storage) UpdateConsensusGraph(id, name, description string) (*domain.ConsensusGraph, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := s.db.Exec(
		"UPDATE consensus_graphs SET name = ?, description = ?, updated_at = ? WHERE id = ?",
		name,
		description,
		now,
		id,
	)
	if err != nil {
		return nil, err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if changed == 0 {
		return nil, nil
	}
	return s.GetConsensusGraph(id)
}

// ListConsensusGraphs 列出共识图元数据。
func (s *Storage) ListConsensusGraphs() ([]*domain.ConsensusGraph, error) {
	rows, err := s.db.Query(
		"SELECT id, name, description, created_at, updated_at FROM consensus_graphs ORDER BY created_at",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	graphs := make([]*domain.ConsensusGraph, 0)
	for rows.Next() {
		graph, err := scanConsensusGraphRows(rows)
		if err != nil {
			return nil, err
		}
		graphs = append(graphs, graph)
	}
	return graphs, rows.Err()
}

// GetConsensusGraph 获取共识图完整内容，并按拓扑顺序返回节点。
func (s *Storage) GetConsensusGraph(id string) (*domain.ConsensusGraph, error) {
	row := s.db.QueryRow(
		"SELECT id, name, description, created_at, updated_at FROM consensus_graphs WHERE id = ?",
		id,
	)
	graph, err := scanConsensusGraph(row)
	if err != nil || graph == nil {
		return graph, err
	}

	graph.Nodes, err = s.listGraphNodes(id)
	if err != nil {
		return nil, err
	}
	graph.NodePositions, err = s.listGraphNodePositions(id)
	if err != nil {
		return nil, err
	}
	graph.Edges, err = s.listGraphEdges(id)
	if err != nil {
		return nil, err
	}
	graph.Nodes, err = topologicalConsensusOrder(graph.Nodes, graph.Edges)
	if err != nil {
		return nil, err
	}
	return graph, nil
}

// AddConsensusGraphNode 将已有共识加入图。
func (s *Storage) AddConsensusGraphNode(graphID, consensusID string) (*domain.ConsensusGraph, error) {
	graph, err := s.GetConsensusGraph(graphID)
	if err != nil || graph == nil {
		return graph, err
	}
	consensus, err := s.GetConsensus(consensusID)
	if err != nil {
		return nil, err
	}
	if consensus == nil {
		return nil, nil
	}
	if _, err := s.db.Exec(
		"INSERT OR IGNORE INTO consensus_graph_nodes (graph_id, consensus_id) VALUES (?, ?)",
		graphID,
		consensusID,
	); err != nil {
		return nil, err
	}
	if err := s.touchConsensusGraph(graphID); err != nil {
		return nil, err
	}
	return s.GetConsensusGraph(graphID)
}

// UpdateConsensusGraphNodePosition 保存图中节点的画布位置。
func (s *Storage) UpdateConsensusGraphNodePosition(
	graphID,
	consensusID string,
	position domain.ConsensusGraphNodePosition,
) (*domain.ConsensusGraph, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	result, err := tx.Exec(
		`UPDATE consensus_graph_nodes
		SET position_x = ?, position_y = ?
		WHERE graph_id = ? AND consensus_id = ?`,
		position.X,
		position.Y,
		graphID,
		consensusID,
	)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if changed == 0 {
		_ = tx.Rollback()
		return nil, nil
	}
	if _, err := tx.Exec(
		"UPDATE consensus_graphs SET updated_at = ? WHERE id = ?",
		time.Now().UTC().Format(time.RFC3339),
		graphID,
	); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.GetConsensusGraph(graphID)
}

// RemoveConsensusGraphNode 从图中移除共识及其相关图边。
func (s *Storage) RemoveConsensusGraphNode(graphID, consensusID string) (*domain.ConsensusGraph, error) {
	graph, err := s.GetConsensusGraph(graphID)
	if err != nil || graph == nil {
		return graph, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(`
		DELETE FROM consensus_graph_edges
		WHERE graph_id = ?
		AND relation_id IN (
			SELECT e.relation_id
			FROM consensus_graph_edges e
			JOIN consensus_relations r ON r.id = e.relation_id
			WHERE e.graph_id = ? AND (r.from_id = ? OR r.to_id = ?)
		)
	`, graphID, graphID, consensusID, consensusID)
	if err == nil {
		_, err = tx.Exec(
			"DELETE FROM consensus_graph_nodes WHERE graph_id = ? AND consensus_id = ?",
			graphID,
			consensusID,
		)
	}
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	if err := s.touchConsensusGraph(graphID); err != nil {
		return nil, err
	}
	return s.GetConsensusGraph(graphID)
}

// AddConsensusGraphEdge 将已有关系加入图，并拒绝自环和成环关系。
func (s *Storage) AddConsensusGraphEdge(graphID, relationID string) (*domain.ConsensusGraph, error) {
	graph, err := s.GetConsensusGraph(graphID)
	if err != nil || graph == nil {
		return graph, err
	}
	relation, err := s.GetConsensusRelation(relationID)
	if err != nil {
		return nil, err
	}
	if relation == nil || relation.From == relation.To {
		return nil, ErrConsensusGraphEdge
	}

	nodeIDs := make(map[string]struct{}, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodeIDs[node.ID] = struct{}{}
	}
	if _, ok := nodeIDs[relation.From]; !ok {
		return nil, ErrConsensusGraphEdge
	}
	if _, ok := nodeIDs[relation.To]; !ok {
		return nil, ErrConsensusGraphEdge
	}
	for _, edge := range graph.Edges {
		if edge.ID == relationID {
			return graph, nil
		}
	}

	edges := append(append([]*domain.ConsensusRelation{}, graph.Edges...), relation)
	if _, err := topologicalConsensusOrder(graph.Nodes, edges); err != nil {
		return nil, ErrConsensusGraphCycle
	}
	if _, err := s.db.Exec(
		"INSERT INTO consensus_graph_edges (graph_id, relation_id) VALUES (?, ?)",
		graphID,
		relationID,
	); err != nil {
		return nil, err
	}
	if err := s.touchConsensusGraph(graphID); err != nil {
		return nil, err
	}
	return s.GetConsensusGraph(graphID)
}

// AddConsensusGraphRelation 原子创建关系并将其加入图。
func (s *Storage) AddConsensusGraphRelation(graphID string, relation *domain.ConsensusRelation) (*domain.ConsensusGraph, error) {
	graph, err := s.GetConsensusGraph(graphID)
	if err != nil || graph == nil {
		return graph, err
	}
	if relation.From == relation.To || relation.RelationType == "" {
		return nil, ErrConsensusGraphEdge
	}

	nodeIDs := make(map[string]struct{}, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodeIDs[node.ID] = struct{}{}
	}
	if _, ok := nodeIDs[relation.From]; !ok {
		return nil, ErrConsensusGraphEdge
	}
	if _, ok := nodeIDs[relation.To]; !ok {
		return nil, ErrConsensusGraphEdge
	}
	edges := append(append([]*domain.ConsensusRelation{}, graph.Edges...), relation)
	if _, err := topologicalConsensusOrder(graph.Nodes, edges); err != nil {
		return nil, ErrConsensusGraphCycle
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(
		"INSERT INTO consensus_relations (id, from_id, to_id, relation_type) VALUES (?, ?, ?, ?)",
		relation.ID, relation.From, relation.To, relation.RelationType,
	); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if _, err := tx.Exec(
		"INSERT INTO consensus_graph_edges (graph_id, relation_id) VALUES (?, ?)",
		graphID, relation.ID,
	); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	if err := s.touchConsensusGraph(graphID); err != nil {
		return nil, err
	}
	return s.GetConsensusGraph(graphID)
}

// RemoveConsensusGraphEdge 从图中移除边，但保留全局关系。
func (s *Storage) RemoveConsensusGraphEdge(graphID, relationID string) (*domain.ConsensusGraph, error) {
	graph, err := s.GetConsensusGraph(graphID)
	if err != nil || graph == nil {
		return graph, err
	}
	if _, err := s.db.Exec(
		"DELETE FROM consensus_graph_edges WHERE graph_id = ? AND relation_id = ?",
		graphID,
		relationID,
	); err != nil {
		return nil, err
	}
	if err := s.touchConsensusGraph(graphID); err != nil {
		return nil, err
	}
	return s.GetConsensusGraph(graphID)
}

func (s *Storage) listGraphNodes(graphID string) ([]*domain.Consensus, error) {
	rows, err := s.db.Query(`
		SELECT c.id, c.title, c.description, c.status, c.created_at, c.updated_at
		FROM consensus_graph_nodes n
		JOIN consensuses c ON c.id = n.consensus_id
		WHERE n.graph_id = ?
		ORDER BY c.created_at, c.id
	`, graphID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := make([]*domain.Consensus, 0)
	for rows.Next() {
		node, err := scanConsensusRows(rows)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

func (s *Storage) listGraphNodePositions(
	graphID string,
) (map[string]domain.ConsensusGraphNodePosition, error) {
	rows, err := s.db.Query(`
		SELECT consensus_id, position_x, position_y
		FROM consensus_graph_nodes
		WHERE graph_id = ? AND position_x IS NOT NULL AND position_y IS NOT NULL
	`, graphID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	positions := make(map[string]domain.ConsensusGraphNodePosition)
	for rows.Next() {
		var consensusID string
		var position domain.ConsensusGraphNodePosition
		if err := rows.Scan(&consensusID, &position.X, &position.Y); err != nil {
			return nil, err
		}
		positions[consensusID] = position
	}
	return positions, rows.Err()
}

func (s *Storage) listGraphEdges(graphID string) ([]*domain.ConsensusRelation, error) {
	rows, err := s.db.Query(`
		SELECT r.id, r.from_id, r.to_id, r.relation_type
		FROM consensus_graph_edges e
		JOIN consensus_relations r ON r.id = e.relation_id
		WHERE e.graph_id = ?
		ORDER BY r.id
	`, graphID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	edges := make([]*domain.ConsensusRelation, 0)
	for rows.Next() {
		var edge domain.ConsensusRelation
		if err := rows.Scan(&edge.ID, &edge.From, &edge.To, &edge.RelationType); err != nil {
			return nil, err
		}
		edges = append(edges, &edge)
	}
	return edges, rows.Err()
}

func (s *Storage) touchConsensusGraph(graphID string) error {
	_, err := s.db.Exec(
		"UPDATE consensus_graphs SET updated_at = ? WHERE id = ?",
		time.Now().UTC().Format(time.RFC3339),
		graphID,
	)
	return err
}

func scanConsensusGraph(row *sql.Row) (*domain.ConsensusGraph, error) {
	var graph domain.ConsensusGraph
	var createdAt string
	var updatedAt string
	if err := row.Scan(&graph.ID, &graph.Name, &graph.Description, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	graph.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	graph.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	graph.Nodes = make([]*domain.Consensus, 0)
	graph.Edges = make([]*domain.ConsensusRelation, 0)
	graph.NodePositions = make(map[string]domain.ConsensusGraphNodePosition)
	return &graph, nil
}

func scanConsensusGraphRows(rows *sql.Rows) (*domain.ConsensusGraph, error) {
	var graph domain.ConsensusGraph
	var createdAt string
	var updatedAt string
	if err := rows.Scan(&graph.ID, &graph.Name, &graph.Description, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	graph.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	graph.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	graph.Nodes = make([]*domain.Consensus, 0)
	graph.Edges = make([]*domain.ConsensusRelation, 0)
	graph.NodePositions = make(map[string]domain.ConsensusGraphNodePosition)
	return &graph, nil
}

func topologicalConsensusOrder(nodes []*domain.Consensus, edges []*domain.ConsensusRelation) ([]*domain.Consensus, error) {
	nodeByID := make(map[string]*domain.Consensus, len(nodes))
	indegree := make(map[string]int, len(nodes))
	adjacency := make(map[string][]string, len(nodes))
	for _, node := range nodes {
		nodeByID[node.ID] = node
		indegree[node.ID] = 0
	}
	for _, edge := range edges {
		if _, fromExists := nodeByID[edge.From]; !fromExists {
			return nil, ErrConsensusGraphEdge
		}
		if _, toExists := nodeByID[edge.To]; !toExists {
			return nil, ErrConsensusGraphEdge
		}
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
		indegree[edge.To]++
	}

	ordered := make([]*domain.Consensus, 0, len(nodes))
	remaining := append([]*domain.Consensus{}, nodes...)
	for len(remaining) > 0 {
		next := -1
		for index, node := range remaining {
			if indegree[node.ID] == 0 {
				next = index
				break
			}
		}
		if next == -1 {
			return nil, ErrConsensusGraphCycle
		}
		node := remaining[next]
		remaining = append(remaining[:next], remaining[next+1:]...)
		ordered = append(ordered, node)
		for _, target := range adjacency[node.ID] {
			indegree[target]--
		}
	}
	return ordered, nil
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func scanMessage(row *sql.Row) (*domain.Message, error) {
	var msg domain.Message
	var createdAt string
	var updatedAt sql.NullString
	err := row.Scan(&msg.ID, &msg.Content, &msg.Type, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	parsedCreatedAt, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, err
	}
	msg.CreatedAt = parsedCreatedAt
	if updatedAt.Valid {
		t, err := time.Parse(time.RFC3339, updatedAt.String)
		if err == nil {
			msg.UpdatedAt = &t
		}
	}
	return &msg, nil
}

func scanMessageRows(rows *sql.Rows) (*domain.Message, error) {
	var msg domain.Message
	var createdAt string
	var updatedAt sql.NullString
	err := rows.Scan(&msg.ID, &msg.Content, &msg.Type, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	parsedCreatedAt, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, err
	}
	msg.CreatedAt = parsedCreatedAt
	if updatedAt.Valid {
		t, err := time.Parse(time.RFC3339, updatedAt.String)
		if err == nil {
			msg.UpdatedAt = &t
		}
	}
	return &msg, nil
}

func scanConsensus(row *sql.Row) (*domain.Consensus, error) {
	var c domain.Consensus
	var createdAt string
	var updatedAt string
	err := row.Scan(&c.ID, &c.Title, &c.Description, &c.Status, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	created, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, err
	}
	updated, err := time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		updated = created
	}
	c.CreatedAt = created
	c.UpdatedAt = updated
	return &c, nil
}

func scanConsensusRows(rows *sql.Rows) (*domain.Consensus, error) {
	var c domain.Consensus
	var createdAt string
	var updatedAt string
	err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.Status, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	created, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, err
	}
	updated, err := time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		updated = created
	}
	c.CreatedAt = created
	c.UpdatedAt = updated
	return &c, nil
}

func scanRelationRows(rows *sql.Rows) (*domain.Relation, error) {
	var r domain.Relation
	err := rows.Scan(&r.ID, &r.MessageID, &r.ConsensusID)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
