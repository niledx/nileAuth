const path = require('path');

const adapterName = process.env.DB_ADAPTER || 'memory';

function loadAdapter(name) {
  if (name === 'sqlite') return require('./adapters/sqlite');
  if (name === 'postgres' || name === 'pg') return require('./adapters/postgres');
  return require('./adapters/memory');
}

module.exports = function () {
  const adapter = loadAdapter(adapterName);
  if (adapter && typeof adapter.init === 'function') {
    if (adapterName === 'sqlite') {
      const dbPath = process.env.SQLITE_FILE || path.join(__dirname, '..', 'data', 'nile.db');
      adapter.init({ dbPath });
    } else if (adapterName === 'postgres' || adapterName === 'pg') {
      try {
        const connectionString = process.env.POSTGRES_URL;
        adapter.init({ connectionString });
      } catch (err) {
        const { logger } = require('../logger');
        logger.error({ err }, 'Postgres adapter init error');
      }
    } else {
      adapter.init();
    }
  }
  return adapter;
};
