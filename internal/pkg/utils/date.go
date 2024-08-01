package utils

import "time"

func HasUpdated(lastUpdate time.Time) bool {
	now := time.Now()
	year, month, day := now.Date()
	todayMidnight := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
	return lastUpdate.After(todayMidnight)
}
