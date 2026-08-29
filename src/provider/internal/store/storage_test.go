package store

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/quanttide/qtcloud-connect/provider/internal/domain"
)

func TestNewCreatesParentDirectory(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "nested", "qtcloud-connect.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer s.Close()

	consensuses, err := s.ListConsensuses(nil)
	if err != nil {
		t.Fatalf("ListConsensuses() error = %v", err)
	}
	if len(consensuses) != 0 {
		t.Fatalf("ListConsensuses() len = %d", len(consensuses))
	}
}

func TestUpdateConsensusReturnsNilForMissingRecord(t *testing.T) {
	s, err := New(filepath.Join(t.TempDir(), "qtcloud-connect.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer s.Close()

	consensus, err := s.UpdateConsensus("missing", "title", "description")
	if err != nil {
		t.Fatalf("UpdateConsensus() error = %v", err)
	}
	if consensus != nil {
		t.Fatalf("UpdateConsensus() = %#v", consensus)
	}
}

func TestListConsensusesFiltersByStatus(t *testing.T) {
	s, err := New(filepath.Join(t.TempDir(), "qtcloud-connect.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer s.Close()

	now := time.Date(2026, time.August, 29, 0, 0, 0, 0, time.UTC)
	for _, consensus := range []*domain.Consensus{
		{ID: "c1", Title: "待确认", Status: "proposed", CreatedAt: now, UpdatedAt: now},
		{ID: "c2", Title: "已确认", Status: "confirmed", CreatedAt: now, UpdatedAt: now},
	} {
		if err := s.AddConsensus(consensus); err != nil {
			t.Fatalf("AddConsensus() error = %v", err)
		}
	}

	status := "confirmed"
	consensuses, err := s.ListConsensuses(&status)
	if err != nil {
		t.Fatalf("ListConsensuses() error = %v", err)
	}
	if len(consensuses) != 1 || consensuses[0].ID != "c2" {
		t.Fatalf("ListConsensuses() = %#v", consensuses)
	}
}

func TestNewMigratesLegacyConsensusContentToTitle(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "qtcloud-connect.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open legacy database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE consensuses (
			id TEXT PRIMARY KEY,
			content TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'proposed',
			created_at TEXT NOT NULL,
			updated_at TEXT
		);
		INSERT INTO consensuses (id, content, status, created_at, updated_at)
		VALUES ('c1', '旧版共识内容', 'confirmed', '2026-08-29T00:00:00Z', '');
	`)
	if err != nil {
		t.Fatalf("seed legacy database: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close legacy database: %v", err)
	}

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer s.Close()

	consensus, err := s.GetConsensus("c1")
	if err != nil {
		t.Fatalf("GetConsensus() error = %v", err)
	}
	if consensus == nil || consensus.Title != "旧版共识内容" {
		t.Fatalf("GetConsensus() = %#v", consensus)
	}
}

func TestNewMigratesLegacyConsensusNullUpdatedAt(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "qtcloud-connect.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open legacy database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE consensuses (
			id TEXT PRIMARY KEY,
			content TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'proposed',
			created_at TEXT NOT NULL,
			updated_at TEXT
		);
		INSERT INTO consensuses (id, content, status, created_at, updated_at)
		VALUES ('c1', '旧版空更新时间', 'confirmed', '2026-08-29T00:00:00Z', NULL);
	`)
	if err != nil {
		t.Fatalf("seed legacy database: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close legacy database: %v", err)
	}

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer s.Close()

	consensus, err := s.GetConsensus("c1")
	if err != nil {
		t.Fatalf("GetConsensus() error = %v", err)
	}
	if consensus == nil || !consensus.UpdatedAt.Equal(consensus.CreatedAt) {
		t.Fatalf("GetConsensus() = %#v", consensus)
	}
}
