const crypto = require('crypto');

// Token service: generates refresh tokens and validates rotation/revocation
// Tokens are opaque random strings stored in DB with expiry and user association

function makeToken() {
  return crypto.randomBytes(48).toString('hex');
}

module.exports = function TokenService(DB) {
  async function create(userId, ttlSeconds = 60 * 60 * 24 * 30) { // 30 days
    const token = makeToken();
    const expiresAt = Date.now() + ttlSeconds * 1000;
    const record = await DB.createRefreshToken({ token, userId, expiresAt });
    return { token: record.token, expiresAt: record.expiresAt };
  }

  async function get(token) {
    return DB.getRefreshToken(token);
  }

  async function revoke(token) {
    return DB.revokeRefreshToken(token);
  }

  async function rotate(oldToken, userId) {
    // revoke old and create new
    await revoke(oldToken);
    return create(userId);
  }

  return { create, get, revoke, rotate };
};
