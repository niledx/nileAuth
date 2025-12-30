package main

import (
	"database/sql"
	"errors"
)

// DB interface for database operations
type DB interface {
	Init() error
	// User operations
	CreateUser(email, password string, applicationID *int64) (*User, error)
	GetUserByEmail(email string) (*User, error)
	// Token operations
	CreateRefreshToken(token string, userId int64, expiresAt int64, applicationID *int64) error
	GetRefreshToken(token string) (*RefreshToken, error)
	RevokeRefreshToken(token string) error
	RevokeAllRefreshTokensForUser(userId int64) error
	// Application operations
	GetApplicationByAPIKeyPrefix(prefix string) ([]*Application, error)
	GetApplicationByID(id int64) (*Application, error)
	CreateApplication(name, domain, apiKeyHash, apiKeyPrefix string, rateLimit int, origins []string) (*Application, error)
	// Scope operations
	GetScopesByApplicationID(applicationID int64) ([]*Scope, error)
	GetScopeByName(name string) (*Scope, error)
}

// Memory DB
type MemDB struct {
	users  map[string]*User
	tokens map[string]*RefreshToken
	seq    int64
}

func NewMemoryDB() *MemDB {
	return &MemDB{users: map[string]*User{}, tokens: map[string]*RefreshToken{}, seq: 1}
}

func (m *MemDB) Init() error { return nil }
func (m *MemDB) CreateUser(email, password string, applicationID *int64) (*User, error) {
	if _, ok := m.users[email]; ok {
		return nil, errors.New("exists")
	}
	u := &User{ID: m.seq, Email: email, Password: password, ApplicationID: applicationID}
	m.seq++
	m.users[email] = u
	return u, nil
}
func (m *MemDB) GetUserByEmail(email string) (*User, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, nil
}
func (m *MemDB) CreateRefreshToken(token string, userId int64, expiresAt int64, applicationID *int64) error {
	m.tokens[token] = &RefreshToken{Token: token, UserID: userId, ExpiresAt: expiresAt, ApplicationID: applicationID}
	return nil
}
func (m *MemDB) GetRefreshToken(token string) (*RefreshToken, error) {
	if t, ok := m.tokens[token]; ok {
		return t, nil
	}
	return nil, nil
}
func (m *MemDB) RevokeRefreshToken(token string) error {
	if t, ok := m.tokens[token]; ok {
		t.Revoked = true
		return nil
	}
	return nil
}
func (m *MemDB) RevokeAllRefreshTokensForUser(userId int64) error {
	for _, t := range m.tokens {
		if t.UserID == userId {
			t.Revoked = true
		}
	}
	return nil
}

// Enterprise features for Memory DB (simplified implementations)
func (m *MemDB) GetApplicationByAPIKeyPrefix(prefix string) ([]*Application, error) {
	// Memory DB: return empty for now (can be extended with in-memory storage)
	return []*Application{}, nil
}

func (m *MemDB) GetApplicationByID(id int64) (*Application, error) {
	return nil, nil
}

func (m *MemDB) CreateApplication(name, domain, apiKeyHash, apiKeyPrefix string, rateLimit int, origins []string) (*Application, error) {
	return nil, errors.New("not implemented in memory DB")
}

func (m *MemDB) GetScopesByApplicationID(applicationID int64) ([]*Scope, error) {
	return []*Scope{}, nil
}

func (m *MemDB) GetScopeByName(name string) (*Scope, error) {
	return nil, nil
}

// SQLite DB
type SQLiteDB struct {
	db   *sql.DB
	path string
}

func NewSQLiteDB(path string) (*SQLiteDB, error) {
	d, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	s := &SQLiteDB{db: d, path: path}
	if err := s.Init(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SQLiteDB) Init() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT UNIQUE, password TEXT, application_id INTEGER, created_at TEXT);`,
		`CREATE TABLE IF NOT EXISTS refresh_tokens (token TEXT PRIMARY KEY, user_id INTEGER, application_id INTEGER, expires_at INTEGER, revoked INTEGER DEFAULT 0, created_at TEXT);`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// Enterprise features for SQLite DB
func (s *SQLiteDB) GetApplicationByAPIKeyPrefix(prefix string) ([]*Application, error) {
	rows, err := s.db.Query(`SELECT id,name,domain,api_key_hash,api_key_prefix,rate_limit_per_minute,active,created_at,updated_at FROM applications WHERE api_key_prefix = ? AND active = 1`, prefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var apps []*Application
	for rows.Next() {
		var app Application
		var active int
		var createdAt, updatedAt string
		if err := rows.Scan(&app.ID, &app.Name, &app.Domain, &app.APIKeyHash, &app.APIKeyPrefix, &app.RateLimitPerMinute, &active, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		app.Active = active != 0
		apps = append(apps, &app)
	}
	return apps, nil
}

func (s *SQLiteDB) GetApplicationByID(id int64) (*Application, error) {
	row := s.db.QueryRow(`SELECT id,name,domain,api_key_hash,api_key_prefix,rate_limit_per_minute,active,created_at,updated_at FROM applications WHERE id = ?`, id)
	var app Application
	var active int
	var createdAt, updatedAt string
	if err := row.Scan(&app.ID, &app.Name, &app.Domain, &app.APIKeyHash, &app.APIKeyPrefix, &app.RateLimitPerMinute, &active, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	app.Active = active != 0
	return &app, nil
}

func (s *SQLiteDB) CreateApplication(name, domain, apiKeyHash, apiKeyPrefix string, rateLimit int, origins []string) (*Application, error) {
	res, err := s.db.Exec(`INSERT INTO applications(name,domain,api_key_hash,api_key_prefix,rate_limit_per_minute,created_at,updated_at) VALUES(?,?,?,?,?,datetime('now'),datetime('now'))`, name, domain, apiKeyHash, apiKeyPrefix, rateLimit)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &Application{ID: id, Name: name, Domain: domain, APIKeyHash: apiKeyHash, APIKeyPrefix: apiKeyPrefix, RateLimitPerMinute: rateLimit, AllowedOrigins: origins, Active: true}, nil
}

func (s *SQLiteDB) GetScopesByApplicationID(applicationID int64) ([]*Scope, error) {
	rows, err := s.db.Query(`SELECT s.id,s.name,s.description,s.created_at FROM scopes s JOIN application_scopes aps ON s.id = aps.scope_id WHERE aps.application_id = ?`, applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var scopes []*Scope
	for rows.Next() {
		var scope Scope
		var createdAt string
		if err := rows.Scan(&scope.ID, &scope.Name, &scope.Description, &createdAt); err != nil {
			return nil, err
		}
		scopes = append(scopes, &scope)
	}
	return scopes, nil
}

func (s *SQLiteDB) GetScopeByName(name string) (*Scope, error) {
	row := s.db.QueryRow(`SELECT id,name,description,created_at FROM scopes WHERE name = ?`, name)
	var scope Scope
	var createdAt string
	if err := row.Scan(&scope.ID, &scope.Name, &scope.Description, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &scope, nil
}

func (s *SQLiteDB) CreateUser(email, password string, applicationID *int64) (*User, error) {
	res, err := s.db.Exec(`INSERT INTO users(email,password,application_id,created_at) VALUES(?,?,?,datetime('now'))`, email, password, applicationID)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &User{ID: id, Email: email, Password: password, ApplicationID: applicationID}, nil
}

func (s *SQLiteDB) GetUserByEmail(email string) (*User, error) {
	row := s.db.QueryRow(`SELECT id,email,password,application_id,created_at FROM users WHERE email = ?`, email)
	var u User
	var created string
	var appID sql.NullInt64
	if err := row.Scan(&u.ID, &u.Email, &u.Password, &appID, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if appID.Valid {
		u.ApplicationID = &appID.Int64
	}
	return &u, nil
}

func (s *SQLiteDB) CreateRefreshToken(token string, userId int64, expiresAt int64, applicationID *int64) error {
	_, err := s.db.Exec(`INSERT INTO refresh_tokens(token,user_id,application_id,expires_at,created_at) VALUES(?,?,?,?,datetime('now'))`, token, userId, applicationID, expiresAt)
	return err
}

func (s *SQLiteDB) GetRefreshToken(token string) (*RefreshToken, error) {
	row := s.db.QueryRow(`SELECT token,user_id,application_id,expires_at,revoked FROM refresh_tokens WHERE token = ?`, token)
	var t RefreshToken
	var revoked int
	var appID sql.NullInt64
	if err := row.Scan(&t.Token, &t.UserID, &appID, &t.ExpiresAt, &revoked); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	t.Revoked = revoked != 0
	if appID.Valid {
		t.ApplicationID = &appID.Int64
	}
	return &t, nil
}

func (s *SQLiteDB) RevokeRefreshToken(token string) error {
	_, err := s.db.Exec(`UPDATE refresh_tokens SET revoked = 1 WHERE token = ?`, token)
	return err
}

func (s *SQLiteDB) RevokeAllRefreshTokensForUser(userId int64) error {
	_, err := s.db.Exec(`UPDATE refresh_tokens SET revoked = 1 WHERE user_id = ?`, userId)
	return err
}

// lifecycle helpers
func (m *MemDB) close() error { return nil }
func (m *MemDB) ping() bool   { return true }

func (s *SQLiteDB) close() error { return s.db.Close() }
func (s *SQLiteDB) ping() bool   { return s.db.Ping() == nil }
