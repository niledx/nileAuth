const users = new Map();
const tokens = new Map();
let idCounter = 1;

module.exports = {
  init: () => {},
  createUser: async ({ email, password }) => {
    const id = idCounter++;
    const user = { id, email, password, created_at: new Date().toISOString() };
    users.set(email, user);
    return user;
  },
  getUserByEmail: async (email) => {
    return users.get(email) || null;
  },
  createRefreshToken: async ({ token, userId, expiresAt }) => {
    tokens.set(token, { token, userId, expiresAt, revoked: false });
    return { token, userId, expiresAt };
  },
  getRefreshToken: async (token) => {
    const r = tokens.get(token);
    if (!r) return null;
    return { token: r.token, userId: r.userId, expiresAt: r.expiresAt, revoked: r.revoked };
  },
  revokeRefreshToken: async (token) => {
    const r = tokens.get(token);
    if (!r) return false;
    r.revoked = true;
    tokens.set(token, r);
    return true;
  }
  ,
  revokeAllRefreshTokensForUser: async (userId) => {
    for (const [k, v] of tokens.entries()) {
      if (v.userId === userId) {
        v.revoked = true;
        tokens.set(k, v);
      }
    }
    return true;
  }
};
