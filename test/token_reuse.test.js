const fs = require('fs');
const path = require('path');
process.env.DB_ADAPTER = 'sqlite';
process.env.SQLITE_FILE = path.join(__dirname, '..', 'data', 'test.db');
process.env.JWT_SECRET = 'test-secret';

const app = require('../src/server');

let server;

beforeAll(async () => {
  const p = path.join(__dirname, '..', 'data');
  if (!fs.existsSync(p)) fs.mkdirSync(p, { recursive: true });
  try { if (fs.existsSync(process.env.SQLITE_FILE)) fs.unlinkSync(process.env.SQLITE_FILE); } catch (e) {}
  server = app.listen(0);
});

afterAll(async () => {
  server && server.close();
  try { if (fs.existsSync(process.env.SQLITE_FILE)) fs.unlinkSync(process.env.SQLITE_FILE); } catch (e) {}
});

test('refresh token reuse -> revoke all tokens', async () => {
  const email = `reuse+${Date.now()}@example.com`;
  const pass = 'Password123!';
  const base = `http://127.0.0.1:${server.address().port}`;

  // register
  const regRes = await fetch(base + '/api/auth/register', {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email, password: pass })
  });
  expect(regRes.status).toBe(201);
  const reg = await regRes.json();

  // login -> get token A
  const loginRes = await fetch(base + '/api/auth/login', {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email, password: pass })
  });
  expect(loginRes.status).toBe(200);
  const login = await loginRes.json();
  const tokenA = login.refreshToken;

  // use A to rotate -> get token B
  const rotateRes = await fetch(base + '/api/auth/refresh', {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ refreshToken: tokenA })
  });
  expect(rotateRes.status).toBe(200);
  const rotate = await rotateRes.json();
  const tokenB = rotate.refreshToken;

  // reuse A -> should detect reuse and revoke all tokens
  const reuseRes = await fetch(base + '/api/auth/refresh', {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ refreshToken: tokenA })
  });
  expect(reuseRes.status).toBe(400);
  const reuseBody = await reuseRes.json();
  expect(reuseBody).toHaveProperty('error');
  expect(reuseBody.error.toLowerCase()).toMatch(/reuse/);

  // token B should now be revoked as well
  const afterRes = await fetch(base + '/api/auth/refresh', {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ refreshToken: tokenB })
  });
  expect(afterRes.status).toBe(400);
});
