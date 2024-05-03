package data_handling_test

import (
	"database/sql"
	"net/url"
	"testing"
	"time"

	data_handling "github.com/Molnes/Nyhetsjeger/internal/utils/data"
)

const expectedText = "Expected %s, but got %s"

// TestConvertNullStringToURL tests the ConvertNullStringToURL function
func TestConvertNullStringToURL(t *testing.T) {
	// Test when the string is not null
	nullString := sql.NullString{
		String: "https://www.google.com",
		Valid:  true,
	}
	expected, _ := url.Parse("https://www.google.com")
	actual, _ := data_handling.ConvertNullStringToURL(&nullString)
	if actual.String() != expected.String() {
		t.Errorf(expectedText, expected, actual)
	}

	// Test when the string is null
	nullString = sql.NullString{
		String: "",
		Valid:  false,
	}
	expected, _ = url.Parse("")
	actual, _ = data_handling.ConvertNullStringToURL(&nullString)
	if actual.String() != expected.String() {
		t.Errorf(expectedText, expected, actual)
	}
}

// TestConvertNullStringToTime tests if time is converted to UTC
func TestNorwayTimeToUtc(t *testing.T) {
	// Test when the date is "2021-01-01T12:00"
	expected, _ := time.Parse("2006-01-02T15:04Z07:00", "2021-01-01T11:00+00:00")
	expected = expected.UTC()
	actual, _ := data_handling.NorwayTimeToUtc("2021-01-01T12:00")
	if actual != expected {
		t.Errorf(expectedText, expected, actual)
	}
}

// TestFormatNumberWithSpaces tests if numbers are corrected to norwegian format
func TestFormatNumberWithSpaces(t *testing.T) {
	// Test when the number is 1000000
	expected := "1 000 000"
	actual := data_handling.FormatNumberWithSpaces(1000000)
	if actual != expected {
		t.Errorf(expectedText, expected, actual)
	}

	// Test when the number is 1000
	expected = "1 000"
	actual = data_handling.FormatNumberWithSpaces(1000)
	if actual != expected {
		t.Errorf(expectedText, expected, actual)
	}

	// Test when the number is 1000000000000000000
	expected = "1 000 000 000 000 000 000"
	actual = data_handling.FormatNumberWithSpaces(1000000000000000000)
	if actual != expected {
		t.Errorf(expectedText, expected, actual)
	}
}
