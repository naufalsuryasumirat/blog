package utils

import (
	"time"
)

func GetTimestamp() (time.Time, string) {
	const format string = "2006_01_02-15_04"
	timestamp := time.Now().UTC()
	timestampFormatted := timestamp.Format(format)

	return timestamp, timestampFormatted
}
