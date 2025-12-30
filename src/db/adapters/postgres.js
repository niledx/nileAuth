const { Pool } = require('pg');
let pool;

module.exports = {
  init: ({ connectionString } = {}) => {
    const cs = connectionString || process.env.POSTGRES_URL || `postgresql://${process.env.POSTGRES_USER || 'postgres'}:${process.env.POSTGRES_PASSWORD || 'postgres'}@${process.env.POSTGRES_HOST || 'localhost'}:${process.env.POSTGRES_PORT || 5432}/${process.env.POSTGRES_DB || 'nileauth'}`;
    pool = new Pool({ connectionString: cs });
    // create tables if not exist - fire-and-forget so init is synchronous
    pool.query(`
      CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        email TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
      );
    `).catch(err => {
      try { require('../../logger').logger.error({ err }, 'pg init users error'); } catch (e) { require('pino')().error('pg init users error', err); }
    });
    pool.query(`
      CREATE TABLE IF NOT EXISTS refresh_tokens (
        token TEXT PRIMARY KEY,
        user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
        expires_at BIGINT,
        revoked BOOLEAN DEFAULT false,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
      );
    `).catch(err => {
      try { require('../../logger').logger.error({ err }, 'pg init refresh_tokens error'); } catch (e) { require('pino')().error('pg init refresh_tokens error', err); }
    });
  },
  createUser: async ({ email, password }) => {
    const res = await pool.query('INSERT INTO users(email,password) VALUES($1,$2) RETURNING id,email', [email, password]);
    return { id: res.rows[0].id, email: res.rows[0].email, password };
  },
  getUserByEmail: async (email) => {
    const res = await pool.query('SELECT id,email,password,created_at FROM users WHERE email = $1', [email]);
    return res.rows[0] || null;
  },
  createRefreshToken: async ({ token, userId, expiresAt }) => {
    await pool.query('INSERT INTO refresh_tokens(token,user_id,expires_at) VALUES($1,$2,$3)', [token, userId, Math.floor(expiresAt)]);
    return { token, userId, expiresAt };
  },
  getRefreshToken: async (token) => {
    const res = await pool.query('SELECT token,user_id AS "userId",expires_at AS "expiresAt",revoked FROM refresh_tokens WHERE token = $1', [token]);
    if (!res.rows[0]) return null;
    return { token: res.rows[0].token, userId: res.rows[0].userId, expiresAt: res.rows[0].expiresAt, revoked: res.rows[0].revoked };
  },
  revokeRefreshToken: async (token) => {
    const res = await pool.query('UPDATE refresh_tokens SET revoked = true WHERE token = $1', [token]);
    return res.rowCount > 0;
  }
  ,
  revokeAllRefreshTokensForUser: async (userId) => {
    const res = await pool.query('UPDATE refresh_tokens SET revoked = true WHERE user_id = $1', [userId]);
    return res.rowCount >= 0;
  }
};
