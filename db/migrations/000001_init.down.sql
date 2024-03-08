BEGIN;

DROP TABLE IF EXISTS http_sessions;
DROP INDEX IF EXISTS http_sessions_expiry_idx;
DROP INDEX IF EXISTS http_sessions_key_idx;

DROP TABLE IF EXISTS user_answers;

DROP TABLE IF EXISTS answer_alternatives;

DROP TRIGGER IF EXISTS set_question_arrangement ON questions;

DROP FUNCTION IF EXISTS set_question_arrangement;

DROP TABLE IF EXISTS questions;

DROP TABLE IF EXISTS quizzes;

DROP TABLE IF EXISTS articles;

DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS user_role;

END;