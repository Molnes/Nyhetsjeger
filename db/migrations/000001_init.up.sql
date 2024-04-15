BEGIN;

CREATE TYPE user_role AS ENUM ('user', 'quiz_admin', 'organization_admin');

CREATE TABLE IF NOT EXISTS adjectives (
    adjective TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS nouns (
    noun TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sso_user_id TEXT NOT NULL, -- The user id from the SSO provider
    username_adjective TEXT NOT NULL REFERENCES adjectives(adjective) ON UPDATE CASCADE,
    username_noun TEXT NOT NULL REFERENCES nouns(noun) ON UPDATE CASCADE,
    email TEXT NOT NULL UNIQUE,
    phone TEXT,
    opt_in_ranking BOOLEAN NOT NULL,
    accepted_terms BOOL NOT NULL DEFAULT false,
    role user_role NOT NULL DEFAULT 'user',
    access_token TEXT NOT NULL,
    token_expires_at TIMESTAMP,
    refresh_token TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    image_url TEXT
);

CREATE TABLE IF NOT EXISTS quizzes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    image_url TEXT,
    active_from TIMESTAMP NOT NULL,
    active_to TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    last_modified_at TIMESTAMP NOT NULL DEFAULT now(),
    published BOOLEAN NOT NULL DEFAULT FALSE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS quiz_articles (
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    PRIMARY KEY (quiz_id, article_id)
);


CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question TEXT NOT NULL,
    image_url TEXT,
    arrangement INTEGER NOT NULL,
    article_id UUID REFERENCES articles(id) ON DELETE CASCADE,
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    time_limit_seconds INTEGER NOT NULL DEFAULT 30 CHECK (time_limit_seconds > 0),
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
    arrangement INTEGER NOT NULL,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    CONSTRAINT unique_alternative_question UNIQUE (id, question_id),
    CONSTRAINT unique_arrangement_question UNIQUE (arrangement, question_id)
);

-- Trigger to set the arrangement of answer alternatives
CREATE OR REPLACE FUNCTION set_answer_alternative_arrangement()
RETURNS TRIGGER AS $$
DECLARE
    max_arrangement INTEGER;
BEGIN
    SELECT MAX(arrangement) INTO max_arrangement
    FROM answer_alternatives
    WHERE question_id = NEW.question_id;

    IF max_arrangement IS NULL THEN
        max_arrangement := 0;
    END IF;
    NEW.arrangement := max_arrangement + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER set_answer_alternative_arrangement
    BEFORE INSERT ON answer_alternatives
    FOR EACH ROW
    EXECUTE FUNCTION set_answer_alternative_arrangement();


CREATE OR REPLACE FUNCTION insert_quiz_articles()
RETURNS TRIGGER AS $$
BEGIN
    -- If the article_id is not null and the quiz_article combo does not already exist, insert a new entry
    IF NEW.article_id IS NOT NULL AND NOT EXISTS (
        SELECT 1
        FROM quiz_articles
        WHERE quiz_id = NEW.quiz_id AND article_id = NEW.article_id
    )
    THEN
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

-- calculates points awarded for a question based on the time spent on the question, question's max points and duration/time limit
CREATE OR REPLACE FUNCTION calculate_points_awarded(
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    duration_seconds INTEGER,
    max_points INTEGER
    )
RETURNS INTEGER AS $$
DECLARE
    awarded_points INTEGER;
    time_spent float8;
    min_points INTEGER;
BEGIN
    time_spent := EXTRACT(EPOCH FROM end_time - start_time);
    min_points := max_points / 5;

    -- y= -ax + b, where a = (max_points - min_points) / duration_seconds, b = max_points, x = time_spent
    awarded_points := -1.0 * ( float8((max_points - min_points)) / duration_seconds) * time_spent + max_points;

    IF awarded_points < min_points THEN
        awarded_points := min_points;
    END IF;

    RETURN awarded_points;
END;
$$ LANGUAGE plpgsql;


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


CREATE VIEW available_usernames AS
    SELECT a.adjective, n.noun
    FROM adjectives a
    CROSS JOIN nouns n
    WHERE NOT EXISTS (
        SELECT 1
        FROM users u
        WHERE u.username_adjective = a.adjective AND u.username_noun = n.noun
    );

-- Emails of users who should be granted a role when they sign up
CREATE TABLE IF NOT EXISTS preassigned_roles(
    email TEXT NOT NULL PRIMARY KEY,
    role user_role NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
    CONSTRAINT not_user_role CHECK (role != 'user')
);

-- View with all questions user has answered, points awarded and when the question was answered
CREATE OR REPLACE VIEW user_question_points AS
SELECT
ua.user_id,
ua.question_id,
q.quiz_id,
ua.answered_at AS answered_at,
ua.chosen_answer_alternative_id, 
CASE
    WHEN aa.correct THEN
        calculate_points_awarded(ua.question_presented_at, ua.answered_at, q.time_limit_seconds, q.points)
    ELSE
        0
END AS points_awarded

FROM user_answers ua
JOIN questions q ON ua.question_id = q.id
JOIN answer_alternatives aa ON ua.chosen_answer_alternative_id = aa.id
WHERE ua.answered_at IS NOT NULL;


-- View to get all quizzes user has played, total points, last answer time, is_completed and answered_within_active_time
CREATE OR REPLACE VIEW user_quizzes AS
SELECT
uqp.user_id,
uqp.quiz_id,
SUM(uqp.points_awarded) AS total_points_awarded,
ic.is_completed,
MAX(uqp.answered_at) AS finished_at,
MAX(uqp.answered_at) < qz.active_to AS answered_within_active_time
FROM user_question_points uqp
JOIN questions q ON uqp.question_id = q.id
JOIN quizzes qz ON q.quiz_id = qz.id,
LATERAL (
    -- check if there are questions user has not answered in the quiz
    SELECT COUNT(q.id) = 0 as is_completed
		FROM questions q
		WHERE quiz_id = uqp.quiz_id 
		AND id NOT IN (
			SELECT question_id
			FROM user_answers
			WHERE chosen_answer_alternative_id IS NOT NULL
			AND user_id = uqp.user_id
		)
) ic
GROUP BY uqp.user_id, uqp.quiz_id, ic.is_completed, qz.active_to;



END;