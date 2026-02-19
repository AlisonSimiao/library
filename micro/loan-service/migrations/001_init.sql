CREATE TABLE IF NOT EXISTS loans (
    id          SERIAL PRIMARY KEY,
    user_id     INT NOT NULL,
    book_id     INT NOT NULL,
    loaned_at   TIMESTAMP DEFAULT NOW(),
    returned_at TIMESTAMP,
    status      VARCHAR(20) DEFAULT 'active'
);
