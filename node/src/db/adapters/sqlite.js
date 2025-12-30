const Database = require('better-sqlite3');
let db;

module.exports = {
  init: ({ dbPath }) => {
    const fs = require('fs');
    const path = require('path');
    const dir = path.dirname(dbPath);
    if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
    db = new Database(dbPath);
    db.prepare(
      `CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT UNIQUE,
        password TEXT,
        created_at TEXT
      )`
    ).run();
    db.prepare(
      `CREATE TABLE IF NOT EXISTS refresh_tokens (
         token TEXT PRIMARY KEY,
         user_id INTEGER,
         expires_at INTEGER,
         revoked INTEGER DEFAULT 0,
         created_at TEXT
       )`
    ).run();
  },
  createUser: async ({ email, password }) => {
    const stmt = db.prepare('INSERT INTO users (email,password,created_at) VALUES (?, ?, ?)');
    const info = stmt.run(email, password, new Date().toISOString());
    return { id: info.lastInsertRowid, email, password, created_at: new Date().toISOString() };
  },
  getUserByEmail: async (email) => {
    const stmt = db.prepare('SELECT id,email,password,created_at FROM users WHERE email = ?');
    return stmt.get(email) || null;
  }
  ,
  createRefreshToken: async ({ token, userId, expiresAt }) => {
    const stmt = db.prepare('INSERT INTO refresh_tokens (token,user_id,expires_at,created_at) VALUES (?, ?, ?, ?)');
    stmt.run(token, userId, Math.floor(expiresAt), new Date().toISOString());
    return { token, userId, expiresAt };
  },
  getRefreshToken: async (token) => {
    const stmt = db.prepare('SELECT token,user_id as userId,expires_at as expiresAt,revoked FROM refresh_tokens WHERE token = ?');
    const row = stmt.get(token);
    if (!row) return null;
    return { token: row.token, userId: row.userId, expiresAt: row.expiresAt, revoked: !!row.revoked };
  },
  revokeRefreshToken: async (token) => {
    const stmt = db.prepare('UPDATE refresh_tokens SET revoked = 1 WHERE token = ?');
    const info = stmt.run(token);
    return info.changes > 0;
  }
  ,
  revokeAllRefreshTokensForUser: async (userId) => {
    const stmt = db.prepare('UPDATE refresh_tokens SET revoked = 1 WHERE user_id = ?');
    const info = stmt.run(userId);
    return info.changes >= 0;
  }
};
