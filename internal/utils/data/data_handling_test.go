package data_handling_test

import (
	"database/sql"
	"github.com/Molnes/Nyhetsjeger/internal/utils/data"
	"net/url"
	"testing"
)

func TestConvertNullStringToURL(t *testing.T) {
	// Test when the string is not null
	nullString := sql.NullString{
		String: "https://www.google.com",
		Valid:  true,
	}
	expected, _ := url.Parse("https://www.google.com")
	actual, _ := data_handling.ConvertNullStringToURL(&nullString)
	if actual.String() != expected.String() {
		t.Errorf("Expected %s, but got %s", expected, actual)
	}

	// Test when the string is null
	nullString = sql.NullString{
		String: "",
		Valid:  false,
	}
	expected, _ = url.Parse("")
	actual, _ = data_handling.ConvertNullStringToURL(&nullString)
	if actual.String() != expected.String() {
		t.Errorf("Expected %s, but got %s", expected, actual)
	}
}

func TestFormatNumberWithSpaces(t *testing.T) {
        // Test when the number is 1000000
        expected := "1 000 000"
        actual := data_handling.FormatNumberWithSpaces(1000000)
        if actual != expected {
                t.Errorf("Expected %s, but got %s", expected, actual)
        }

        // Test when the number is 1000
        expected = "1 000"
        actual = data_handling.FormatNumberWithSpaces(1000)
        if actual != expected {
                t.Errorf("Expected %s, but got %s", expected, actual)
        }

        // Test when the number is 1000000000000000000
        expected = "1 000 000 000 000 000 000"
        actual = data_handling.FormatNumberWithSpaces(1000000000000000000)
        if actual != expected {
                t.Errorf("Expected %s, but got %s", expected, actual)
        }
}
