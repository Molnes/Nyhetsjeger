package data_handling

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
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

// Converts a datetime string to Norway's timezone
func DateStringToNorwayTime(dateTime string) (time.Time, error) {
	// Get the time in Norway's timezone
	norwayLocation, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		return time.Unix(0, 0), err
	}

	// Parse the time
	tempTime, err := time.ParseInLocation("2006-01-02T15:04", dateTime, norwayLocation)
	tempTime = tempTime.UTC()
	if err != nil {
		return time.Unix(0, 0), echo.NewHTTPError(http.StatusBadRequest, "Failed to parse time")
	}

	return tempTime, nil
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
