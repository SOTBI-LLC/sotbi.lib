package times

import (
	"fmt"
	"time"
)

const (
	MSK                   = " MSK"
	MSKWithMonthStartDate = "-01 MSK"
)

func GetMoscowLocation() *time.Location {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		location = time.FixedZone("UTC+3", 3*60*60)
	}

	return location
}

func DurationFmt(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute

	return fmt.Sprintf("%02d:%02d", h, m)
}
