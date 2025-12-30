require('dotenv').config();
const express = require('express');
const authRoutes = require('./routes/auth');
const helmet = require('helmet');
const cors = require('cors');
const rateLimit = require('express-rate-limit');
const { http: pinoHttp, logger } = require('./logger');

const limiter = rateLimit({
	windowMs: 15 * 60 * 1000,
	max: 200
});

// security: ensure JWT secret is set in production
if (process.env.NODE_ENV === 'production' && (!process.env.JWT_SECRET || process.env.JWT_SECRET === 'change-me')) {
	const { logger } = require('./logger');
	logger.fatal('FATAL: JWT_SECRET must be set in production');
	process.exit(1);
}

const app = express();
app.use(helmet());
app.use(cors());
app.use(pinoHttp);
app.use(express.json());
app.use(limiter);

app.use('/api/auth', authRoutes);

app.get('/health', (req, res) => res.json({ok: true}));

// central error handler
app.use((err, req, res, next) => {
	logger.error({ err }, 'unhandled error');
	res.status(err.status || 500).json({ error: err.message || 'internal error' });
});

if (require.main === module) {
	const port = process.env.PORT || 3000;
	app.listen(port, () => {
		const { logger } = require('./logger');
		logger.info({ port }, `Server listening on ${port}`);
	});
}

module.exports = app;
