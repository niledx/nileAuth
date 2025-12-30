const fs = require('fs');
const path = require('path');
process.env.DB_ADAPTER = 'sqlite';
process.env.SQLITE_FILE = path.join(__dirname, '..', 'data', 'test.db');
process.env.JWT_SECRET = 'test-secret';

let server;

beforeAll(async () => {
  const p = path.join(__dirname, '..', 'data');
  if (!fs.existsSync(p)) fs.mkdirSync(p, { recursive: true });
  try {
    if (fs.existsSync(process.env.SQLITE_FILE)) fs.unlinkSync(process.env.SQLITE_FILE);
  } catch (e) {
    // ignore busy/locked file; fresh DB will be used
  }
  const app = require('../src/server');
  server = app.listen(0);
});

afterAll(async () => {
  server && server.close();
  try {
    if (fs.existsSync(process.env.SQLITE_FILE)) fs.unlinkSync(process.env.SQLITE_FILE);
  } catch (e) {}
});

test('register -> login -> refresh', async () => {
  const email = `test+${Date.now()}@example.com`;
  const pass = 'Password123!';
  const base = `http://127.0.0.1:${server.address().port}`;

  const regRes = await fetch(base + '/api/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password: pass })
  });
  expect(regRes.status).toBe(201);
  const reg = await regRes.json();
  expect(reg).toHaveProperty('accessToken');
  expect(reg).toHaveProperty('refreshToken');

  const loginRes = await fetch(base + '/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password: pass })
  });
  expect(loginRes.status).toBe(200);
  const login = await loginRes.json();
  expect(login).toHaveProperty('accessToken');
  expect(login).toHaveProperty('refreshToken');

  const refreshRes = await fetch(base + '/api/auth/refresh', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refreshToken: login.refreshToken })
  });
  expect(refreshRes.status).toBe(200);
  const refresh = await refreshRes.json();
  expect(refresh).toHaveProperty('accessToken');
  expect(refresh).toHaveProperty('refreshToken');
});
