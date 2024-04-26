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

type OldNew struct {
	Old string
	New string
}

const (
	NounTable      = "noun-table"
	AdjectiveTable = "adjective-table"
)

// setUaiAdjInfo sets the adjectives and relevant information for the UsernameAdminInfo struct.
func (uai *UsernameAdminInfo) setUaiAdjInfo(db *sql.DB, searchParam string) error {

	err := db.QueryRow(`
	WITH offsetvalue AS (
		SELECT (
			CASE 
				WHEN $2 < 1 
					THEN 1
				WHEN ($3 != '') IS TRUE AND $2 > (SELECT COUNT(*) FROM adjectives WHERE adjective SIMILAR TO '%' || $3 || '%') / $1
					THEN CEIL((SELECT COUNT(*) FROM adjectives WHERE adjective SIMILAR TO '%' || $3 || '%')::float8 / $1::float8)
				WHEN $2 > (SELECT COUNT(*) FROM adjectives) / $1
					THEN CEIL((SELECT COUNT(*) FROM adjectives)::float8 / $1::float8)
				ELSE $2
			END
		) as value
	)
					
	SELECT ARRAY (
		SELECT adjective
			FROM adjectives
			WHERE adjective SIMILAR TO '%' || $3 ||'%'
			ORDER BY adjective ASC
			LIMIT $1 OFFSET $1* (offsetvalue.value - 1)
		)
	as foo, COUNT(a.*), offsetvalue.value
		FROM adjectives a, offsetvalue
		WHERE a.adjective SIMILAR TO '%' || $3 ||'%'
		GROUP BY offsetvalue.value;
		`,
		uai.UsernamesPerPage, uai.AdjPage, searchParam).Scan(pq.Array(&uai.Adjectives), &uai.AdjWordCount, &uai.AdjPage)

	if err == sql.ErrNoRows {
		uai.Adjectives = []string{}
		uai.AdjWordCount = 0
		uai.AdjPage = 1
		return nil
	} else {
		return err
	}
}

// setUaiNounInfo sets the nouns and relevant information for the UsernameAdminInfo struct.
func (uai *UsernameAdminInfo) setUaiNounInfo(db *sql.DB, searchParam string) error {
	err := db.QueryRow(
		`
		WITH offsetvalue AS (
			SELECT (
				CASE 
					WHEN $2 < 1 
						THEN 1
					WHEN ($3 != '') IS TRUE AND $2 > (SELECT COUNT(*) FROM nouns WHERE noun SIMILAR TO '%' || $3 ||'%') / $1
						THEN CEIL((SELECT COUNT(*) FROM nouns WHERE noun SIMILAR TO '%' || $3 || '%')::float8 / $1::float8)
					WHEN $2 > (SELECT COUNT(*) FROM nouns) / $1
						THEN CEIL((SELECT COUNT(*) FROM nouns)::float8 / $1::float8)
					ELSE $2
				END
			) as value
		)
						
		SELECT ARRAY (
			SELECT noun
				FROM nouns
				WHERE noun SIMILAR TO '%' || $3 ||'%'
				ORDER BY noun ASC
				LIMIT $1 OFFSET $1* (offsetvalue.value - 1)
			)
		as foo, COUNT(a.*), offsetvalue.value
			FROM nouns a, offsetvalue
			WHERE a.noun SIMILAR TO '%' || $3 ||'%'
			GROUP BY offsetvalue.value;
			`,
		uai.UsernamesPerPage, uai.NounPage, searchParam).Scan(pq.Array(&uai.Nouns), &uai.NounWordCount, &uai.NounPage)
	if err == sql.ErrNoRows {
		uai.Nouns = []string{}
		uai.NounWordCount = 0
		uai.NounPage = 1
		return nil
	} else {
		return err
	}
}

// GetUsernameAdminInfo returns a UsernameAdminInfo struct containing the adjectives and nouns and relevant information
// for rendering the username administration page.
func GetUsernameAdminInfo(db *sql.DB, adjPage int, nounPage int, usernamesPerPage int, search string) (*UsernameAdminInfo, error) {
	uai := UsernameAdminInfo{}

	uai.AdjPage = adjPage
	uai.NounPage = nounPage
	uai.UsernamesPerPage = usernamesPerPage

	err := uai.setUaiAdjInfo(db, search)

	if err != nil {
		return nil, err
	}

	err = uai.setUaiNounInfo(db, search)

	if err != nil {
		return nil, err
	}

	return &uai, nil
}

// AddWordToTable adds a word to the specified table.
func AddWordToTable(db *sql.DB, word string, tableId string) error {
	if tableId == NounTable {
		_, err := db.Exec(`INSERT INTO nouns VALUES ($1);`, word)
		return err
	} else if tableId == AdjectiveTable {
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
func UpdateAdjectives(db *sql.DB, oldNewAdj []OldNew) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, oldNew := range oldNewAdj {
		_, err = tx.Exec(`UPDATE adjectives SET adjective = $1 WHERE adjective = $2;`, oldNew.New, oldNew.Old)
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
func UpdateNouns(db *sql.DB, oldNewNoun []OldNew) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, oldNew := range oldNewNoun {
		_, err = tx.Exec(`UPDATE nouns SET noun = $1 WHERE noun = $2;`, oldNew.New, oldNew.Old)
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil

}
