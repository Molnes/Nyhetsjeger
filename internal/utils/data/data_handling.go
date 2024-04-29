package data_handling

import (
	"database/sql"
	"log"
	"net/url"
	"strconv"
	"time"
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

// Takes in a dateTime string, such as "1970-01-01T12:00" and assumes it to be in Norway's timezone.
// It then converts it to UTC time.
func NorwayTimeToUtc(dateTime string) (time.Time, error) {
	// Load Norway's timezone
	norwayTZ, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		log.Println("Failed to load Norway's timezone")
		return time.Unix(0, 0), err
	}

	// Parse the input string in Norway's timezone
	norwayTime, err := time.ParseInLocation("2006-01-02T15:04", dateTime, norwayTZ)
	if err != nil {
		log.Println("Failed to parse time", dateTime)
		return time.Unix(0, 0), err
	}

	// Convert the parsed time to UTC
	utcTime := norwayTime.UTC()
	return utcTime, nil
}

// Converts any time to Norway's timezone
func GetNorwayTime(norwayTime time.Time) time.Time {
	norwayLocation, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		log.Println("Failed to load Norway's timezone")
	}

	return norwayTime.In(norwayLocation)
}

// Format a number with spaces.
// Example: "1000000" becomes "1 000 000".
func FormatNumberWithSpaces(num int) string {
	// This function was written by ChatGPT
	numStr := strconv.Itoa(num)
	var formattedNum string

	for i := len(numStr) - 1; i >= 0; i-- {
		formattedNum = string(numStr[i]) + formattedNum
		if (len(numStr)-i)%3 == 0 && i != 0 {
			formattedNum = " " + formattedNum
		}
	}

	return formattedNum
}
