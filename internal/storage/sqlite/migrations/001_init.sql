CREATE TABLE IF NOT EXISTS links (
                                     id          INTEGER PRIMARY KEY AUTOINCREMENT,
                                     code        TEXT NOT NULL UNIQUE,
                                     url         TEXT NOT NULL,
                                     created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                     expires_at  TIMESTAMP NULL,
                                     hit_count   INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_links_code ON links(code);