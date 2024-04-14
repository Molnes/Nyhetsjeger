package usernames

import (
	"database/sql"
	"math"
)

type UsernameAdminInfo struct {
	Adjectives []string
	Nouns      []string
	APage      int
	NPage      int
	AWordCount int
	NWordCount int
}

// UsernamesPerPage is the number of usernames to display per page
var UsernamesPerPage = 25

// GetUsernameAdminInfo returns a UsernameAdminInfo struct containing the adjectives and nouns and relevant information
// for rendering the username administration page.
func GetUsernameAdminInfo(db *sql.DB, aPage int, nPage int) (*UsernameAdminInfo, error) {
	uai := UsernameAdminInfo{}

	//Get the total number of words in the adjectives and nouns tables
	err := db.QueryRow(`SELECT COUNT(*) FROM adjectives;`).Scan(&uai.AWordCount)
	if err != nil {
		return nil, err
	}
	err = db.QueryRow(`SELECT COUNT(*) FROM nouns;`).Scan(&uai.NWordCount)
	if err != nil {
		return nil, err
	}

	//Insert adjective page and noun page into the struct and make sure they are within bounds
	if aPage < 1 {
		aPage = 1
	}
	if nPage < 1 {
		nPage = 1
	}
	if aPage > (uai.AWordCount / UsernamesPerPage) {
		aPage = int(math.Ceil((float64(uai.AWordCount) / float64(UsernamesPerPage))))
	}
	if nPage > (uai.NWordCount / UsernamesPerPage) {
		nPage = int(math.Ceil((float64(uai.NWordCount) / float64(UsernamesPerPage))))
	}
	uai.APage = aPage
	uai.NPage = nPage

	rows, err := db.Query(`
					SELECT adjective 
						FROM adjectives
						ORDER BY adjective ASC
							LIMIT $1 OFFSET $2;`,
		UsernamesPerPage, UsernamesPerPage*(aPage-1))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//Loop through the adjectives
	for rows.Next() {
		var adjective string
		err = rows.Scan(&adjective)
		if err != nil {
			return nil, err
		}
		uai.Adjectives = append(uai.Adjectives, adjective)
	}

	rows, err = db.Query(`
	SELECT noun
		FROM nouns
		ORDER BY noun ASC
			LIMIT $1 OFFSET $2;`,
		UsernamesPerPage, UsernamesPerPage*(nPage-1))
	if err != nil {
		return nil, err
	}

	//Loop through the nouns
	for rows.Next() {
		var noun string
		err = rows.Scan(&noun)
		if err != nil {
			return nil, err
		}
		uai.Nouns = append(uai.Nouns, noun)
	}

	return &uai, nil
}
