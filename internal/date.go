package internal

import "time"

// DateYYYYMMDD returns the current date in YYYYMMDD format.
func DateYYYYMMDD(time time.Time) int {
	currentDate := time.Year()*10000 + int(time.Month())*100 + time.Day()

	return currentDate
}
