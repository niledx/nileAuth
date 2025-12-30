# Nile Auth API (modular DB)

Lightweight Express auth API with modular database adapters to make migrations and adapter swaps easy.

Quick start

1. Navigate to the node directory:
```bash
cd node
```

2. Copy `.env.example` to `.env` and edit if needed.
3. Install dependencies:

```bash
npm install
```

4. Run migration (creates SQLite DB if using sqlite):

```bash
npm run migrate
```

5. Start server:

```bash
npm start
```

Running tests

Run the integration tests (they use a temporary SQLite file under `../data/test.db`):

```bash
npm test
```

Tests use `jest` and `supertest` and will clean up the temporary DB file after running.

Endpoints

- POST `/api/auth/register` { email, password } -> { user, token }
- POST `/api/auth/login` { email, password } -> { user, token }
 - POST `/api/auth/refresh` { refreshToken } -> { accessToken, refreshToken }
 - POST `/api/auth/logout` { refreshToken } -> { ok }

Switching DB adapters

Set `DB_ADAPTER=memory` (default) or `DB_ADAPTER=sqlite` in `.env`. When using sqlite, set `SQLITE_FILE` to the DB path.

Notes

- Passwords are hashed with `bcryptjs`.
- JWT secret comes from `JWT_SECRET` in `.env`.

Production recommendations

- Use a managed relational DB (Postgres) and implement adapter similar to `src/db/adapters/sqlite.js`.
- Postgres: a working adapter is provided at `src/db/adapters/postgres.js`. To use locally with Docker (from repo root):

```bash
# From repo root
docker compose up -d

# From node directory
cd node
export DB_ADAPTER=postgres
export POSTGRES_USER=nile
export POSTGRES_PASSWORD=nilepass
export POSTGRES_DB=nileauth
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
npm run migrate
npm start
```
To run Postgres migrations (uses `node-pg-migrate`):

```bash
# ensure Postgres env vars are set (or DATABASE_URL)
npm run migrate:pg

# to rollback last migration
npm run migrate:pg:down
```
- Keep `JWT_SECRET` in a secret store (Vault, AWS Secrets Manager) and rotate regularly.
- Use HTTPS and set `secure` cookies if storing tokens in cookies.
- Consider rotating refresh tokens and detect reuse to prevent token theft.
- Add monitoring (Prometheus) and structured logs (JSON) for observability.
nileAuth

Release workflow

Run a release which will update `CHANGELOG.md`, create a git tag, and bump package.json via `standard-version`:

```bash
# generate changelog, tag, and bump version
npm run release

# push tags and commits
git push --follow-tags origin main
```

Which of the remaining tasks would you like me to do next?
