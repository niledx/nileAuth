exports.shorthands = undefined;

exports.up = (pgm) => {
  pgm.createTable('users', {
    id: { type: 'serial', primaryKey: true },
    email: { type: 'text', notNull: true, unique: true },
    password: { type: 'text', notNull: true },
    created_at: { type: 'timestamp with time zone', notNull: true, default: pgm.func('current_timestamp') }
  });

  pgm.createTable('refresh_tokens', {
    token: { type: 'text', primaryKey: true },
    user_id: { type: 'integer', notNull: true },
    expires_at: { type: 'bigint' },
    revoked: { type: 'boolean', notNull: true, default: false },
    created_at: { type: 'timestamp with time zone', notNull: true, default: pgm.func('current_timestamp') }
  });

  pgm.addConstraint('refresh_tokens', 'fk_refresh_user', 'FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE');
  pgm.createIndex('refresh_tokens', 'user_id');
};

exports.down = (pgm) => {
  pgm.dropTable('refresh_tokens');
  pgm.dropTable('users');
};
