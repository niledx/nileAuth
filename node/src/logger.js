const pino = require('pino');
const pinoHttp = require('pino-http');

const level = process.env.LOG_LEVEL || (process.env.NODE_ENV === 'production' ? 'info' : 'debug');

let logger;
// file logging when enabled
if (process.env.LOG_TO_FILE === 'true') {
  const logPath = process.env.LOG_FILE_PATH || './logs/app.log';
  const dest = pino.destination({ dest: logPath, sync: false });
  logger = pino({ level, base: { pid: false } }, dest);
} else {
  logger = pino({ level, base: { pid: false } });
}

// optional Sentry integration
if (process.env.SENTRY_DSN) {
  try {
    const Sentry = require('@sentry/node');
    Sentry.init({ dsn: process.env.SENTRY_DSN });
    logger.info({ dsn: true }, 'Sentry initialized');
    // attach Sentry capture on error
    const origError = logger.error.bind(logger);
    logger.error = function () {
      try { Sentry.captureException(arguments[0] instanceof Error ? arguments[0] : new Error(JSON.stringify(arguments))); } catch (e) {}
      return origError.apply(null, arguments);
    };
  } catch (e) {
    logger.warn({ err: e }, 'Sentry integration failed (optional dependency)');
  }
}

module.exports = {
  logger,
  http: pinoHttp({ logger })
};
