CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS books (
    id         SERIAL PRIMARY KEY,
    title      VARCHAR(255) NOT NULL,
    author     VARCHAR(255) NOT NULL,
    isbn       VARCHAR(20) UNIQUE,
    quantity   INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS loans (
    id          SERIAL PRIMARY KEY,
    user_id     INT NOT NULL REFERENCES users(id),
    book_id     INT NOT NULL REFERENCES books(id),
    loaned_at   TIMESTAMP DEFAULT NOW(),
    returned_at TIMESTAMP,
    status      VARCHAR(20) DEFAULT 'active'
);
