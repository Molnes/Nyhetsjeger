package date_utils_test

import (
	"github.com/Molnes/Nyhetsjeger/internal/utils/date"
	"testing"
	"time"
)

// TestDateToNorwegianString tests if the date is converted to Norwegian time
func TestDateToNorwegianString(t *testing.T) {
	date := time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)
    expected := "Fredag, 1. januar 00:00"
	actual := date_utils.DateToNorwegianString(date)
	if actual != expected {
		t.Errorf("Expected %s, but got %s", expected, actual)
	}
}

// TestDateToNorwegianString2 tests if the date is converted to Norwegian time
func TestDateToNorwegianString2(t *testing.T) {
        date := time.Date(2021, time.January, 2, 0, 0, 0, 0, time.UTC)
        expected := "Lørdag, 2. januar 00:00"
        actual := date_utils.DateToNorwegianString(date)
        if actual != expected {
                t.Errorf("Expected %s, but got %s", expected, actual)
        }
}
