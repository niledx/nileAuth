package main

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	db  *sql.DB
	dsn string
}

func NewPostgresDB(dsn string) (*PostgresDB, error) {
	d, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	p := &PostgresDB{db: d, dsn: dsn}
	if err := p.Init(); err != nil {
		d.Close()
		return nil, err
	}
	return p, nil
}

func (p *PostgresDB) Init() error {
	// rely on migrations to create tables; just verify connectivity
	if err := p.db.Ping(); err != nil {
		return err
	}
	return nil
}

func (p *PostgresDB) CreateUser(email, password string, applicationID *int64) (*User, error) {
	var id int64
	err := p.db.QueryRow(`INSERT INTO users(email,password,application_id,created_at) VALUES($1,$2,$3,now()) RETURNING id`, email, password, applicationID).Scan(&id)
	if err != nil {
		// unique violation
		return nil, err
	}
	return &User{ID: id, Email: email, Password: password, ApplicationID: applicationID}, nil
}

func (p *PostgresDB) GetUserByEmail(email string) (*User, error) {
	row := p.db.QueryRow(`SELECT id,email,password,application_id,created_at FROM users WHERE email = $1`, email)
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

func (p *PostgresDB) CreateRefreshToken(token string, userId int64, expiresAt int64, applicationID *int64) error {
	_, err := p.db.Exec(`INSERT INTO refresh_tokens(token,user_id,application_id,expires_at,created_at) VALUES($1,$2,$3,$4,now())`, token, userId, applicationID, expiresAt)
	return err
}

func (p *PostgresDB) GetRefreshToken(token string) (*RefreshToken, error) {
	row := p.db.QueryRow(`SELECT token,user_id,application_id,expires_at,revoked FROM refresh_tokens WHERE token = $1`, token)
	var t RefreshToken
	var appID sql.NullInt64
	if err := row.Scan(&t.Token, &t.UserID, &appID, &t.ExpiresAt, &t.Revoked); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if appID.Valid {
		t.ApplicationID = &appID.Int64
	}
	return &t, nil
}

func (p *PostgresDB) RevokeRefreshToken(token string) error {
	res, err := p.db.Exec(`UPDATE refresh_tokens SET revoked = true WHERE token = $1`, token)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("not found")
	}
	return nil
}

func (p *PostgresDB) RevokeAllRefreshTokensForUser(userId int64) error {
	_, err := p.db.Exec(`UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`, userId)
	return err
}

func (p *PostgresDB) close() error { return p.db.Close() }
func (p *PostgresDB) ping() bool   { return p.db.Ping() == nil }

// Enterprise features for Postgres DB
func (p *PostgresDB) GetApplicationByAPIKeyPrefix(prefix string) ([]*Application, error) {
	rows, err := p.db.Query(`SELECT id,name,domain,api_key_hash,api_key_prefix,rate_limit_per_minute,allowed_origins,active,created_at,updated_at FROM applications WHERE api_key_prefix = $1 AND active = true`, prefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var apps []*Application
	for rows.Next() {
		var app Application
		var origins []string
		if err := rows.Scan(&app.ID, &app.Name, &app.Domain, &app.APIKeyHash, &app.APIKeyPrefix, &app.RateLimitPerMinute, &origins, &app.Active, &app.CreatedAt, &app.UpdatedAt); err != nil {
			return nil, err
		}
		app.AllowedOrigins = origins
		apps = append(apps, &app)
	}
	return apps, nil
}

func (p *PostgresDB) GetApplicationByID(id int64) (*Application, error) {
	row := p.db.QueryRow(`SELECT id,name,domain,api_key_hash,api_key_prefix,rate_limit_per_minute,allowed_origins,active,created_at,updated_at FROM applications WHERE id = $1`, id)
	var app Application
	var origins []string
	if err := row.Scan(&app.ID, &app.Name, &app.Domain, &app.APIKeyHash, &app.APIKeyPrefix, &app.RateLimitPerMinute, &origins, &app.Active, &app.CreatedAt, &app.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	app.AllowedOrigins = origins
	return &app, nil
}

func (p *PostgresDB) CreateApplication(name, domain, apiKeyHash, apiKeyPrefix string, rateLimit int, origins []string) (*Application, error) {
	var id int64
	err := p.db.QueryRow(`INSERT INTO applications(name,domain,api_key_hash,api_key_prefix,rate_limit_per_minute,allowed_origins,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,now(),now()) RETURNING id`, name, domain, apiKeyHash, apiKeyPrefix, rateLimit, origins).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &Application{ID: id, Name: name, Domain: domain, APIKeyHash: apiKeyHash, APIKeyPrefix: apiKeyPrefix, RateLimitPerMinute: rateLimit, AllowedOrigins: origins, Active: true}, nil
}

func (p *PostgresDB) GetScopesByApplicationID(applicationID int64) ([]*Scope, error) {
	rows, err := p.db.Query(`SELECT s.id,s.name,s.description,s.created_at FROM scopes s JOIN application_scopes aps ON s.id = aps.scope_id WHERE aps.application_id = $1`, applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var scopes []*Scope
	for rows.Next() {
		var scope Scope
		if err := rows.Scan(&scope.ID, &scope.Name, &scope.Description, &scope.CreatedAt); err != nil {
			return nil, err
		}
		scopes = append(scopes, &scope)
	}
	return scopes, nil
}

func (p *PostgresDB) GetScopeByName(name string) (*Scope, error) {
	row := p.db.QueryRow(`SELECT id,name,description,created_at FROM scopes WHERE name = $1`, name)
	var scope Scope
	if err := row.Scan(&scope.ID, &scope.Name, &scope.Description, &scope.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &scope, nil
}
