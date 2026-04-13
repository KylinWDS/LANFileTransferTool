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

// HistoryRecord 历史记录结构体（扩展版）
type HistoryRecord struct {
	ID           int
	FileName     string
	FileSize     int64
	FilePath     string    // 文件绝对路径
	Action       string    // 动作：upload/download/share
	Status       string    // 状态：completed/failed/pending
	Protocol     string    // 使用的传输协议
	DownloadLink string    // 下载链接
	Duration     int64     // 传输时长（秒）
	CreatedAt    time.Time
	UpdatedAt    time.Time // 更新时间（用于延长链接）
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

	// 迁移旧表结构
	if err := s.migrateTables(); err != nil {
		logger.Warn("数据库迁移失败: %v", err)
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
		file_path TEXT,
		action TEXT NOT NULL,
		status TEXT NOT NULL,
		protocol TEXT,
		download_link TEXT,
		duration INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS file_registry (
		id TEXT PRIMARY KEY,
		file_path TEXT NOT NULL,
		file_name TEXT NOT NULL,
		file_size INTEGER NOT NULL,
		checksum TEXT,
		content_type TEXT,
		download_count INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME
	);

	CREATE INDEX IF NOT EXISTS idx_history_created_at ON history(created_at);
	CREATE INDEX IF NOT EXISTS idx_file_registry_expires ON file_registry(expires_at);
	`

	_, err := s.db.Exec(query)
	return err
}

// migrateTables 迁移旧表结构
func (s *Storage) migrateTables() error {
	// 检查是否需要添加新列到history表
	historyColumns := []string{
		"file_path",
		"protocol",
		"download_link",
		"duration",
		"updated_at",
	}

	for _, col := range historyColumns {
		// 尝试添加列，如果已存在会报错但忽略
		query := fmt.Sprintf(`ALTER TABLE history ADD COLUMN %s TEXT`, col)
		if col == "file_size" || col == "duration" {
			query = fmt.Sprintf(`ALTER TABLE history ADD COLUMN %s INTEGER`, col)
		}
		s.db.Exec(query) // 忽略错误
	}

	// 检查是否需要添加新列到file_registry表
	registryColumns := []struct {
		name string
		typ  string
	}{
		{"download_count", "INTEGER"},
	}

	for _, col := range registryColumns {
		query := fmt.Sprintf(`ALTER TABLE file_registry ADD COLUMN %s %s`, col.name, col.typ)
		s.db.Exec(query) // 忽略错误
	}

	return nil
}

func (s *Storage) AddHistory(record *HistoryRecord) error {
	query := `INSERT INTO history 
		(file_name, file_size, file_path, action, status, protocol, download_link, duration, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	now := time.Now()
	if record.CreatedAt.IsZero() {
		record.CreatedAt = now
	}
	record.UpdatedAt = now

	result, err := s.db.Exec(query, 
		record.FileName, 
		record.FileSize, 
		record.FilePath,
		record.Action, 
		record.Status,
		record.Protocol,
		record.DownloadLink,
		record.Duration,
		record.CreatedAt,
		record.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, _ := result.LastInsertId()
	record.ID = int(id)
	return nil
}

func (s *Storage) UpdateHistory(id int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// 自动更新 updated_at
	updates["updated_at"] = time.Now()

	query := "UPDATE history SET "
	args := []interface{}{}
	first := true

	for key, value := range updates {
		if !first {
			query += ", "
		}
		query += key + " = ?"
		args = append(args, value)
		first = false
	}

	query += " WHERE id = ?"
	args = append(args, id)

	_, err := s.db.Exec(query, args...)
	return err
}

func (s *Storage) GetHistory(limit int) ([]*HistoryRecord, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `SELECT id, file_name, file_size, file_path, action, status, protocol, 
		download_link, duration, created_at, updated_at 
		FROM history ORDER BY created_at DESC LIMIT ?`
	
	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*HistoryRecord
	for rows.Next() {
		record := &HistoryRecord{}
		var filePath, protocol, downloadLink sql.NullString
		var duration sql.NullInt64
		
		err := rows.Scan(
			&record.ID, 
			&record.FileName, 
			&record.FileSize,
			&filePath,
			&record.Action, 
			&record.Status,
			&protocol,
			&downloadLink,
			&duration,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		record.FilePath = filePath.String
		record.Protocol = protocol.String
		record.DownloadLink = downloadLink.String
		record.Duration = duration.Int64
		
		records = append(records, record)
	}

	return records, nil
}

func (s *Storage) GetHistoryByID(id int) (*HistoryRecord, error) {
	query := `SELECT id, file_name, file_size, file_path, action, status, protocol, 
		download_link, duration, created_at, updated_at 
		FROM history WHERE id = ?`
	
	row := s.db.QueryRow(query, id)
	record := &HistoryRecord{}
	
	var filePath, protocol, downloadLink sql.NullString
	var duration sql.NullInt64
	
	err := row.Scan(
		&record.ID, 
		&record.FileName, 
		&record.FileSize,
		&filePath,
		&record.Action, 
		&record.Status,
		&protocol,
		&downloadLink,
		&duration,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	record.FilePath = filePath.String
	record.Protocol = protocol.String
	record.DownloadLink = downloadLink.String
	record.Duration = duration.Int64
	
	return record, nil
}

func (s *Storage) ClearHistory() error {
	query := `DELETE FROM history`
	_, err := s.db.Exec(query)
	return err
}

func (s *Storage) DeleteHistory(id int) error {
	query := `DELETE FROM history WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

func (s *Storage) RegisterFile(id, filePath, fileName string, fileSize int64, checksum, contentType string, expiresAt *time.Time) error {
	query := `INSERT OR REPLACE INTO file_registry (id, file_path, file_name, file_size, checksum, content_type, created_at, expires_at) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)`
	_, err := s.db.Exec(query, id, filePath, fileName, fileSize, checksum, contentType, expiresAt)
	return err
}

func (s *Storage) GetFileInfo(id string) (*FileInfo, error) {
	query := `SELECT id, file_path, file_name, file_size, checksum, content_type, download_count, created_at, expires_at FROM file_registry WHERE id = ?`
	row := s.db.QueryRow(query, id)

	info := &FileInfo{}
	err := row.Scan(&info.ID, &info.Path, &info.Name, &info.Size, &info.Checksum, &info.ContentType, &info.DownloadCount, &info.CreatedAt, &info.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (s *Storage) GetAvailableFiles() ([]*FileInfo, error) {
	query := `SELECT id, file_path, file_name, file_size, checksum, content_type, download_count, created_at, expires_at FROM file_registry WHERE expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*FileInfo
	for rows.Next() {
		info := &FileInfo{}
		if err := rows.Scan(&info.ID, &info.Path, &info.Name, &info.Size, &info.Checksum, &info.ContentType, &info.DownloadCount, &info.CreatedAt, &info.ExpiresAt); err != nil {
			return nil, err
		}
		files = append(files, info)
	}

	return files, nil
}

// IncrementDownloadCount 增加文件下载次数
func (s *Storage) IncrementDownloadCount(filePath string) error {
	query := `UPDATE file_registry SET download_count = download_count + 1 WHERE file_path = ?`
	_, err := s.db.Exec(query, filePath)
	return err
}

// GetDownloadCount 获取文件下载次数
func (s *Storage) GetDownloadCount(filePath string) (int, error) {
	query := `SELECT download_count FROM file_registry WHERE file_path = ?`
	row := s.db.QueryRow(query, filePath)

	var count int
	err := row.Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetDownloadCountByID 根据文件ID获取下载次数
func (s *Storage) GetDownloadCountByID(fileID string) (int, error) {
	query := `SELECT download_count FROM file_registry WHERE id = ?`
	row := s.db.QueryRow(query, fileID)

	var count int
	err := row.Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return count, nil
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
	ID            string
	Path          string
	Name          string
	Size          int64
	Checksum      string
	ContentType   string
	DownloadCount int
	CreatedAt     time.Time
	ExpiresAt     *time.Time
}
