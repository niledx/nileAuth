const express = require('express');
const router = express.Router();
const { registerUser, loginUser, refreshAccessToken, revokeRefreshToken } = require('../services/auth');
const { body, validationResult } = require('express-validator');

router.post('/register',
  body('email').isEmail(),
  body('password').isLength({ min: 8 }),
  async (req, res) => {
    const errors = validationResult(req);
    if (!errors.isEmpty()) return res.status(400).json({ errors: errors.array() });
    try {
      const { email, password } = req.body;
      const result = await registerUser({ email, password });
      res.status(201).json(result);
    } catch (err) {
      res.status(400).json({ error: err.message });
    }
  }
);

router.post('/login',
  body('email').isEmail(),
  body('password').exists(),
  async (req, res) => {
    const errors = validationResult(req);
    if (!errors.isEmpty()) return res.status(400).json({ errors: errors.array() });
    try {
      const { email, password } = req.body;
      const result = await loginUser({ email, password });
      res.json(result);
    } catch (err) {
      res.status(400).json({ error: err.message });
    }
  }
);

router.post('/refresh',
  body('refreshToken').exists(),
  async (req, res) => {
    const errors = validationResult(req);
    if (!errors.isEmpty()) return res.status(400).json({ errors: errors.array() });
    try {
      const { refreshToken } = req.body;
      const result = await refreshAccessToken({ refreshToken });
      res.json(result);
    } catch (err) {
      res.status(400).json({ error: err.message });
    }
  }
);

router.post('/logout',
  body('refreshToken').exists(),
  async (req, res) => {
    const errors = validationResult(req);
    if (!errors.isEmpty()) return res.status(400).json({ errors: errors.array() });
    try {
      const { refreshToken } = req.body;
      await revokeRefreshToken({ refreshToken });
      res.json({ ok: true });
    } catch (err) {
      res.status(400).json({ error: err.message });
    }
  }
);

module.exports = router;
