const dbFactory = require('../db');
const bcrypt = require('bcryptjs');
const jwt = require('jsonwebtoken');
const TokenServiceFactory = require('./token');

const DB = dbFactory();
const TokenService = TokenServiceFactory(DB);
const JWT_SECRET = process.env.JWT_SECRET || 'change-me';
const ACCESS_EXPIRES = process.env.ACCESS_EXPIRES || '1h';

async function registerUser({ email, password }) {
  if (!email || !password) throw new Error('email and password required');
  const existing = await DB.getUserByEmail(email);
  if (existing) throw new Error('user already exists');
  const hash = await bcrypt.hash(password, 12);
  const user = await DB.createUser({ email, password: hash });
  const accessToken = jwt.sign({ userId: user.id }, JWT_SECRET, { expiresIn: ACCESS_EXPIRES });
  const refresh = await TokenService.create(user.id);
  return { user: { id: user.id, email: user.email }, accessToken, refreshToken: refresh.token };
}

async function loginUser({ email, password }) {
  if (!email || !password) throw new Error('email and password required');
  const user = await DB.getUserByEmail(email);
  if (!user) throw new Error('invalid credentials');
  const ok = await bcrypt.compare(password, user.password);
  if (!ok) throw new Error('invalid credentials');
  const accessToken = jwt.sign({ userId: user.id }, JWT_SECRET, { expiresIn: ACCESS_EXPIRES });
  const refresh = await TokenService.create(user.id);
  return { accessToken, refreshToken: refresh.token, user: { id: user.id, email: user.email } };
}

async function refreshAccessToken({ refreshToken }) {
  if (!refreshToken) throw new Error('refresh token required');
  const row = await TokenService.get(refreshToken);
  if (!row) throw new Error('invalid refresh token');
  if (row.revoked) {
    // possible token reuse detected: revoke all tokens for this user
    if (DB && typeof DB.revokeAllRefreshTokensForUser === 'function') {
      await DB.revokeAllRefreshTokensForUser(row.userId);
    }
    throw new Error('token reuse detected');
  }
  if (row.expiresAt < Date.now()) throw new Error('token expired');
  // rotate
  const newRefresh = await TokenService.rotate(refreshToken, row.userId);
  const accessToken = jwt.sign({ userId: row.userId }, JWT_SECRET, { expiresIn: ACCESS_EXPIRES });
  return { accessToken, refreshToken: newRefresh.token };
}

async function revokeRefreshToken({ refreshToken }) {
  if (!refreshToken) throw new Error('refresh token required');
  await TokenService.revoke(refreshToken);
  return true;
}

module.exports = { registerUser, loginUser, refreshAccessToken, revokeRefreshToken };
