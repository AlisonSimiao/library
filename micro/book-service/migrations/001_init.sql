CREATE TABLE IF NOT EXISTS books (
    id         SERIAL PRIMARY KEY,
    title      VARCHAR(255) NOT NULL,
    author     VARCHAR(255) NOT NULL,
    isbn       VARCHAR(20) UNIQUE,
    quantity   INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW()
);
