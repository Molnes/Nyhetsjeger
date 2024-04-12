BEGIN;

DROP TABLE IF EXISTS phone_number_verification;

DROP VIEW IF EXISTS user_quizzes;
DROP VIEW IF EXISTS user_question_points;

DROP TABLE IF EXISTS preassigned_roles;

DROP VIEW IF EXISTS available_usernames;

DROP TABLE IF EXISTS http_sessions;
DROP INDEX IF EXISTS http_sessions_expiry_idx;
DROP INDEX IF EXISTS http_sessions_key_idx;

DROP FUNCTION IF EXISTS calculate_points_awarded;

DROP TABLE IF EXISTS user_answers;

DROP TRIGGER IF EXISTS update_quiz_articles_trigger ON questions;
DROP FUNCTION IF EXISTS update_quiz_articles;

DROP TRIGGER IF EXISTS insert_quiz_articles_trigger ON questions;
DROP FUNCTION IF EXISTS insert_quiz_articles;

DROP TRIGGER IF EXISTS set_answer_alternative_arrangement ON answer_alternatives;
DROP FUNCTION IF EXISTS set_answer_alternative_arrangement;
DROP TABLE IF EXISTS answer_alternatives;

DROP TRIGGER IF EXISTS set_question_arrangement ON questions;
DROP FUNCTION IF EXISTS set_question_arrangement;
DROP TABLE IF EXISTS questions;

DROP TABLE IF EXISTS quiz_articles;

DROP TABLE IF EXISTS quizzes;

DROP TABLE IF EXISTS articles;

DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS nouns;
DROP TABLE IF EXISTS adjectives;

DROP TYPE IF EXISTS user_role;

END;