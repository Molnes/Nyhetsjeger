package date_utils

import (
	"strconv"
	"time"
)

// DateToNorwegianString converts a date to a string in the format "Mandag 2. januar 2006"
func DateToNorwegianString(date time.Time) string {
	timeString := timeString(date)
	day := extractDayFromDate(date)
	month := extractMonthFromDate(date)
	dayNumber := extractDayNumberFromDate(date)

	return englishDayNameToNorwegian(day) + ", " + strconv.Itoa(dayNumber) + ". " + englishMonthNameToNorwegian(month) + " " + timeString
}

// TimeToTimeString converts a time to a string in the format "15:04"
func timeString(time time.Time) string {
	return time.Format("15:04")
}

// extractDayFromDate extracts the day from a date in string format
func extractDayFromDate(date time.Time) string {
	return date.Format("Monday")
}

// extractMonthFromDate extracts the month from a date in string format
func extractMonthFromDate(date time.Time) string {
	return date.Format("January")
}

// extractDayNumberFromDate extracts the day number from a date in int format
func extractDayNumberFromDate(date time.Time) int {
	return date.Day()
}

// extractYearFromDate extracts the year from a date in int format
func extractYearFromDate(date time.Time) int {
	return date.Year()
}

// Takes a day name in English and returns the Norwegian equivalent
func englishDayNameToNorwegian(day string) string {
	switch day {
	case "Monday":
		return "Mandag"
	case "Tuesday":
		return "Tirsdag"
	case "Wednesday":
		return "Onsdag"
	case "Thursday":
		return "Torsdag"
	case "Friday":
		return "Fredag"
	case "Saturday":
		return "Lørdag"
	case "Sunday":
		return "Søndag"
	default:
		return ""
	}
}

// Takes a month name in English and returns the Norwegian equivalent
func englishMonthNameToNorwegian(month string) string {
	switch month {
	case "January":
		return "januar"
	case "February":
		return "februar"
	case "March":
		return "mars"
	case "April":
		return "april"
	case "May":
		return "mai"
	case "June":
		return "juni"
	case "July":
		return "juli"
	case "August":
		return "august"
	case "September":
		return "september"
	case "October":
		return "oktober"
	case "November":
		return "november"
	case "December":
		return "desember"
	default:
		return ""
	}
}
