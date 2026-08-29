package domain

import "time"

// Message 是消息，沟通的基本载体。
type Message struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	Type      string     `json:"type"` // "user" / "agent" / "system"
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// Consensus 是共识，团队成员通过讨论达成的一致决策或结论。
type Consensus struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status,omitempty"` // "proposed" / "confirmed" / "deprecated"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Relation 是消息与共识之间的多对多溯源关联。
type Relation struct {
	ID          string `json:"id"`
	MessageID   string `json:"message_id"`
	ConsensusID string `json:"consensus_id"`
}

// ConsensusRelation 是共识之间的有向关系。
type ConsensusRelation struct {
	ID           string `json:"id"`
	From         string `json:"from"`
	To           string `json:"to"`
	RelationType string `json:"relation_type"`
}

// ConsensusGraphNodePosition 是一个共识节点在特定图中的画布坐标。
type ConsensusGraphNodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ConsensusGraph 是以 DAG 组织共识节点和关系边的图。
type ConsensusGraph struct {
	ID            string                                `json:"id"`
	Name          string                                `json:"name"`
	Description   string                                `json:"description"`
	Nodes         []*Consensus                          `json:"nodes"`
	Edges         []*ConsensusRelation                  `json:"edges"`
	NodePositions map[string]ConsensusGraphNodePosition `json:"node_positions"`
	CreatedAt     time.Time                             `json:"created_at"`
	UpdatedAt     time.Time                             `json:"updated_at"`
}
