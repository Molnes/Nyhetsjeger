package data_handling

import (
	"database/sql"
	"net/url"
)

// Convert a null string to a URL
func ConvertNullStringToURL(newURL *sql.NullString) (*url.URL, error) {
	// If it is not NULL
	if newURL.Valid {
		tempURL, err := url.Parse(newURL.String)
		if err == nil {
			return tempURL, nil
		} else {
			return nil, err
		}
	}

	// If it is NULL
	return &url.URL{}, nil
}
