package store

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/quanttide/qtcloud-connect/provider/internal/domain"
)

// Storage 是 SQLite 存储层。
type Storage struct {
	db *sql.DB
}

// New 创建新的存储实例。
func New(dbPath string) (*Storage, error) {
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

// initTables 初始化数据库表。
func (s *Storage) initTables() error {
	query := `
		CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY, content TEXT NOT NULL, type TEXT NOT NULL,
			created_at TEXT NOT NULL, updated_at TEXT
		);
		CREATE TABLE IF NOT EXISTS consensuses (
			id TEXT PRIMARY KEY, content TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'proposed',
			created_at TEXT NOT NULL, updated_at TEXT
		);
		CREATE TABLE IF NOT EXISTS relations (
			id TEXT PRIMARY KEY, message_id TEXT NOT NULL, consensus_id TEXT NOT NULL
		);
	`
	_, err := s.db.Exec(query)
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

// UpdateMessage 更新消息内容。
func (s *Storage) UpdateMessage(id, content string) (*domain.Message, error) {
	now := time.Now()
	_, err := s.db.Exec(
		"UPDATE messages SET content = ?, updated_at = ? WHERE id = ?",
		content, now.Format(time.RFC3339), id,
	)
	if err != nil {
		return nil, err
	}
	return s.GetMessage(id)
}

// AddConsensus 添加共识。
func (s *Storage) AddConsensus(c *domain.Consensus) error {
	_, err := s.db.Exec(
		"INSERT INTO consensuses (id, content, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		c.ID, c.Content, c.Status, c.CreatedAt.Format(time.RFC3339), formatTime(c.UpdatedAt),
	)
	return err
}

// GetConsensus 获取共识。
func (s *Storage) GetConsensus(id string) (*domain.Consensus, error) {
	row := s.db.QueryRow("SELECT * FROM consensuses WHERE id = ?", id)
	return scanConsensus(row)
}

// ListConsensuses 列出所有共识。
func (s *Storage) ListConsensuses(status *string) ([]*domain.Consensus, error) {
	var rows *sql.Rows
	var err error

	if status != nil {
		rows, err = s.db.Query("SELECT * FROM consensuses WHERE status = ? ORDER BY created_at", *status)
	} else {
		rows, err = s.db.Query("SELECT * FROM consensuses ORDER BY created_at")
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
	_, err := s.db.Exec(
		"UPDATE consensuses SET status = ?, updated_at = ? WHERE id = ?",
		status, now.Format(time.RFC3339), id,
	)
	if err != nil {
		return nil, err
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

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func scanMessage(row *sql.Row) (*domain.Message, error) {
	var msg domain.Message
	var updatedAt sql.NullString
	err := row.Scan(&msg.ID, &msg.Content, &msg.Type, &msg.CreatedAt, &updatedAt)
	if err != nil {
		return nil, err
	}
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
	var updatedAt sql.NullString
	err := rows.Scan(&msg.ID, &msg.Content, &msg.Type, &msg.CreatedAt, &updatedAt)
	if err != nil {
		return nil, err
	}
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
	var updatedAt sql.NullString
	err := row.Scan(&c.ID, &c.Content, &c.Status, &c.CreatedAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	if updatedAt.Valid {
		t, err := time.Parse(time.RFC3339, updatedAt.String)
		if err == nil {
			c.UpdatedAt = &t
		}
	}
	return &c, nil
}

func scanConsensusRows(rows *sql.Rows) (*domain.Consensus, error) {
	var c domain.Consensus
	var updatedAt sql.NullString
	err := rows.Scan(&c.ID, &c.Content, &c.Status, &c.CreatedAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	if updatedAt.Valid {
		t, err := time.Parse(time.RFC3339, updatedAt.String)
		if err == nil {
			c.UpdatedAt = &t
		}
	}
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
