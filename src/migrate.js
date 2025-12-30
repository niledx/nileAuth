require('dotenv').config();
const dbFactory = require('./db');
const { logger } = require('./logger');

logger.info({ adapter: process.env.DB_ADAPTER || 'memory' }, 'Running migrate/init');
const DB = dbFactory();
if (DB && typeof DB.init === 'function') {
  try {
    DB.init({ dbPath: process.env.SQLITE_FILE });
  } catch (err) {
    logger.error({ err }, 'DB init failed');
  }
}
logger.info({ adapter: process.env.DB_ADAPTER || 'memory' }, 'Migration complete');
