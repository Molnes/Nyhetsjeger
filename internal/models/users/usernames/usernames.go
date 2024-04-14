package usernames

import (
	"database/sql"
	"math"
)

type UsernameAdminInfo struct {
	Adjectives []string
	Nouns      []string
	AdjPage      int
	NounPage      int
	AdjWordCount int
	NounWordCount int
	UsernamesPerPage int
}

// GetUsernameAdminInfo returns a UsernameAdminInfo struct containing the adjectives and nouns and relevant information
// for rendering the username administration page.
func GetUsernameAdminInfo(db *sql.DB, adjPage int, nounPage int, usernamesPerPage int) (*UsernameAdminInfo, error) {
	uai := UsernameAdminInfo{}

	//Get the total number of words in the adjectives and nouns tables
	err := db.QueryRow(`SELECT COUNT(*) FROM adjectives;`).Scan(&uai.AdjWordCount)
	if err != nil {
		return nil, err
	}
	err = db.QueryRow(`SELECT COUNT(*) FROM nouns;`).Scan(&uai.NounWordCount)
	if err != nil {
		return nil, err
	}

	//Insert adjective page and noun page into the struct and make sure they are within bounds
	if adjPage < 1 {
		adjPage = 1
	}
	if nounPage < 1 {
		nounPage = 1
	}
	if adjPage > (uai.AdjWordCount / usernamesPerPage) {
		adjPage = int(math.Ceil((float64(uai.AdjWordCount) / float64(usernamesPerPage))))
	}
	if nounPage > (uai.NounWordCount / usernamesPerPage) {
		nounPage = int(math.Ceil((float64(uai.NounWordCount) / float64(usernamesPerPage))))
	}
	uai.AdjPage = adjPage
	uai.NounPage = nounPage
	uai.UsernamesPerPage = usernamesPerPage

	rows, err := db.Query(`
					SELECT adjective 
						FROM adjectives
						ORDER BY adjective ASC
							LIMIT $1 OFFSET $2;`,
		usernamesPerPage, usernamesPerPage*(adjPage-1))
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
		usernamesPerPage, usernamesPerPage*(nounPage-1))
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
