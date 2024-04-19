package usernames

import (
	"context"
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

const (
	nounTable      = "noun-table"
	adjectiveTable = "adjective-table"
)

// setUaiAdjInfo sets the adjectives and relevant information for the UsernameAdminInfo struct.
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

// setUaiNounInfo sets the nouns and relevant information for the UsernameAdminInfo struct.
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
	if tableId == nounTable {
		_, err := db.Exec(`INSERT INTO nouns VALUES ($1);`, word)
		return err
	} else if tableId == adjectiveTable {
		_, err := db.Exec(`INSERT INTO adjectives VALUES ($1);`, word)
		return err
	} else {
		//Return error
		return errors.New("table ID not recognized")
	}
}

// DeleteWordsFromTable deletes the words from the adjectives and nouns tables and updates the usernames in the users table.
func DeleteWordsFromTable(db *sql.DB, ctx context.Context, words []string) error {

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
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
						WHERE users.id = ( 
							SELECT id
							FROM users
							WHERE username_adjective = ANY($1)
							OR
							username_noun = ANY($1)
						);
					`, pq.Array(words))
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM adjectives WHERE adjective = ANY($1);`, pq.Array(words))
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM nouns WHERE noun = ANY($1);`, pq.Array(words))
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// UpdateAdjectives updates the adjectives in the adjectives table.
// The oldNewAdjArr is an 2 dimensional array where each inner array contains two strings: the old adjective and the new adjective.
//
// For example:
//  oldNewAdjArr := [][]string{{"oldAdj1", "newAdj1"}, {"oldAdj2", "newAdj2"}}
func UpdateAdjectives(db *sql.DB, oldNewAdjArr [][]string) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	pq.Array(oldNewAdjArr)

	for _, oldNewAdj := range oldNewAdjArr {
		_, err = tx.Exec(`UPDATE adjectives SET adjective = $1 WHERE adjective = $2;`, oldNewAdj[1], oldNewAdj[0])
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil

}

// UpdateNouns updates the nouns in the nouns table.
// The oldNewNounArr is an 2 dimensional array where each inner array contains two strings: the old noun and the new noun.
//
// For example:
//  oldNewNounArr := [][]string{{"oldNoun1", "newNoun1"}, {"oldNoun2", "newNoun2"}}
func UpdateNouns(db *sql.DB, oldNewNounArr [][]string) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, oldNewNoun := range oldNewNounArr {
		_, err = tx.Exec(`UPDATE nouns SET noun = $1 WHERE noun = $2;`, oldNewNoun[1], oldNewNoun[0])
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil

}
