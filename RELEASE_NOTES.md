# Release Notes - Nile Auth API

Version: 0.1.0 - initial production-ready scaffold
Date: 2025-12-31

Highlights:
- Modular auth API with pluggable DB adapters (memory, SQLite, Postgres).
- Secure defaults: `helmet`, `cors`, request logging, rate limiting.
- Access tokens (JWT) + opaque refresh tokens with rotation and revocation.
- Validation via `express-validator` and hashed passwords with `bcryptjs`.
- Migration/init script and Docker + docker-compose for local Postgres.
- CI workflow and integration tests covering register/login/refresh flows.

Files of interest:
- `src/server.js` - Express app (exports `app` for testing)
- `src/services/auth.js` - Auth logic (access + refresh tokens)
- `src/services/token.js` - Refresh token generation & rotation
- `src/db/adapters` - `memory.js`, `sqlite.js`, `postgres.js`
- `test/auth.test.js` - Integration test (register -> login -> refresh)
- `.github/workflows/ci.yml` - CI including `npm test`

Upgrade notes:
- Move to a managed secrets store for `JWT_SECRET`.
- Replace ad-hoc migrations with a migration tool for Postgres (e.g. `node-pg-migrate`).

