package data_handling

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"
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
func DateStringToNorwayTime(dateTime string, c echo.Context) (time.Time, error) {
	// Get the time in Norway's timezone
	norwayLocation, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		return time.Unix(0, 0), echo.NewHTTPError(http.StatusInternalServerError, "Failed to get Norway's time zone")
	}

	// Parse the time
	activeStartTime, err := time.ParseInLocation("2006-01-02T15:04", dateTime, norwayLocation)
	activeStartTime = activeStartTime.UTC()
	if err != nil {
		return time.Unix(0, 0), echo.NewHTTPError(http.StatusBadRequest, "Failed to parse active start time")
	}

	return activeStartTime, nil
}

// Converts any time to Norway's timezone
func GetNorwayTime(norwayTime time.Time) time.Time {
	norwayLocation, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		log.Println("Failed to load Norway's timezone")
	}

	return norwayTime.In(norwayLocation)
}
