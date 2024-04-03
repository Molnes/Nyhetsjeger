BEGIN;

CREATE TEMP TABLE whitelist_words_temp
(
    adjective text, noun text
) ON COMMIT DROP;

COPY whitelist_words_temp(adjective, noun)
FROM '/tmp/whitelist-words.csv'
DELIMITER ';'
CSV HEADER;

INSERT INTO adjectives
SELECT adjective
FROM whitelist_words_temp
WHERE adjective IS NOT NULL
ON CONFLICT DO NOTHING;

INSERT INTO nouns
SELECT noun
FROM whitelist_words_temp
WHERE noun IS NOT NULL
ON CONFLICT DO NOTHING;

COMMIT;