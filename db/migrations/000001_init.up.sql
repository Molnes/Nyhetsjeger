BEGIN;

CREATE TABLE IF NOT EXISTS questions (
    id SERIAL PRIMARY KEY,
    question TEXT NOT NULL,
    article_id INTEGER NOT NULL
);




END;