BEGIN;

CREATE TYPE user_role AS ENUM ('user', 'quiz_admin', 'organization_admin');

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sso_user_id TEXT NOT NULL, -- The user id from the SSO provider
    username TEXT,
    email TEXT NOT NULL,
    phone TEXT,
    opt_in_ranking BOOLEAN NOT NULL,
    role user_role NOT NULL DEFAULT 'user',
    access_token TEXT NOT NULL,
    token_expires_at TIMESTAMP,
    refresh_token TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    image_url TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS quizzes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    image_url TEXT NOT NULL,
    available_from TIMESTAMP NOT NULL,
    available_to TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    last_modified_at TIMESTAMP NOT NULL DEFAULT now(),
    published BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS quiz_articles (
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    PRIMARY KEY (quiz_id, article_id)
);


CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question TEXT NOT NULL,
    arrangement INTEGER NOT NULL,
    article_id UUID REFERENCES articles(id) ON DELETE CASCADE,
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    points INTEGER NOT NULL,
    CONSTRAINT question_arrangement UNIQUE (arrangement, quiz_id)
);

CREATE OR REPLACE FUNCTION set_question_arrangement()
RETURNS TRIGGER AS $$
DECLARE
    max_arrangement INTEGER;
BEGIN
    SELECT MAX(arrangement) INTO max_arrangement
    FROM questions
    WHERE quiz_id = NEW.quiz_id;

    IF max_arrangement IS NULL THEN
        max_arrangement := 0;
    END IF;
    NEW.arrangement := max_arrangement + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER set_question_arrangement
    BEFORE INSERT ON questions
    FOR EACH ROW
    EXECUTE FUNCTION set_question_arrangement();

CREATE TABLE IF NOT EXISTS answer_alternatives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    text TEXT NOT NULL,
    correct BOOLEAN NOT NULL,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    CONSTRAINT unique_alternative_question UNIQUE (id, question_id)
);

CREATE OR REPLACE FUNCTION insert_quiz_articles()
RETURNS TRIGGER AS $$
BEGIN
    -- If the article_id is not null, insert a new entry
    IF NEW.article_id IS NOT NULL THEN
        INSERT INTO quiz_articles (quiz_id, article_id) VALUES (NEW.quiz_id, NEW.article_id);
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER insert_quiz_articles_trigger
    AFTER INSERT ON questions
    FOR EACH ROW
    EXECUTE PROCEDURE insert_quiz_articles();

CREATE OR REPLACE FUNCTION update_quiz_articles()
RETURNS TRIGGER AS $$
BEGIN
    -- If the article_id is updated
    IF NEW.article_id IS DISTINCT FROM OLD.article_id THEN
        -- If the old article_id is not null, delete the old entry
        IF OLD.article_id IS NOT NULL THEN
            DELETE FROM quiz_articles WHERE article_id = OLD.article_id AND quiz_id = OLD.quiz_id;
        END IF;

        -- If the new article_id is not null, insert a new entry
        -- And the new article_id and quiz_id combination already exists, do nothing
        IF NEW.article_id IS NOT NULL THEN
            INSERT INTO quiz_articles (quiz_id, article_id) VALUES (NEW.quiz_id, NEW.article_id)
            ON CONFLICT DO NOTHING;
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_quiz_articles_trigger
    AFTER UPDATE OF article_id ON questions
    FOR EACH ROW
    EXECUTE PROCEDURE update_quiz_articles();

CREATE TABLE IF NOT EXISTS user_answers (
    user_id UUID NOT NULL  REFERENCES users(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    question_presented_at TIMESTAMP NOT NULL DEFAULT now(),
     -- following columns are nullable; if they are null, the user has not answered the question yet
    chosen_answer_alternative_id UUID REFERENCES answer_alternatives(id) ON DELETE CASCADE,
    answered_at TIMESTAMP,
    points_awarded INTEGER CHECK (points_awarded >= 0),
    PRIMARY KEY (user_id, question_id),
    CONSTRAINT ans_alt_belong_to_question
        FOREIGN KEY (question_id, chosen_answer_alternative_id)
        REFERENCES answer_alternatives(question_id, id)
);

-- Table expected by package we use for sessions
-- code taken directly from https://github.com/antonlindstrom/pgstore/blob/e3a6e3fed12a32697b352a4636d78204f9dbdc81/pgstore.go#L234
CREATE TABLE IF NOT EXISTS http_sessions (
              id BIGSERIAL PRIMARY KEY,
              key BYTEA,
              data BYTEA,
              created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
              modified_on TIMESTAMPTZ,
              expires_on TIMESTAMPTZ);
              CREATE INDEX IF NOT EXISTS http_sessions_expiry_idx ON http_sessions (expires_on);
              CREATE INDEX IF NOT EXISTS http_sessions_key_idx ON http_sessions (key);
END;