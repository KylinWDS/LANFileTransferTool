package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"lanfiletransfertool/pkg/logger"
)

type Storage struct {
	db *sql.DB
}

type HistoryRecord struct {
	ID        int
	FileName  string
	FileSize  int64
	Action    string
	Status    string
	CreatedAt time.Time
}

func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	s := &Storage{db: db}
	if err := s.initTables(); err != nil {
		return nil, fmt.Errorf("初始化表失败: %w", err)
	}

	logger.Info("数据库初始化成功")
	return s, nil
}

func (s *Storage) initTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_name TEXT NOT NULL,
		file_size INTEGER,
		action TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS file_registry (
		id TEXT PRIMARY KEY,
		file_path TEXT NOT NULL,
		file_name TEXT NOT NULL,
		file_size INTEGER NOT NULL,
		checksum TEXT,
		content_type TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME
	);

	CREATE INDEX IF NOT EXISTS idx_history_created_at ON history(created_at);
	CREATE INDEX IF NOT EXISTS idx_file_registry_expires ON file_registry(expires_at);
	`

	_, err := s.db.Exec(query)
	return err
}

func (s *Storage) AddHistory(record *HistoryRecord) error {
	query := `INSERT INTO history (file_name, file_size, action, status) VALUES (?, ?, ?, ?)`
	result, err := s.db.Exec(query, record.FileName, record.FileSize, record.Action, record.Status)
	if err != nil {
		return err
	}

	id, _ := result.LastInsertId()
	record.ID = int(id)
	return nil
}

func (s *Storage) GetHistory(limit int) ([]*HistoryRecord, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `SELECT id, file_name, file_size, action, status, created_at FROM history ORDER BY created_at DESC LIMIT ?`
	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*HistoryRecord
	for rows.Next() {
		record := &HistoryRecord{}
		if err := rows.Scan(&record.ID, &record.FileName, &record.FileSize, &record.Action, &record.Status, &record.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

func (s *Storage) ClearHistory() error {
	query := `DELETE FROM history`
	_, err := s.db.Exec(query)
	return err
}

func (s *Storage) RegisterFile(id, filePath, fileName string, fileSize int64, checksum, contentType string, expiresAt *time.Time) error {
	query := `INSERT OR REPLACE INTO file_registry (id, file_path, file_name, file_size, checksum, content_type, created_at, expires_at) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)`
	_, err := s.db.Exec(query, id, filePath, fileName, fileSize, checksum, contentType, expiresAt)
	return err
}

func (s *Storage) GetFileInfo(id string) (*FileInfo, error) {
	query := `SELECT id, file_path, file_name, file_size, checksum, content_type, created_at, expires_at FROM file_registry WHERE id = ?`
	row := s.db.QueryRow(query, id)

	info := &FileInfo{}
	err := row.Scan(&info.ID, &info.Path, &info.Name, &info.Size, &info.Checksum, &info.ContentType, &info.CreatedAt, &info.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (s *Storage) GetAvailableFiles() ([]*FileInfo, error) {
	query := `SELECT id, file_path, file_name, file_size, checksum, content_type, created_at, expires_at FROM file_registry WHERE expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*FileInfo
	for rows.Next() {
		info := &FileInfo{}
		if err := rows.Scan(&info.ID, &info.Path, &info.Name, &info.Size, &info.Checksum, &info.ContentType, &info.CreatedAt, &info.ExpiresAt); err != nil {
			return nil, err
		}
		files = append(files, info)
	}

	return files, nil
}

func (s *Storage) RemoveFile(id string) error {
	query := `DELETE FROM file_registry WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}

type FileInfo struct {
	ID          string
	Path        string
	Name        string
	Size        int64
	Checksum    string
	ContentType string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
}
