package usernames

import (
	"database/sql"
	"errors"
	"github.com/lib/pq"
)

type UsernameAdminInfo struct {
	Adjectives       []string
	Nouns            []string
	AdjPage          int
	NounPage         int
	AdjWordCount     int
	NounWordCount    int
	UsernamesPerPage int
}

func (uai *UsernameAdminInfo) setUaiAdjInfo(db *sql.DB) error {

	err := db.QueryRow(`
	WITH offsetvalue AS (
		SELECT (
				CASE 
					WHEN $2 < 1 THEN 1
					WHEN $2 > (SELECT COUNT(*) FROM adjectives) / $1
						THEN CEIL((SELECT COUNT(*) FROM adjectives)::float8 / $1::float8)
					ELSE $2
				END
			) as value
		)
		
		SELECT ARRAY (
		SELECT adjective
		FROM adjectives
		ORDER BY adjective ASC
		LIMIT $1 OFFSET $1 * (offsetvalue.value - 1)
		) as foo, COUNT(a.*), offsetvalue.value
		FROM adjectives a, offsetvalue
		GROUP BY offsetvalue.value;`,
		uai.UsernamesPerPage, uai.AdjPage).Scan(pq.Array(&uai.Adjectives), &uai.AdjWordCount, &uai.AdjPage)

	return err
}

func (uai *UsernameAdminInfo) setUaiNounInfo(db *sql.DB) error {
	err := db.QueryRow(
		`WITH offsetvalue AS (
			SELECT (
					CASE 
						WHEN $2 < 1 THEN 1
						WHEN $2 > (SELECT COUNT(*) FROM nouns) / $1
							THEN CEIL((SELECT COUNT(*) FROM nouns)::float8 / $1::float8)
						ELSE $2
					END
				) as value
			)
			
			SELECT ARRAY (
			SELECT noun
			FROM nouns
			ORDER BY noun ASC
			LIMIT $1 OFFSET $1 * (offsetvalue.value - 1)
			) as foo, COUNT(a.*), offsetvalue.value
			FROM nouns a, offsetvalue
			GROUP BY offsetvalue.value;`,
		uai.UsernamesPerPage, uai.NounPage).Scan(pq.Array(&uai.Nouns), &uai.NounWordCount, &uai.NounPage)

	return err
}

// GetUsernameAdminInfo returns a UsernameAdminInfo struct containing the adjectives and nouns and relevant information
// for rendering the username administration page.
func GetUsernameAdminInfo(db *sql.DB, adjPage int, nounPage int, usernamesPerPage int) (*UsernameAdminInfo, error) {
	uai := UsernameAdminInfo{}

	uai.AdjPage = adjPage
	uai.NounPage = nounPage
	uai.UsernamesPerPage = usernamesPerPage

	err := uai.setUaiAdjInfo(db)

	if err != nil {
		return nil, err
	}

	err = uai.setUaiNounInfo(db)

	if err != nil {
		return nil, err
	}

	return &uai, nil
}

// AddWordToTable adds a word to the specified table.
func AddWordToTable(db *sql.DB, word string, tableId string) error {
	if tableId == "noun-table" {
		_, err := db.Exec(`INSERT INTO nouns VALUES ($1);`, word)
		return err
	} else if tableId == "adjective-table" {
		_, err := db.Exec(`INSERT INTO adjectives VALUES ($1);`, word)
		return err
	} else {
		//Return error
		return errors.New("table ID not recognized")
	}
}

func DeleteWordsFromTable(db *sql.DB, words []string) error {

	_, err := db.Exec(`

					UPDATE users
					SET
						username_adjective = random_username.adjective,
						username_noun = random_username.noun
					FROM (
						SELECT adjective, noun
						FROM available_usernames 
						OFFSET floor(random() * (
						SELECT COUNT(*) FROM available_usernames)
						) 
					LIMIT 1) AS random_username
						WHERE users.id = ( SELECT id
						FROM users
						WHERE username_adjective = 'brilleslange'
						OR
						username_noun = 'brilleslange' 
					);

					DELETE FROM adjectives WHERE adjectives = ANY($1);
					DELETE FROM nouns WHERE nouns = ANY($1); 
						`, pq.Array(words))
	return err
}
