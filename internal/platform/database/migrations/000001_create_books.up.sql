CREATE TABLE IF NOT EXISTS books (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    author      VARCHAR(255) NOT NULL,
    year        INTEGER      NOT NULL,
    masterpiece BOOLEAN      NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_books_author      ON books (author);
CREATE INDEX IF NOT EXISTS idx_books_year        ON books (year);
CREATE INDEX IF NOT EXISTS idx_books_masterpiece ON books (masterpiece);
